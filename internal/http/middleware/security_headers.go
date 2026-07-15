package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders menambahkan baseline header yang dapat diterapkan langsung
// oleh aplikasi. HSTS tetap menjadi tanggung jawab terminasi TLS/reverse proxy.
func SecurityHeaders(c *fiber.Ctx) error {
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderXFrameOptions, "DENY")
	c.Set("Referrer-Policy", "no-referrer")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

	path := c.Path()
	if strings.HasPrefix(path, "/api/v1/auth/") ||
		strings.HasPrefix(path, "/api/v1/mobile/patient/") ||
		strings.HasPrefix(path, "/api/v1/mobile/booking/general") {
		c.Set(fiber.HeaderCacheControl, "no-store")
		c.Set("Pragma", "no-cache")
	}

	return c.Next()
}
