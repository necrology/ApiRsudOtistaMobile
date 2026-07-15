package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAuthWorkLimitRejectsConcurrentOverflow(t *testing.T) {
	slots := make(chan struct{}, 1)
	slots <- struct{}{}

	app := fiber.New()
	app.Post("/auth", authWorkLimit(slots), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/auth", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
	if res.Header.Get(fiber.HeaderRetryAfter) != "1" {
		t.Fatalf("Retry-After = %q, want 1", res.Header.Get(fiber.HeaderRetryAfter))
	}
}

func TestAuthWorkLimitReleasesSlotAfterRequest(t *testing.T) {
	slots := make(chan struct{}, 1)
	app := fiber.New()
	app.Post("/auth", authWorkLimit(slots), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for requestNumber := 1; requestNumber <= 2; requestNumber++ {
		res, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/auth", nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("request %d status = %d, want 200", requestNumber, res.StatusCode)
		}
	}
}
