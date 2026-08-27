package routes

import (
	"time"

	"apirusdotistamobile/internal/http/handlers"
	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(
	v1 fiber.Router,
	authHandler *handlers.AuthHandler,
	requireSession fiber.Handler,
) {
	auth := v1.Group("/auth")
	limitAuthWork := authmiddleware.AuthWorkLimit()

	auth.Post("/register", authmiddleware.GlobalRateLimit(30, time.Minute), authmiddleware.RateLimit(5, time.Minute, "email"), limitAuthWork, authHandler.Register)
	auth.Post("/login", authmiddleware.GlobalRateLimit(120, time.Minute), authmiddleware.RateLimit(10, time.Minute, "identifier", "email", "no_rm", "noRm", "username"), limitAuthWork, authHandler.Login)
	auth.Post("/verify-otp", authmiddleware.GlobalRateLimit(180, time.Minute), authmiddleware.RateLimit(10, time.Minute, "identifier", "email", "username"), limitAuthWork, authHandler.VerifyOTPLogin)
	auth.Post("/verify-otp-new-user", authmiddleware.GlobalRateLimit(120, time.Minute), authmiddleware.RateLimit(10, time.Minute, "email"), limitAuthWork, authHandler.VerifyOTPNewUser)
	auth.Post("/set-password", authmiddleware.GlobalRateLimit(60, time.Minute), authmiddleware.RateLimit(10, time.Minute, "registration_ticket"), limitAuthWork, authHandler.SetPassword)
	auth.Post("/forgot-password", authmiddleware.GlobalRateLimit(30, time.Minute), authmiddleware.RateLimit(5, time.Minute, "identifier", "email"), limitAuthWork, authHandler.ForgotPassword)
	auth.Post("/reset-password", authmiddleware.GlobalRateLimit(60, time.Minute), authmiddleware.RateLimit(10, time.Minute, "identifier", "email"), limitAuthWork, authHandler.ResetPassword)
	auth.Post("/account-deletion/web/request", authmiddleware.GlobalRateLimit(30, time.Minute), authmiddleware.RateLimit(5, 15*time.Minute, "identifier", "email"), limitAuthWork, authHandler.RequestAccountDeletionWeb)
	auth.Post("/account-deletion/web/confirm", authmiddleware.GlobalRateLimit(60, time.Minute), authmiddleware.RateLimit(10, 15*time.Minute, "identifier", "email"), limitAuthWork, authHandler.ConfirmAccountDeletionWeb)
	auth.Post("/refresh", authmiddleware.GlobalRateLimit(300, time.Minute), authmiddleware.RateLimit(30, time.Minute, "refresh_token"), limitAuthWork, authHandler.RefreshSession)
	// Endpoint token verifikasi legacy dipensiunkan; OTP adalah satu-satunya
	// jalur verifikasi email yang aktif.
	auth.Get("/verify", func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found")
	})

	protected := auth.Group("", requireSession)
	protected.Get("/me", authHandler.Me)
	protected.Post("/logout", authHandler.Logout)
	protected.Post("/medical-record/request", authmiddleware.GlobalRateLimit(30, time.Minute), authmiddleware.RateLimit(5, 5*time.Minute), limitAuthWork, authHandler.RequestMedicalRecordClaim)
	protected.Post("/medical-record/confirm", authmiddleware.GlobalRateLimit(60, time.Minute), authmiddleware.RateLimit(10, 5*time.Minute), limitAuthWork, authHandler.ConfirmMedicalRecordClaim)
	protected.Post("/account-deletion/request", authmiddleware.GlobalRateLimit(30, time.Minute), authmiddleware.RateLimit(5, 15*time.Minute), limitAuthWork, authHandler.RequestAccountDeletion)
	protected.Post("/account-deletion/confirm", authmiddleware.GlobalRateLimit(60, time.Minute), authmiddleware.RateLimit(10, 15*time.Minute), limitAuthWork, authHandler.ConfirmAccountDeletion)
}
