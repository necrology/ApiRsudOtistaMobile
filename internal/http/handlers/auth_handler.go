package handlers

import (
	"apirusdotistamobile/internal/auth"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{
		Service: service,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {

	var req auth.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.Register(req)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "register success",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var req auth.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.Login(req)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "login success",
	})

}
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {

	token := c.Query("token")

	err := h.Service.VerifyEmail(token)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "email verified",
	})
}

func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {

	var req auth.VerifyOTPRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.VerifyOTP(req)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "otp verified",
	})
}
