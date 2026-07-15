package routes

import (
	"context"
	"database/sql"
	"time"

	"apirusdotistamobile/internal/auth"
	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/http/handlers"
	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

func Register(
	app *fiber.App,
	db *sql.DB,
	cfg config.Config,
) {
	authRepo := repository.NewAuthRepository(db)
	authService := auth.NewService(
		authRepo,
		cfg.SMTP,
	)
	sessionRepo := repository.NewSessionUserMobileRepository(db)
	sessionService := auth.NewSessionService(sessionRepo)
	requireSession := authmiddleware.RequireSession(sessionService)
	authHandler := handlers.NewAuthHandler(authService, sessionService)
	app.Hooks().OnShutdown(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		return authService.Close(ctx)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/api/v1/health")
	})

	v1 := app.Group("/api/v1")

	healthHandler := handlers.NewHealthHandler(db)
	v1.Get("/health", healthHandler.Check)

	mobilePatientRepo := repository.NewMobilePatientRepository(db)
	mobilePatientHandler := handlers.NewMobilePatientHandler(mobilePatientRepo)
	hospitalRepo := repository.NewHospitalRepository(db)
	hospitalHandler := handlers.NewHospitalHandler(hospitalRepo)
	holidayRepo := repository.NewHolidayRepository(db, cfg.Holiday)
	bookingRepo := repository.NewBookingRepository(db, holidayRepo)
	bookingHandler := handlers.NewBookingHandler(bookingRepo, holidayRepo)

	// Route database generik sengaja tidak diregistrasikan. Seluruh resource
	// publik harus memakai handler dengan proyeksi kolom yang eksplisit.
	MobilePatientRoutes(v1, mobilePatientHandler, requireSession)
	HospitalRoutes(v1, hospitalHandler)
	BookingRoutes(v1, bookingHandler, requireSession)
	AuthRoutes(v1, authHandler, requireSession)
}
