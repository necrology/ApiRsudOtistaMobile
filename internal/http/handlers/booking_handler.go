package handlers

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	repository *repository.BookingRepository
	holidays   *repository.HolidayRepository
}

func NewBookingHandler(repository *repository.BookingRepository, holidays *repository.HolidayRepository) *BookingHandler {
	return &BookingHandler{repository: repository, holidays: holidays}
}

type bookingQueueRequest struct {
	PoliID      int64  `json:"poli_id"`
	Tanggal     string `json:"tanggal"`
	Bayar       string `json:"bayar"`
	JenisPasien string `json:"jenis_pasien"`
	DoctorID    string `json:"dokter_id"`
	QueueGroup  string `json:"queue_group"`
	IsJkn       bool   `json:"is_jkn"`
}

func (h *BookingHandler) CreateGeneralBooking(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}
	if !principal.HasLinkedPatient() {
		return response.Error(c, fiber.StatusForbidden, "akun belum terhubung dengan rekam medis")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	var req bookingQueueRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	tanggal, err := parseBookingDate(req.Tanggal)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.repository.CreateGeneralBooking(ctx, repository.BookingTarget{
		PatientID:   principal.PatientID,
		NoRM:        principal.NoRM,
		Email:       principal.Email,
		PoliID:      req.PoliID,
		Tanggal:     tanggal,
		Bayar:       strings.TrimSpace(req.Bayar),
		JenisPasien: strings.TrimSpace(req.JenisPasien),
		DoctorID:    strings.TrimSpace(req.DoctorID),
		QueueGroup:  strings.TrimSpace(req.QueueGroup),
		IsJkn:       req.IsJkn,
	})
	if err != nil {
		var validationError *repository.BookingValidationError
		if errors.As(err, &validationError) {
			return response.Error(c, fiber.StatusBadRequest, validationError.Message)
		}
		log.Printf("CreateGeneralBooking failed error_type=%T", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal membuat booking")
	}

	return response.OK(c, fiber.Map{
		"message": "booking berhasil dibuat",
		"data":    result,
	})
}

func (h *BookingHandler) ListMyGeneralBookings(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}
	if !principal.HasLinkedPatient() {
		return response.Error(c, fiber.StatusForbidden, "akun belum terhubung dengan rekam medis")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 12*time.Second)
	defer cancel()

	filter, err := bookingListFilterFromQuery(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	result, err := h.repository.ListPatientGeneralBookings(ctx, principal.PatientID, filter)
	if err != nil {
		log.Printf("ListMyGeneralBookings error: %v", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil data booking")
	}

	return response.OK(c, result)
}

func (h *BookingHandler) BookingOptions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 8*time.Second)
	defer cancel()

	poliID, err := strconv.ParseInt(strings.TrimSpace(c.Params("poli_id")), 10, 64)
	if err != nil || poliID <= 0 {
		poliID, err = strconv.ParseInt(strings.TrimSpace(c.Query("poli_id")), 10, 64)
	}
	if err != nil || poliID <= 0 {
		return response.Error(c, fiber.StatusBadRequest, "poli tidak valid")
	}

	result, err := h.repository.BookingOptions(ctx, poliID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil opsi booking")
	}

	return response.OK(c, result)
}

func (h *BookingHandler) BookingCalendar(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 12*time.Second)
	defer cancel()

	now := time.Now()
	year := c.QueryInt("year", now.Year())
	month := c.QueryInt("month", int(now.Month()))
	poliID := int64(c.QueryInt("poli_id", 0))

	result, err := h.holidays.BookingCalendar(ctx, year, month, poliID)
	if err != nil {
		var validationError *repository.BookingValidationError
		if errors.As(err, &validationError) {
			return response.Error(c, fiber.StatusBadRequest, validationError.Message)
		}
		log.Printf("BookingCalendar failed error_type=%T", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil kalender booking")
	}

	return response.OK(c, result)
}

func parseBookingDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Now(), nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err == nil {
		return parsed, nil
	}

	if unix, convErr := strconv.ParseInt(trimmed, 10, 64); convErr == nil {
		return time.Unix(unix, 0), nil
	}

	return time.Time{}, errors.New("tanggal booking harus format YYYY-MM-DD")
}

func bookingListFilterFromQuery(c *fiber.Ctx) (repository.BookingListFilter, error) {
	filter := repository.BookingListFilter{
		Status: strings.TrimSpace(c.Query("status")),
		Limit:  c.QueryInt("limit", 50),
	}

	dateValue := strings.TrimSpace(c.Query("tanggal"))
	if dateValue == "" {
		dateValue = strings.TrimSpace(c.Query("date"))
	}
	if dateValue == "" && strings.TrimSpace(c.Query("all_dates")) == "" {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		filter.Date = &today
		return filter, nil
	}
	if dateValue == "" {
		return filter, nil
	}

	parsedDate, err := parseBookingDate(dateValue)
	if err != nil {
		return filter, err
	}
	normalizedDate := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
	filter.Date = &normalizedDate

	return filter, nil
}
