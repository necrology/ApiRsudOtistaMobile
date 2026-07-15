package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestRateLimitSubjectCanonicalizesJSONKeysAndEmail(t *testing.T) {
	lower := rateLimitSubject([]byte(`{"email":"patient@example.test"}`), []string{"email"})
	flutterStyle := rateLimitSubject([]byte(`{"Email":"Patient@Example.Test"}`), []string{"email"})
	displayName := rateLimitSubject([]byte(`{"EMAIL":"Patient <patient@example.test>"}`), []string{"email"})

	if lower == "" || lower != flutterStyle || lower != displayName {
		t.Fatalf("canonical subjects differ: lower=%q flutter=%q display=%q", lower, flutterStyle, displayName)
	}
	if strings.Contains(lower, "patient") || strings.Contains(lower, "@") {
		t.Fatal("rate limiter key contains raw PII")
	}
}

func TestRateLimitUsesIndependentHashedSubjects(t *testing.T) {
	app := fiber.New()
	app.Post("/auth", RateLimit(1, time.Minute, "email"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	request := func(body string) int {
		req := httptest.NewRequest(fiber.MethodPost, "/auth", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		return res.StatusCode
	}

	if got := request(`{"Email":"one@example.test"}`); got != fiber.StatusOK {
		t.Fatalf("first status = %d, want 200", got)
	}
	if got := request(`{"email":"ONE@example.test"}`); got != fiber.StatusTooManyRequests {
		t.Fatalf("same subject status = %d, want 429", got)
	}
	if got := request(`{"email":"two@example.test"}`); got != fiber.StatusOK {
		t.Fatalf("independent subject status = %d, want 200", got)
	}
}

func TestGlobalRateLimitCannotBeBypassedByChangingSubject(t *testing.T) {
	app := fiber.New()
	app.Post("/auth", GlobalRateLimit(1, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for index, body := range []string{
		`{"email":"one@example.test"}`,
		`{"email":"two@example.test"}`,
	} {
		req := httptest.NewRequest(fiber.MethodPost, "/auth", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		want := fiber.StatusOK
		if index == 1 {
			want = fiber.StatusTooManyRequests
		}
		if res.StatusCode != want {
			t.Fatalf("request %d status = %d, want %d", index+1, res.StatusCode, want)
		}
	}
}
