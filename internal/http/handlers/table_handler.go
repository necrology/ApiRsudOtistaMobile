package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"apirusdotistamobile/internal/http/response"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type TableHandler struct {
	repository *repository.ResourceRepository
}

func NewTableHandler(repository *repository.ResourceRepository) *TableHandler {
	return &TableHandler{repository: repository}
}

func (h *TableHandler) Tables(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	tables, err := h.repository.Tables(ctx)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "gagal mengambil daftar tabel")
	}

	return response.OK(c, tables)
}

func (h *TableHandler) TableInfo(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	table, err := h.repository.Table(ctx, c.Params("table"))
	if err != nil {
		return h.handleRepositoryError(c, err, "gagal mengambil informasi tabel")
	}

	return response.OK(c, table)
}

func (h *TableHandler) Search(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	search := c.Query("q")
	if strings.TrimSpace(search) == "" {
		search = c.Query("search")
	}

	records, pagination, err := h.repository.Search(ctx, c.Params("table"), repository.ListOptions{
		Page:          c.QueryInt("page", 1),
		Limit:         c.QueryInt("limit", 20),
		Search:        search,
		Columns:       csvQuery(c.Query("columns")),
		SearchColumns: csvQuery(c.Query("search_columns")),
		WithTotal:     queryBool(c, "with_total") || queryBool(c, "include_total"),
	})
	if err != nil {
		return h.handleRepositoryError(c, err, "gagal mengambil data")
	}

	return response.Paginated(c, records, pagination)
}

func (h *TableHandler) Detail(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	record, err := h.repository.FindByID(ctx, c.Params("table"), c.Params("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, fiber.StatusNotFound, "data tidak ditemukan")
		}

		return h.handleRepositoryError(c, err, "gagal mengambil detail data")
	}

	return response.OK(c, record)
}

func (h *TableHandler) handleRepositoryError(c *fiber.Ctx, err error, fallback string) error {
	if errors.Is(err, repository.ErrTableNotFound) {
		return response.Error(c, fiber.StatusNotFound, "tabel tidak ditemukan")
	}

	var validationError *repository.ValidationError
	if errors.As(err, &validationError) {
		return response.Error(c, fiber.StatusBadRequest, validationError.Message)
	}

	return response.Error(c, fiber.StatusInternalServerError, fallback)
}

func csvQuery(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}

	return items
}

func queryBool(c *fiber.Ctx, key string) bool {
	switch strings.ToLower(strings.TrimSpace(c.Query(key))) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
