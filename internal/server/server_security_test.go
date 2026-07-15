package server

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	"apirusdotistamobile/internal/config"

	"github.com/gofiber/fiber/v2"
)

func TestSensitiveAndGenericRoutesAreNotPublic(t *testing.T) {
	app := New(config.Config{App: config.AppConfig{Name: "security-test"}}, &sql.DB{})

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: fiber.MethodGet, path: "/api/v1/tables", want: fiber.StatusNotFound},
		{method: fiber.MethodGet, path: "/api/v1/user_mobile", want: fiber.StatusNotFound},
		{method: fiber.MethodGet, path: "/api/v1/session_user_mobile/1", want: fiber.StatusNotFound},
		{method: fiber.MethodGet, path: "/api/v1/auth/verify?token=legacy", want: fiber.StatusNotFound},
		{method: fiber.MethodGet, path: "/api/v1/mobile/booking/general", want: fiber.StatusNotFound},
		{method: fiber.MethodGet, path: "/api/v1/mobile/patient/profile?email=victim@example.test&no_rm=RM1", want: fiber.StatusUnauthorized},
		{method: fiber.MethodGet, path: "/api/v1/mobile/booking/general/mine", want: fiber.StatusUnauthorized},
		{method: fiber.MethodPost, path: "/api/v1/mobile/booking/general", body: `{}`, want: fiber.StatusUnauthorized},
		{method: fiber.MethodPost, path: "/api/v1/auth/medical-record/request", body: `{}`, want: fiber.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.body != "" {
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			}
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != test.want {
				t.Fatalf("status = %d, want %d", res.StatusCode, test.want)
			}
		})
	}
}
