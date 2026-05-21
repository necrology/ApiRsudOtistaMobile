package routes

import (
	"apirusdotistamobile/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(
	v1 fiber.Router,
	authHandler *handlers.AuthHandler,
) {

	auth := v1.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/verify", authHandler.VerifyEmail)
	auth.Post("/verify-otp", authHandler.VerifyOTPLogin)
	auth.Post("/verify-otp-new-user", authHandler.VerifyOTPNewUser)
	auth.Post("/set-password", authHandler.SetPassword)
}
