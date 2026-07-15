package routes

import (
	"apirusdotistamobile/internal/http/handlers"
	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

func BookingRoutes(
	v1 fiber.Router,
	bookingHandler *handlers.BookingHandler,
	requireSession fiber.Handler,
) {
	booking := v1.Group("/mobile/booking")
	booking.Get("/options/:poli_id", bookingHandler.BookingOptions)
	booking.Get("/calendar", bookingHandler.BookingCalendar)
	booking.Get("/general", func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found")
	})

	patientBooking := booking.Group("", requireSession, authmiddleware.RequireLinkedPatient)
	patientBooking.Get("/general/mine", bookingHandler.ListMyGeneralBookings)
	patientBooking.Post("/general", bookingHandler.CreateGeneralBooking)
}
