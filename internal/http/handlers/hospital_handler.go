package handlers

import (
	"context"
	"log"
	"strings"
	"time"

	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type HospitalHandler struct {
	repository *repository.HospitalRepository
}

func NewHospitalHandler(repository *repository.HospitalRepository) *HospitalHandler {
	return &HospitalHandler{repository: repository}
}

func (h *HospitalHandler) Polyclinics(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 8*time.Second)
	defer cancel()

	search := strings.TrimSpace(c.Query("q"))
	if search == "" {
		search = strings.TrimSpace(c.Query("search"))
	}

	items, pagination, err := h.repository.Polyclinics(
		ctx,
		search,
		c.QueryInt("page", 1),
		c.QueryInt("limit", 20),
	)
	if err != nil {
		log.Printf("polyclinics failed: %v", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil data poliklinik")
	}

	return response.Paginated(c, items, pagination)
}

func (h *HospitalHandler) RoomAvailabilities(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()

	search := strings.TrimSpace(c.Query("q"))
	if search == "" {
		search = strings.TrimSpace(c.Query("search"))
	}

	items, err := h.repository.RoomAvailabilities(ctx, search, c.QueryInt("limit", 100))
	if err != nil {
		log.Printf("room availabilities failed: %v", err)
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil ketersediaan kamar")
	}

	return response.OK(c, fiber.Map{
		"items": items,
		"total": len(items),
	})
}
