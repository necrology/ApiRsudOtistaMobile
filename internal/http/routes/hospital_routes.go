package routes

import (
	"apirusdotistamobile/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func HospitalRoutes(v1 fiber.Router, hospitalHandler *handlers.HospitalHandler) {
	// Alias /polis dipertahankan untuk kompatibilitas aplikasi lama, tetapi
	// sekarang memakai proyeksi aman dan bukan handler tabel generik.
	v1.Get("/polis", hospitalHandler.Polyclinics)

	hospital := v1.Group("/mobile/hospital")

	hospital.Get("/polyclinics", hospitalHandler.Polyclinics)
	hospital.Get("/room-availabilities", hospitalHandler.RoomAvailabilities)
}
