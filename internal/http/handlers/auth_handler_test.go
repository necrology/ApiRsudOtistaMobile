package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"apirusdotistamobile/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func TestHandleAuthErrorReturnsAccountAlreadyRegisteredContract(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return handleAuthError(c, auth.ErrAccountAlreadyRegistered, fiber.StatusBadRequest, "fallback")
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusConflict)
	}

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success {
		t.Fatal("success = true, want false")
	}
	if body.Code != "account_already_registered" {
		t.Fatalf("code = %q", body.Code)
	}
	if body.Message != "Akun dengan email tersebut sudah terdaftar." {
		t.Fatalf("message = %q", body.Message)
	}
}
