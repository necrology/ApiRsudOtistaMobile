package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit menyediakan pembatasan brute-force lapis aplikasi. Untuk route
// publik, identifierFields dipakai sebagai subject agar seluruh pengguna tidak
// berbagi satu bucket ketika API berada di balik reverse proxy. Subject di-hash
// sebelum masuk storage limiter sehingga email maupun token mentah tidak
// tersimpan sebagai key. State tetap per proses; edge limiter terdistribusi
// perlu ditambahkan saat akses infrastruktur tersedia.
func RateLimit(max int, expiration time.Duration, identifierFields ...string) fiber.Handler {
	if max < 1 {
		max = 1
	}
	if expiration <= 0 {
		expiration = time.Minute
	}

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			if principal, ok := PrincipalFromContext(c); ok {
				return "user:" + strconv.FormatInt(principal.UserID, 10)
			}

			if subject := rateLimitSubject(c.Body(), identifierFields); subject != "" {
				return "subject:" + subject
			}

			// Hanya request tanpa identifier valid yang jatuh ke IP socket. Header
			// forwarded sengaja tidak dipercaya tanpa daftar proxy tepercaya.
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(int(expiration/time.Second)))
			return response.Error(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan, silakan coba lagi nanti")
		},
	})
}

// GlobalRateLimit membatasi total pekerjaan berat per proses meskipun attacker
// terus mengganti identifier. Gunakan bersama RateLimit per subject, bukan
// sebagai penggantinya.
func GlobalRateLimit(max int, expiration time.Duration) fiber.Handler {
	if max < 1 {
		max = 1
	}
	if expiration <= 0 {
		expiration = time.Minute
	}

	return limiter.New(limiter.Config{
		Max:          max,
		Expiration:   expiration,
		KeyGenerator: func(*fiber.Ctx) string { return "process" },
		LimitReached: func(c *fiber.Ctx) error {
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(int(expiration/time.Second)))
			return response.Error(c, fiber.StatusTooManyRequests, "layanan autentikasi sedang sibuk, silakan coba lagi nanti")
		},
	})
}

func rateLimitSubject(body []byte, identifierFields []string) string {
	if len(body) == 0 || len(identifierFields) == 0 {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	normalizedPayload := make(map[string]any, len(payload))
	for key, value := range payload {
		normalizedPayload[strings.ToLower(strings.TrimSpace(key))] = value
	}

	for _, field := range identifierFields {
		value, ok := normalizedPayload[strings.ToLower(strings.TrimSpace(field))].(string)
		value = canonicalRateLimitValue(value)
		if !ok || value == "" {
			continue
		}

		digest := sha256.Sum256([]byte(value))
		return hex.EncodeToString(digest[:])
	}

	return ""
}

func canonicalRateLimitValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "@") {
		if address, err := mail.ParseAddress(value); err == nil && address.Address != "" {
			value = address.Address
		}
	}
	return strings.ToLower(value)
}
