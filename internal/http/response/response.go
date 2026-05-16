package response

import "github.com/gofiber/fiber/v2"

func OK(c *fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func Paginated(c *fiber.Ctx, data any, pagination any) error {
	return c.JSON(fiber.Map{
		"success":    true,
		"data":       data,
		"pagination": pagination,
	})
}

func Error(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
