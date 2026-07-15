package middleware

import (
	"context"
	"errors"
	"strings"

	"apirusdotistamobile/internal/auth"
	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/model"

	"github.com/gofiber/fiber/v2"
)

const SessionPrincipalLocalKey = "session_principal_mobile"

var ErrInvalidAuthorizationHeader = errors.New("invalid bearer authorization header")

// SessionAuthenticator is implemented by auth.SessionService and kept narrow
// so middleware tests and future identity providers do not require a database.
type SessionAuthenticator interface {
	Authenticate(context.Context, string) (*model.SessionPrincipal, error)
}

// RequireSession rejects missing or invalid access tokens with a uniform 401
// response and places the trusted principal in Fiber locals on success.
func RequireSession(authenticator SessionAuthenticator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if authenticator == nil {
			return authenticationUnavailable(c)
		}

		accessToken, err := ParseBearerToken(c.Get(fiber.HeaderAuthorization))
		if err != nil {
			return unauthorized(c)
		}

		principal, err := authenticator.Authenticate(c.UserContext(), accessToken)
		if errors.Is(err, auth.ErrInvalidAccessToken) || (err == nil && principal == nil) {
			return unauthorized(c)
		}
		if err != nil {
			return authenticationUnavailable(c)
		}

		c.Locals(SessionPrincipalLocalKey, principal)
		return c.Next()
	}
}

func ParseBearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", ErrInvalidAuthorizationHeader
	}
	return parts[1], nil
}

// PrincipalFromContext returns the identity established by RequireSession.
func PrincipalFromContext(c *fiber.Ctx) (*model.SessionPrincipal, bool) {
	principal, ok := c.Locals(SessionPrincipalLocalKey).(*model.SessionPrincipal)
	return principal, ok && principal != nil
}

// RequireLinkedPatient can be placed after RequireSession on medical routes.
// It prevents unlinked accounts from reaching handlers that require a patient.
func RequireLinkedPatient(c *fiber.Ctx) error {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		return unauthorized(c)
	}
	if !principal.HasLinkedPatient() {
		return response.Error(c, fiber.StatusForbidden, "akun belum terhubung dengan rekam medis")
	}
	return c.Next()
}

func unauthorized(c *fiber.Ctx) error {
	c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="rsud-otista-mobile"`)
	return response.Error(c, fiber.StatusUnauthorized, "autentikasi diperlukan")
}

func authenticationUnavailable(c *fiber.Ctx) error {
	c.Set(fiber.HeaderRetryAfter, "5")
	return response.Error(c, fiber.StatusServiceUnavailable, "layanan autentikasi sementara tidak tersedia")
}
