package handlers

import (
	"database/sql"
	"errors"
	"log"

	"apirusdotistamobile/internal/auth"
	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/model"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Service  *auth.Service
	Sessions *auth.SessionService
}

func NewAuthHandler(service *auth.Service, sessions *auth.SessionService) *AuthHandler {
	return &AuthHandler{
		Service:  service,
		Sessions: sessions,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {

	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.RegisterUser(req)

	if err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal memproses registrasi")
	}

	return c.JSON(fiber.Map{
		"message": "otp terkirim ke email, silahkan cek email untuk verifikasi",
	})
}

func (h *AuthHandler) VerifyOTPNewUser(c *fiber.Ctx) error {

	var req model.VerifyOTPNewUser

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	ticket, err := h.Service.VerifyOTPNewUser(c.UserContext(), req)

	if err != nil {
		return handleAuthError(c, err, fiber.StatusUnauthorized, "gagal memverifikasi otp")
	}

	return c.JSON(fiber.Map{
		"message": "Register otp verified",
		"data": fiber.Map{
			"registration_ticket": ticket.Token,
			"expires_at":          ticket.ExpiresAt,
		},
	})
}

func (h *AuthHandler) SetPassword(
	c *fiber.Ctx,
) error {

	var req model.SetPasswordRequest

	if err := c.BodyParser(&req); err != nil {

		return c.Status(400).JSON(
			fiber.Map{
				"message": "invalid request body",
			},
		)
	}

	user, err := h.Service.SetPassword(c.UserContext(), req)

	if err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal menyelesaikan registrasi")
	}

	tokens, err := h.Sessions.Issue(c.UserContext(), user.ID)
	if err != nil {
		return authUnavailable(c, err, "gagal membuat sesi akun")
	}

	return c.JSON(fiber.Map{
		"message": "register success",
		"data":    userIdentityWithSession(user, tokens),
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.Login(req)

	if err != nil {
		return handleAuthError(c, err, fiber.StatusUnauthorized, "gagal memproses login")
	}

	return c.JSON(fiber.Map{
		"message": "login success",
	})

}

func (h *AuthHandler) VerifyOTPLogin(c *fiber.Ctx) error {

	var req model.VerifyOTPRequestLogin

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	user, err := h.Service.VerifyOTPLogin(req)

	if err != nil {
		return handleAuthError(c, err, fiber.StatusUnauthorized, "gagal memverifikasi otp")
	}

	tokens, err := h.Sessions.Issue(c.UserContext(), user.ID)
	if err != nil {
		return authUnavailable(c, err, "gagal membuat sesi akun")
	}

	return c.JSON(fiber.Map{
		"message": "otp verified",
		"data":    userIdentityWithSession(user, tokens),
	})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req model.ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	if err := h.Service.ForgotPassword(req); err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal memproses reset password")
	}

	return c.JSON(fiber.Map{
		"message": "otp reset password terkirim ke email",
	})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req model.ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.ResetPassword(req)
	if err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal memperbarui password")
	}

	return c.JSON(fiber.Map{
		"message": "password berhasil diperbarui",
	})
}

func (h *AuthHandler) RequestMedicalRecordClaim(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}

	var payload struct {
		Password  string `json:"password"`
		NoRM      string `json:"no_rm"`
		NIK       string `json:"nik"`
		BirthDate string `json:"birth_date"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	req := model.RequestMedicalRecordClaim{
		Password:  payload.Password,
		NoRM:      payload.NoRM,
		NIK:       payload.NIK,
		BirthDate: payload.BirthDate,
	}

	if err := h.Service.RequestMedicalRecordClaim(principal.UserID, req); err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal memproses klaim rekam medis")
	}

	return c.JSON(fiber.Map{
		"message": "otp verifikasi no rm terkirim ke email",
	})
}

func (h *AuthHandler) ConfirmMedicalRecordClaim(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}

	var payload struct {
		OTP string `json:"otp"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}
	req := model.ConfirmMedicalRecordClaim{OTP: payload.OTP}

	user, err := h.Service.ConfirmMedicalRecordClaim(principal.UserID, req)
	if err != nil {
		return handleAuthError(c, err, fiber.StatusBadRequest, "gagal mengonfirmasi klaim rekam medis")
	}

	return c.JSON(fiber.Map{
		"message": "no rm berhasil terhubung",
		"data":    userIdentityResponse(user),
	})
}

func (h *AuthHandler) RefreshSession(c *fiber.Ctx) error {
	var req model.RefreshSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	tokens, err := h.Sessions.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrRefreshTooSoon):
			c.Set(fiber.HeaderRetryAfter, "2")
			return response.Error(c, fiber.StatusTooManyRequests, "refresh terlalu cepat, silakan coba lagi")
		case errors.Is(err, auth.ErrInvalidRefreshToken),
			errors.Is(err, auth.ErrRefreshTokenExpired),
			errors.Is(err, auth.ErrRefreshTokenReused):
			return response.Error(c, fiber.StatusUnauthorized, "refresh token tidak valid atau kedaluwarsa")
		default:
			return authUnavailable(c, err, "gagal memperbarui sesi")
		}
	}

	return response.OK(c, tokens)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}

	if err := h.Sessions.RevokeFamily(
		c.UserContext(),
		principal.FamilyID,
		"logout",
	); err != nil {
		return authUnavailable(c, err, "gagal mengakhiri sesi")
	}

	return response.OK(c, fiber.Map{"message": "logout berhasil"})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	principal, ok := authmiddleware.PrincipalFromContext(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
	}

	user, err := h.Service.Repo.FindUserByID(principal.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return response.Error(c, fiber.StatusUnauthorized, "sesi akun tidak valid")
	}
	if err != nil {
		return authUnavailable(c, err, "gagal mengambil data akun")
	}

	return response.OK(c, userIdentityResponse(user))
}

func handleAuthError(c *fiber.Ctx, err error, clientStatus int, fallback string) error {
	if message, ok := auth.ClientErrorMessage(err); ok {
		return response.Error(c, clientStatus, message)
	}
	if errors.Is(err, auth.ErrMailQueueFull) || errors.Is(err, auth.ErrMailUnavailable) {
		return authUnavailable(c, err, "layanan pengiriman otp sementara tidak tersedia")
	}

	log.Printf("auth request %s failed error_type=%T", c.Path(), err)
	return response.Error(c, fiber.StatusInternalServerError, fallback)
}

func authUnavailable(c *fiber.Ctx, err error, fallback string) error {
	log.Printf("auth dependency %s failed error_type=%T", c.Path(), err)
	c.Set(fiber.HeaderRetryAfter, "5")
	return response.Error(c, fiber.StatusServiceUnavailable, fallback)
}

func userIdentityResponse(user *model.User) fiber.Map {
	return fiber.Map{
		"id":                  user.Email,
		"patientId":           user.PatientID,
		"fullName":            user.FullName,
		"email":               user.Email,
		"phoneNumber":         user.Phone,
		"medicalRecordNumber": user.NoRM,
		"familyMembers":       []string{},
	}
}

func userIdentityWithSession(user *model.User, tokens *model.SessionTokenPair) fiber.Map {
	data := userIdentityResponse(user)
	data["token_type"] = tokens.TokenType
	data["access_token"] = tokens.AccessToken
	data["access_expires_at"] = tokens.AccessExpiresAt
	data["access_expires_in"] = tokens.AccessExpiresIn
	data["refresh_token"] = tokens.RefreshToken
	data["refresh_expires_at"] = tokens.RefreshExpiresAt
	data["refresh_expires_in"] = tokens.RefreshExpiresIn
	return data
}
