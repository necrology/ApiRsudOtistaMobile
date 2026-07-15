package server

import (
	"database/sql"
	"time"

	"apirusdotistamobile/internal/config"
	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/http/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(cfg config.Config, db *sql.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		BodyLimit:    1 * 1024 * 1024,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(authmiddleware.SecurityHeaders)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		MaxAge:       600,
	}))
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	}))

	routes.Register(app, db, cfg)

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found")
	})

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	message := "internal server error"

	if fiberError, ok := err.(*fiber.Error); ok {
		statusCode = fiberError.Code
		message = fiberError.Message
	}

	return response.Error(c, statusCode, message)
}
