package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/model"
)

type HolidayRepository struct {
	DB      *sql.DB
	BaseURL string
	Client  *http.Client
}

type holidayRecord struct {
	Date string
	Name string
	Type string
}

type holidayAPIResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		Date string `json:"date"`
		Day  string `json:"day"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"data"`
}

const holidaySourceTanggalMerah = "tanggalmerah.upset.dev"

func NewHolidayRepository(db *sql.DB, cfg config.HolidayConfig) *HolidayRepository {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://tanggalmerah.upset.dev"
	}
	return &HolidayRepository{
		DB:      db,
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (r *HolidayRepository) BookingCalendar(ctx context.Context, year int, month int, poliID int64) (*model.BookingCalendarResponse, error) {
	if year < 1900 || year > 2200 {
		return nil, bookingValidationError("tahun kalender tidak valid")
	}
	if month < 1 || month > 12 {
		return nil, bookingValidationError("bulan kalender tidak valid")
	}

	_ = r.SyncYear(ctx, year)

	firstDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDate := firstDate.AddDate(0, 1, -1)
	holidays, err := r.holidaysBetween(ctx, firstDate, lastDate)
	if err != nil {
		return nil, err
	}

	days := make([]model.BookingCalendarDay, 0, lastDate.Day())
	for date := firstDate; !date.After(lastDate); date = date.AddDate(0, 0, 1) {
		day, err := r.dayStatus(ctx, date, poliID, holidays[date.Format("2006-01-02")])
		if err != nil {
			return nil, err
		}
		days = append(days, day)
	}

	return &model.BookingCalendarResponse{
		Year:  year,
		Month: month,
		Days:  days,
	}, nil
}

func (r *HolidayRepository) ValidateBookingDate(ctx context.Context, date time.Time, poliID int64) error {
	if date.IsZero() {
		return bookingValidationError("tanggal booking tidak valid")
	}
	_ = r.SyncYear(ctx, date.Year())

	holidays, err := r.holidaysBetween(ctx, date, date)
	if err != nil {
		return err
	}
	status, err := r.dayStatus(ctx, date, poliID, holidays[date.Format("2006-01-02")])
	if err != nil {
		return err
	}
	if status.IsOpen {
		return nil
	}
	if status.Reason != "" {
		return bookingValidationError(status.Reason)
	}
	return bookingValidationError("tanggal tersebut tidak tersedia untuk pendaftaran")
}

func (r *HolidayRepository) SyncYear(ctx context.Context, year int) error {
	shouldSync, err := r.shouldSyncYear(ctx, year)
	if err != nil {
		return err
	}
	if !shouldSync {
		return nil
	}

	records, err := r.fetchYear(ctx, year)
	if err != nil {
		return err
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		UPDATE tanggal_libur_rs
		SET aktif = 0, updated_at = NOW()
		WHERE YEAR(tanggal) = ?
		AND sumber = ?
	`, year, holidaySourceTanggalMerah); err != nil {
		return err
	}

	for _, record := range records {
		raw, _ := json.Marshal(record)
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tanggal_libur_rs (
				tanggal,
				nama_libur,
				jenis,
				sumber,
				aktif,
				raw_json,
				synced_at,
				created_at,
				updated_at
			) VALUES (?, ?, ?, ?, 1, ?, NOW(), NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				nama_libur = VALUES(nama_libur),
				jenis = VALUES(jenis),
				sumber = VALUES(sumber),
				aktif = 1,
				raw_json = VALUES(raw_json),
				synced_at = NOW(),
				updated_at = NOW()
		`, record.Date, record.Name, record.Type, holidaySourceTanggalMerah, string(raw)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *HolidayRepository) shouldSyncYear(ctx context.Context, year int) (bool, error) {
	var count int
	var lastSync sql.NullTime
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(synced_at)
		FROM tanggal_libur_rs
		WHERE YEAR(tanggal) = ?
		AND sumber = ?
		AND aktif = 1
	`, year, holidaySourceTanggalMerah).Scan(&count, &lastSync); err != nil {
		return false, err
	}

	if count == 0 || !lastSync.Valid {
		return true, nil
	}

	return time.Since(lastSync.Time) >= 24*time.Hour, nil
}

func (r *HolidayRepository) hasYearCache(ctx context.Context, year int) (bool, error) {
	var count int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM tanggal_libur_rs
		WHERE YEAR(tanggal) = ?
		AND sumber = ?
		AND aktif = 1
	`, year, holidaySourceTanggalMerah).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *HolidayRepository) fetchYear(ctx context.Context, year int) ([]holidayRecord, error) {
	endpoint, err := url.Parse(r.BaseURL + "/api/holidays")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("year", strconv.Itoa(year))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tanggal merah api status %d", resp.StatusCode)
	}

	var decoded holidayAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if !decoded.Success {
		return nil, errors.New("tanggal merah api gagal mengambil data")
	}

	records := make([]holidayRecord, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		holidayType := strings.ToLower(strings.TrimSpace(item.Type))
		if holidayType != "leave" {
			holidayType = "holiday"
		}
		if item.Date == "" || item.Name == "" {
			continue
		}
		records = append(records, holidayRecord{
			Date: item.Date,
			Name: item.Name,
			Type: holidayType,
		})
	}
	return records, nil
}

func (r *HolidayRepository) holidaysBetween(ctx context.Context, startDate time.Time, endDate time.Time) (map[string][]holidayRecord, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			DATE_FORMAT(tanggal, '%Y-%m-%d'),
			COALESCE(nama_libur, ''),
			COALESCE(jenis, 'holiday')
		FROM tanggal_libur_rs
		WHERE aktif = 1
		AND tanggal BETWEEN ? AND ?
		ORDER BY tanggal ASC, jenis ASC, nama_libur ASC
	`, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := map[string][]holidayRecord{}
	for rows.Next() {
		var record holidayRecord
		if err = rows.Scan(&record.Date, &record.Name, &record.Type); err != nil {
			return nil, err
		}
		results[record.Date] = append(results[record.Date], record)
	}
	return results, rows.Err()
}

func (r *HolidayRepository) dayStatus(ctx context.Context, date time.Time, poliID int64, holidays []holidayRecord) (model.BookingCalendarDay, error) {
	dayKey := date.Format("2006-01-02")
	holidayNames := make([]string, 0, len(holidays))
	isHoliday := false
	isLeave := false
	for _, holiday := range holidays {
		if holiday.Name != "" {
			holidayNames = append(holidayNames, holiday.Name)
		}
		switch strings.ToLower(holiday.Type) {
		case "leave":
			isLeave = true
		default:
			isHoliday = true
		}
	}

	isSunday := date.Weekday() == time.Sunday
	schedules, hasConfiguredSchedule, hasPoliSchedule, hasDoctorSchedule, err := r.scheduleStatusForDate(ctx, dayKey, poliID)
	if err != nil {
		return model.BookingCalendarDay{}, err
	}
	hasSundaySchedule := isSunday && hasDoctorSchedule

	reason := ""
	isOffDay := false
	if isHoliday || isLeave {
		isOffDay = true
		if len(holidayNames) > 0 {
			reason = strings.Join(holidayNames, ", ")
		} else if isLeave {
			reason = "Cuti bersama"
		} else {
			reason = "Libur nasional"
		}
	}
	if poliID > 0 && hasConfiguredSchedule && len(schedules) == 0 {
		isOffDay = true
		if reason == "" {
			reason = "Tidak ada jadwal poliklinik/dokter"
		}
	}
	if poliID > 0 && hasConfiguredSchedule && len(schedules) > 0 {
		hasActiveSchedule := false
		for _, schedule := range schedules {
			if schedule.Praktik {
				hasActiveSchedule = true
				break
			}
		}
		if !hasActiveSchedule {
			isOffDay = true
			if reason == "" {
				reason = "Poliklinik tidak aktif"
			}
		}
	}
	if isSunday && !hasSundaySchedule {
		isOffDay = true
		if reason == "" {
			reason = "Hari Minggu"
		}
	}

	return model.BookingCalendarDay{
		Date:              dayKey,
		DayName:           indonesianDayName(date.Weekday()),
		IsSunday:          isSunday,
		HasSundaySchedule: hasSundaySchedule,
		HasPoliSchedule:   hasPoliSchedule,
		HasDoctorSchedule: hasDoctorSchedule,
		IsHoliday:         isHoliday,
		IsLeave:           isLeave,
		IsOffDay:          isOffDay,
		IsOpen:            !isOffDay,
		Reason:            reason,
		HolidayNames:      holidayNames,
		Schedules:         schedules,
	}, nil
}

func indonesianDayName(weekday time.Weekday) string {
	switch weekday {
	case time.Sunday:
		return "Minggu"
	case time.Monday:
		return "Senin"
	case time.Tuesday:
		return "Selasa"
	case time.Wednesday:
		return "Rabu"
	case time.Thursday:
		return "Kamis"
	case time.Friday:
		return "Jumat"
	case time.Saturday:
		return "Sabtu"
	default:
		return ""
	}
}
