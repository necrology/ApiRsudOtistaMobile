package handlers

import (
	"context"
	"database/sql"
	"time"

	"apirusdotistamobile/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	db        *sql.DB
	startedAt time.Time
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startedAt: time.Now(),
	}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	databaseStatus := "ok"
	if err := h.db.PingContext(ctx); err != nil {
		databaseStatus = "error"
	}

	return response.OK(c, fiber.Map{
		"status":   "ok",
		"database": databaseStatus,
		"uptime":   time.Since(h.startedAt).String(),
	})
}
