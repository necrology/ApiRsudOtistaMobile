package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	authmiddleware "apirusdotistamobile/internal/http/middleware"
	"apirusdotistamobile/internal/model"

	"github.com/gofiber/fiber/v2"
)

func TestMobilePatientIdentityIgnoresSpoofedQuery(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(authmiddleware.SessionPrincipalLocalKey, &model.SessionPrincipal{
			UserID:    7,
			PatientID: 9,
		})
		return c.Next()
	})
	app.Get("/identity", func(c *fiber.Ctx) error {
		userID, patientID, err := mobilePatientIdentity(c)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"user_id": userID, "patient_id": patientID})
	})

	res, err := app.Test(httptest.NewRequest(
		fiber.MethodGet,
		"/identity?email=victim@example.test&no_rm=RM-VICTIM&patient_id=999",
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var body map[string]int64
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["user_id"] != 7 || body["patient_id"] != 9 {
		t.Fatalf("identity came from request instead of session: %#v", body)
	}
}
