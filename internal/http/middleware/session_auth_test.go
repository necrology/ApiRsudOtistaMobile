package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"apirusdotistamobile/internal/auth"
	"apirusdotistamobile/internal/model"

	"github.com/gofiber/fiber/v2"
)

type fakeAuthenticator struct {
	token     string
	principal *model.SessionPrincipal
	err       error
}

func (f *fakeAuthenticator) Authenticate(_ context.Context, token string) (*model.SessionPrincipal, error) {
	f.token = token
	return f.principal, f.err
}

func TestRequireSessionStoresTrustedPrincipal(t *testing.T) {
	authenticator := &fakeAuthenticator{
		principal: &model.SessionPrincipal{SessionID: 3, UserID: 7, PatientID: 9, Email: "patient@example.test", NoRM: "00123"},
	}
	app := fiber.New()
	app.Get("/private", RequireSession(authenticator), func(c *fiber.Ctx) error {
		principal, ok := PrincipalFromContext(c)
		if !ok {
			return fiber.ErrInternalServerError
		}
		return c.JSON(fiber.Map{"user_id": principal.UserID, "no_rm": principal.NoRM})
	})

	request := httptest.NewRequest("GET", "/private", nil)
	request.Header.Set("Authorization", "Bearer opaque-token")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if authenticator.token != "opaque-token" {
		t.Fatalf("authenticator token = %q", authenticator.token)
	}
}

func TestRequireSessionRejectsMissingAndInvalidTokensUniformly(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		authError     error
		wantStatus    int
	}{
		{name: "missing header", wantStatus: fiber.StatusUnauthorized},
		{name: "wrong scheme", authorization: "Basic abc", wantStatus: fiber.StatusUnauthorized},
		{name: "invalid token", authorization: "Bearer invalid", authError: auth.ErrInvalidAccessToken, wantStatus: fiber.StatusUnauthorized},
		{name: "session store unavailable", authorization: "Bearer opaque-token", authError: errors.New("database timeout"), wantStatus: fiber.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authenticator := &fakeAuthenticator{err: test.authError}
			app := fiber.New()
			app.Get("/private", RequireSession(authenticator), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			request := httptest.NewRequest("GET", "/private", nil)
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			}
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
			if test.wantStatus == fiber.StatusUnauthorized && response.Header.Get("WWW-Authenticate") == "" {
				t.Fatal("missing WWW-Authenticate header")
			}
			if test.wantStatus == fiber.StatusServiceUnavailable && response.Header.Get("Retry-After") == "" {
				t.Fatal("missing Retry-After header")
			}
		})
	}
}

func TestRequireLinkedPatient(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(SessionPrincipalLocalKey, &model.SessionPrincipal{UserID: 7})
		return c.Next()
	})
	app.Get("/medical", RequireLinkedPatient, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	response, err := app.Test(httptest.NewRequest("GET", "/medical", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.StatusCode)
	}
}
