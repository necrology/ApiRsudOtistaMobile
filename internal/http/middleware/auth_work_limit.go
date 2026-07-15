package middleware

import (
	"strconv"

	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

const defaultConcurrentAuthWork = 4

var processAuthWorkSlots = make(chan struct{}, defaultConcurrentAuthWork)

// AuthWorkLimit membatasi operasi auth mahal (bcrypt/OTP/rotasi sesi) secara serentak di
// seluruh route pada satu proses. Rate limit saja tidak menahan burst request
// yang tiba pada saat yang sama.
func AuthWorkLimit() fiber.Handler {
	return authWorkLimit(processAuthWorkSlots)
}

func authWorkLimit(slots chan struct{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			return c.Next()
		default:
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(1))
			return response.Error(c, fiber.StatusServiceUnavailable, "layanan autentikasi sedang sibuk, silakan coba lagi")
		}
	}
}
