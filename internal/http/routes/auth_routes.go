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
}
