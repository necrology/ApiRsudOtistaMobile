package routes

import (
	"apirusdotistamobile/internal/http/handlers"
	authmiddleware "apirusdotistamobile/internal/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func MobilePatientRoutes(
	v1 fiber.Router,
	mobilePatientHandler *handlers.MobilePatientHandler,
	requireSession fiber.Handler,
) {
	mobile := v1.Group("/mobile")
	patient := mobile.Group("/patient", requireSession, authmiddleware.RequireLinkedPatient)

	patient.Get("/profile", mobilePatientHandler.Profile)
	patient.Get("/visits", mobilePatientHandler.Visits)
	patient.Get("/medical-summaries", mobilePatientHandler.MedicalSummaries)
	patient.Get("/medical-summaries/:registration_id/pdf", mobilePatientHandler.MedicalSummaryPDF)
	patient.Get("/laboratory-results", mobilePatientHandler.LaboratoryResults)
	patient.Get("/radiology-results", mobilePatientHandler.RadiologyResults)
	patient.Get("/prescriptions", mobilePatientHandler.Prescriptions)
}
