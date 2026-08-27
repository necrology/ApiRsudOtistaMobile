package server

import (
	"database/sql"
	"io"
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
		{method: fiber.MethodPost, path: "/api/v1/auth/account-deletion/request", body: `{}`, want: fiber.StatusUnauthorized},
		{method: fiber.MethodPost, path: "/api/v1/auth/account-deletion/confirm", body: `{}`, want: fiber.StatusUnauthorized},
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

func TestPublicPrivacyAndAccountDeletionPages(t *testing.T) {
	app := New(config.Config{App: config.AppConfig{Name: "security-test"}}, &sql.DB{})

	tests := []struct {
		path           string
		bodyContains   string
		cacheControl   string
		requireNoStore bool
	}{
		{
			path:         "/privacy-policy",
			bodyContains: "Kebijakan Privasi SIPANTES",
			cacheControl: "public, max-age=3600",
		},
		{
			path:           "/account-deletion",
			bodyContains:   "Penghapusan Akun",
			requireNoStore: true,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			res, err := app.Test(httptest.NewRequest(fiber.MethodGet, test.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != fiber.StatusOK {
				t.Fatalf("status = %d, want %d", res.StatusCode, fiber.StatusOK)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), test.bodyContains) {
				t.Fatalf("body does not contain %q", test.bodyContains)
			}
			if res.Header.Get("Content-Security-Policy") == "" {
				t.Fatal("Content-Security-Policy header is missing")
			}
			if test.requireNoStore {
				if res.Header.Get(fiber.HeaderCacheControl) != "no-store" {
					t.Fatalf("cache-control = %q", res.Header.Get(fiber.HeaderCacheControl))
				}
			} else if res.Header.Get(fiber.HeaderCacheControl) != test.cacheControl {
				t.Fatalf("cache-control = %q", res.Header.Get(fiber.HeaderCacheControl))
			}
		})
	}
}
