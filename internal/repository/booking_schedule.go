package repository

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"apirusdotistamobile/internal/model"
)

type scheduleDayColumn struct {
	Index  int
	Name   string
	Column string
}

type poliScheduleConfig struct {
	ID       int64
	Name     string
	Buka     string
	Tutup    string
	Praktik  string
	DayQuota map[int]int64
}

type jadwalDokterRow struct {
	Poli        string
	Doctor      string
	Days        string
	Start       string
	End         string
	MatchedName string
}

var bookingScheduleDayColumns = []scheduleDayColumn{
	{Index: 0, Name: "Minggu", Column: "sunday"},
	{Index: 1, Name: "Senin", Column: "monday"},
	{Index: 2, Name: "Selasa", Column: "tuesday"},
	{Index: 3, Name: "Rabu", Column: "wednesday"},
	{Index: 4, Name: "Kamis", Column: "thursday"},
	{Index: 5, Name: "Jumat", Column: "friday"},
	{Index: 6, Name: "Sabtu", Column: "saturday"},
}

func (r *BookingRepository) loadUnifiedBookingSchedules(ctx context.Context, poli model.BookingPoliOption, doctors []model.BookingDoctorOption) ([]model.BookingScheduleOption, map[int64][]model.BookingScheduleOption, error) {
	config, err := r.loadPoliScheduleConfig(ctx, poli.ID, poli.Nama, poli.Buka, poli.Tutup, poli.Praktik)
	if err != nil {
		return nil, nil, err
	}

	schedules := make([]model.BookingScheduleOption, 0)
	schedules = append(schedules, poliSchedulesFromConfig(config)...)

	doctorSchedules, byDoctor, err := r.loadJadwalDokterSchedules(ctx, config, doctors)
	if err != nil {
		return nil, nil, err
	}
	schedules = append(schedules, doctorSchedules...)
	sortBookingSchedules(schedules)
	for doctorID := range byDoctor {
		sortBookingSchedules(byDoctor[doctorID])
	}

	return schedules, byDoctor, nil
}

func (r *HolidayRepository) scheduleStatusForDate(ctx context.Context, dateKey string, poliID int64) ([]model.BookingScheduleOption, bool, bool, bool, error) {
	if poliID <= 0 {
		return nil, false, false, false, nil
	}

	config, err := loadPoliScheduleConfig(ctx, r.DB, poliID, "", "", "", "")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, false, false, nil
		}
		return nil, false, false, false, err
	}

	dayIndex := int(dateFromKeyWeekday(dateKey))
	praktik := schedulePracticeEnabled(config.Praktik)
	hasConfiguredSchedule := hasAnyPoliDayQuota(config)
	schedules := make([]model.BookingScheduleOption, 0)
	hasPoliSchedule := false
	hasDoctorSchedule := false

	if praktik && dayIndex != 0 {
		if quota, ok := config.DayQuota[dayIndex]; ok && quota > 0 {
			hasPoliSchedule = true
			schedules = append(schedules, model.BookingScheduleOption{
				Source:   "polis",
				DayIndex: dayIndex,
				Hari:     bookingDayName(dayIndex),
				Buka:     config.Buka,
				Tutup:    config.Tutup,
				Praktik:  true,
				Kuota:    quota,
				PoliID:   config.ID,
				PoliName: config.Name,
			})
		}
	}

	doctorRows, err := loadJadwalDokterRows(ctx, r.DB, config.Name)
	if err != nil {
		return nil, false, false, false, err
	}
	hasConfiguredSchedule = hasConfiguredSchedule || len(doctorRows) > 0
	for _, row := range doctorRows {
		if !parseScheduleDays(row.Days)[dayIndex] {
			continue
		}
		start, end := splitScheduleHours(row.Start, row.End)
		if start == "" && end == "" {
			start, end = config.Buka, config.Tutup
		}
		hasDoctorSchedule = praktik
		schedules = append(schedules, model.BookingScheduleOption{
			Source:     "jadwaldokters",
			DayIndex:   dayIndex,
			Hari:       bookingDayName(dayIndex),
			Buka:       start,
			Tutup:      end,
			Praktik:    praktik,
			PoliID:     config.ID,
			PoliName:   config.Name,
			DoctorName: row.Doctor,
		})
	}

	sortBookingSchedules(schedules)
	return schedules, hasConfiguredSchedule, hasPoliSchedule, hasDoctorSchedule, nil
}

func (r *BookingRepository) loadPoliScheduleConfig(ctx context.Context, poliID int64, poliName string, buka string, tutup string, praktik string) (poliScheduleConfig, error) {
	return loadPoliScheduleConfig(ctx, r.DB, poliID, poliName, buka, tutup, praktik)
}

func loadPoliScheduleConfig(ctx context.Context, db *sql.DB, poliID int64, poliName string, buka string, tutup string, praktik string) (poliScheduleConfig, error) {
	config := poliScheduleConfig{
		ID:       poliID,
		Name:     poliName,
		Buka:     buka,
		Tutup:    tutup,
		Praktik:  praktik,
		DayQuota: map[int]int64{},
	}

	var sunday, monday, tuesday, wednesday, thursday, friday, saturday sql.NullInt64
	if err := db.QueryRowContext(ctx, `
		SELECT
			COALESCE(nama, ''),
			COALESCE(DATE_FORMAT(buka, '%H:%i'), ''),
			COALESCE(DATE_FORMAT(tutup, '%H:%i'), ''),
			COALESCE(praktik, ''),
			sunday,
			monday,
			tuesday,
			wednesday,
			thursday,
			friday,
			saturday
		FROM polis
		WHERE id = ?
		LIMIT 1
	`, poliID).Scan(
		&config.Name,
		&config.Buka,
		&config.Tutup,
		&config.Praktik,
		&sunday,
		&monday,
		&tuesday,
		&wednesday,
		&thursday,
		&friday,
		&saturday,
	); err != nil {
		return config, err
	}

	for _, item := range []struct {
		index int
		value sql.NullInt64
	}{
		{0, sunday},
		{1, monday},
		{2, tuesday},
		{3, wednesday},
		{4, thursday},
		{5, friday},
		{6, saturday},
	} {
		if item.value.Valid {
			config.DayQuota[item.index] = item.value.Int64
		}
	}

	return config, nil
}

func poliSchedulesFromConfig(config poliScheduleConfig) []model.BookingScheduleOption {
	if !schedulePracticeEnabled(config.Praktik) {
		return nil
	}

	schedules := make([]model.BookingScheduleOption, 0, 6)
	for _, day := range bookingScheduleDayColumns {
		if day.Index == 0 {
			continue
		}
		quota, ok := config.DayQuota[day.Index]
		if !ok || quota <= 0 {
			continue
		}
		schedules = append(schedules, model.BookingScheduleOption{
			Source:   "polis",
			DayIndex: day.Index,
			Hari:     day.Name,
			Buka:     config.Buka,
			Tutup:    config.Tutup,
			Praktik:  true,
			Kuota:    quota,
			PoliID:   config.ID,
			PoliName: config.Name,
		})
	}
	return schedules
}

func (r *BookingRepository) loadJadwalDokterSchedules(ctx context.Context, config poliScheduleConfig, doctors []model.BookingDoctorOption) ([]model.BookingScheduleOption, map[int64][]model.BookingScheduleOption, error) {
	rows, err := loadJadwalDokterRows(ctx, r.DB, config.Name)
	if err != nil {
		return nil, nil, err
	}

	doctorNames := make(map[int64]string, len(doctors))
	for _, doctor := range doctors {
		doctorNames[doctor.ID] = doctor.Nama
	}

	schedules := make([]model.BookingScheduleOption, 0)
	byDoctor := make(map[int64][]model.BookingScheduleOption)
	praktik := schedulePracticeEnabled(config.Praktik)
	for _, row := range rows {
		start, end := splitScheduleHours(row.Start, row.End)
		if start == "" && end == "" {
			start, end = config.Buka, config.Tutup
		}
		doctorID := matchDoctorScheduleID(row.Doctor, doctorNames)
		for dayIndex := range parseScheduleDays(row.Days) {
			schedule := model.BookingScheduleOption{
				Source:     "jadwaldokters",
				DayIndex:   dayIndex,
				Hari:       bookingDayName(dayIndex),
				Buka:       start,
				Tutup:      end,
				Praktik:    praktik,
				PoliID:     config.ID,
				PoliName:   config.Name,
				DoctorID:   doctorID,
				DoctorName: row.Doctor,
			}
			if doctorID > 0 {
				if dbName := doctorNames[doctorID]; dbName != "" {
					schedule.DoctorName = dbName
				}
				byDoctor[doctorID] = append(byDoctor[doctorID], schedule)
			}
			schedules = append(schedules, schedule)
		}
	}
	return schedules, byDoctor, nil
}

func loadJadwalDokterRows(ctx context.Context, db *sql.DB, poliName string) ([]jadwalDokterRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			COALESCE(poli, ''),
			COALESCE(dokter, ''),
			COALESCE(hari, ''),
			COALESCE(jam_mulai, ''),
			COALESCE(jam_berakhir, '')
		FROM jadwaldokters
		WHERE LOWER(TRIM(poli)) = LOWER(TRIM(?))
		OR LOWER(TRIM(poli)) LIKE CONCAT(LOWER(TRIM(?)), ' (%')
		OR LOWER(TRIM(?)) LIKE CONCAT(LOWER(TRIM(poli)), '%')
		ORDER BY poli ASC, dokter ASC, hari ASC
		LIMIT 100
	`, poliName, poliName, poliName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]jadwalDokterRow, 0)
	for rows.Next() {
		var row jadwalDokterRow
		if err = rows.Scan(&row.Poli, &row.Doctor, &row.Days, &row.Start, &row.End); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func parseScheduleDays(value string) map[int]bool {
	days := map[int]bool{}
	for _, part := range strings.Split(value, ",") {
		dayIndex, ok := parseScheduleDay(part)
		if ok {
			days[dayIndex] = true
		}
	}
	return days
}

func parseScheduleDay(value string) (int, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.Trim(normalized, ".")
	switch normalized {
	case "minggu", "ahad", "sunday":
		return 0, true
	case "senin", "monday":
		return 1, true
	case "selasa", "tuesday":
		return 2, true
	case "rabu", "wednesday":
		return 3, true
	case "kamis", "thursday":
		return 4, true
	case "jumat", "jum'at", "friday":
		return 5, true
	case "sabtu", "saturday":
		return 6, true
	default:
		return 0, false
	}
}

func splitScheduleHours(start string, end string) (string, string) {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if end != "" {
		return start, end
	}

	normalized := strings.NewReplacer("\u2013", "-", "\u2014", "-").Replace(start)
	parts := strings.SplitN(normalized, "-", 2)
	if len(parts) != 2 {
		return start, end
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func schedulePracticeEnabled(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "Y")
}

func hasAnyPoliDayQuota(config poliScheduleConfig) bool {
	for _, quota := range config.DayQuota {
		if quota > 0 {
			return true
		}
	}
	return false
}

func bookingDayName(dayIndex int) string {
	for _, day := range bookingScheduleDayColumns {
		if day.Index == dayIndex {
			return day.Name
		}
	}
	return ""
}

func dateFromKeyWeekday(dateKey string) int {
	parts := strings.Split(dateKey, "-")
	if len(parts) != 3 {
		return 0
	}
	year, errYear := strconv.Atoi(parts[0])
	month, errMonth := strconv.Atoi(parts[1])
	day, errDay := strconv.Atoi(parts[2])
	if errYear != nil || errMonth != nil || errDay != nil {
		return 0
	}
	return int(dateFromParts(year, month, day).Weekday())
}

func dateFromParts(year int, month int, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

func matchDoctorScheduleID(scheduleDoctor string, doctorNames map[int64]string) int64 {
	needle := normalizeScheduleName(scheduleDoctor)
	if len(needle) < 5 {
		return 0
	}
	for id, name := range doctorNames {
		candidate := normalizeScheduleName(name)
		if candidate == "" {
			continue
		}
		if candidate == needle || strings.Contains(candidate, needle) || strings.Contains(needle, candidate) {
			return id
		}
	}
	return 0
}

func normalizeScheduleName(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, value)
}

func sortBookingSchedules(schedules []model.BookingScheduleOption) {
	sort.SliceStable(schedules, func(i, j int) bool {
		if schedules[i].DayIndex != schedules[j].DayIndex {
			return schedules[i].DayIndex < schedules[j].DayIndex
		}
		if schedules[i].Source != schedules[j].Source {
			return schedules[i].Source < schedules[j].Source
		}
		if schedules[i].DoctorName != schedules[j].DoctorName {
			return schedules[i].DoctorName < schedules[j].DoctorName
		}
		return schedules[i].Buka < schedules[j].Buka
	})
}
