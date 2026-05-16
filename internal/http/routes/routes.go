package routes

import (
	"database/sql"

	"apirusdotistamobile/internal/http/handlers"
	"apirusdotistamobile/internal/repository"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, db *sql.DB, databaseName string) {
	repo := repository.NewResourceRepository(db, databaseName)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/api/v1/health")
	})

	v1 := app.Group("/api/v1")

	healthHandler := handlers.NewHealthHandler(db)
	v1.Get("/health", healthHandler.Check)

	tableHandler := handlers.NewTableHandler(repo)
	v1.Get("/tables", tableHandler.Tables)
	v1.Get("/tables/:table/search", tableHandler.Search)
	v1.Get("/tables/:table", tableHandler.TableInfo)

	v1.Get("/:table/search", tableHandler.Search)
	v1.Get("/:table", tableHandler.Search)
	v1.Get("/:table/:id", tableHandler.Detail)
}
