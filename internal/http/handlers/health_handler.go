package handlers

import (
	"context"
	"database/sql"
	"time"

	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.Set(fiber.HeaderCacheControl, "no-store")
		c.Set(fiber.HeaderRetryAfter, "5")
		return response.Error(c, fiber.StatusServiceUnavailable, "service unavailable")
	}

	c.Set(fiber.HeaderCacheControl, "no-store")
	return response.OK(c, fiber.Map{
		"status": "ok",
	})
}
