package handlers

import (
	"apirusdotistamobile/internal/auth"
	"apirusdotistamobile/internal/model"

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

	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.RegisterUser(req)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "otp terkirim ke email, silahkan cek email untuk verifikasi",
	})
}

func (h *AuthHandler) VerifyOTPNewUser(c *fiber.Ctx) error {

	var req model.VerifyOTPNewUser

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.VerifyOTPNewUser(req)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Register otp verified",
	})
}

func (h *AuthHandler) SetPassword(
	c *fiber.Ctx,
) error {

	var req model.SetPasswordRequest

	if err := c.BodyParser(&req); err != nil {

		return c.Status(400).JSON(
			fiber.Map{
				"message": "invalid request body",
			},
		)
	}

	err := h.Service.SetPassword(req)

	if err != nil {

		return c.Status(400).JSON(
			fiber.Map{
				"message": err.Error(),
			},
		)
	}

	return c.JSON(fiber.Map{
		"message": "register success",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var req model.LoginRequest

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

func (h *AuthHandler) VerifyOTPLogin(c *fiber.Ctx) error {

	var req model.VerifyOTPRequestLogin

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid request body",
		})
	}

	err := h.Service.VerifyOTPLogin(req)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "otp verified",
	})
}
