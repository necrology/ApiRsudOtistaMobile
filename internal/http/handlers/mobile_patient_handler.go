package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type MobilePatientHandler struct {
	repository *repository.MobilePatientRepository
}

var (
	errMobilePatientUnauthenticated = errors.New("autentikasi diperlukan")
	errMobilePatientUnlinked        = errors.New("akun belum terhubung dengan rekam medis")
)

func NewMobilePatientHandler(repository *repository.MobilePatientRepository) *MobilePatientHandler {
	return &MobilePatientHandler{repository: repository}
}

func (h *MobilePatientHandler) Profile(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 8*time.Second)
	defer cancel()

	profile, err := h.repository.Profile(ctx, userID, patientID)
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, profile)
}

func (h *MobilePatientHandler) Visits(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	visits, err := h.repository.Visits(ctx, userID, patientID, c.QueryInt("limit", 20))
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, visits)
}

func (h *MobilePatientHandler) MedicalSummaries(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	summaries, err := h.repository.MedicalSummaries(ctx, userID, patientID, c.QueryInt("limit", 20))
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, summaries)
}

func (h *MobilePatientHandler) MedicalSummaryPDF(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	registrationID, err := strconv.ParseInt(strings.TrimSpace(c.Params("registration_id")), 10, 64)
	if err != nil || registrationID <= 0 {
		return response.Error(c, fiber.StatusBadRequest, "registrasi tidak valid")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()

	doc, err := h.repository.MedicalResumeDocument(ctx, userID, patientID, registrationID)
	if err != nil {
		return h.handlePatientError(c, err)
	}

	pdfBytes, err := renderMedicalResumePDF(doc)
	if err != nil {
		log.Printf("medical resume pdf failed: %v", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal membuat PDF resume medis")
	}

	disposition := "inline"
	if queryBoolValue(c.Query("download")) {
		disposition = "attachment"
	}
	fileName := fmt.Sprintf("resume-medis-%s-%d.pdf", safeFileName(doc.NoRM), doc.RegistrationID)

	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`%s; filename="%s"`, disposition, fileName))
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Send(pdfBytes)
}

func (h *MobilePatientHandler) LaboratoryResults(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 12*time.Second)
	defer cancel()

	results, err := h.repository.LaboratoryResults(ctx, userID, patientID, c.QueryInt("limit", 20))
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, results)
}

func (h *MobilePatientHandler) RadiologyResults(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 12*time.Second)
	defer cancel()

	results, err := h.repository.RadiologyResults(ctx, userID, patientID, c.QueryInt("limit", 20))
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, results)
}

func (h *MobilePatientHandler) Prescriptions(c *fiber.Ctx) error {
	userID, patientID, err := mobilePatientIdentity(c)
	if err != nil {
		return mobilePatientIdentityError(c, err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 12*time.Second)
	defer cancel()

	results, err := h.repository.Prescriptions(ctx, userID, patientID, c.QueryInt("limit", 20))
	if err != nil {
		return h.handlePatientError(c, err)
	}

	return response.OK(c, results)
}

func (h *MobilePatientHandler) handlePatientError(c *fiber.Ctx, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return response.Error(c, fiber.StatusNotFound, "No. RM belum terhubung ke akun mobile")
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return response.Error(c, fiber.StatusGatewayTimeout, "permintaan data pasien terlalu lama")
	}

	log.Printf("mobile patient %s failed: %v", c.Path(), err)

	return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil data pasien")
}

func mobilePatientIdentity(c *fiber.Ctx) (int64, int64, error) {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return 0, 0, errMobilePatientUnauthenticated
	}
	if !principal.HasLinkedPatient() {
		return 0, 0, errMobilePatientUnlinked
	}

	return principal.UserID, principal.PatientID, nil
}

func mobilePatientIdentityError(c *fiber.Ctx, err error) error {
	if errors.Is(err, errMobilePatientUnlinked) {
		return response.Error(c, fiber.StatusForbidden, err.Error())
	}
	return response.Error(c, fiber.StatusUnauthorized, errMobilePatientUnauthenticated.Error())
}

func queryBoolValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "download":
		return true
	default:
		return false
	}
}

func safeFileName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "pasien"
	}

	var builder strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' ||
			char == '_' {
			builder.WriteRune(char)
		}
	}

	if builder.Len() == 0 {
		return "pasien"
	}
	return builder.String()
}
