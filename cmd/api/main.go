package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/database"
	"apirusdotistamobile/internal/server"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateRuntime(); err != nil {
		log.Fatalf("invalid runtime configuration: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()
	if err = database.ValidateAuthSchema(db, cfg.Database.Name); err != nil {
		log.Fatalf("auth schema is not ready; run the documented migration first: %v", err)
	}

	app := server.New(cfg, db)
	addr := ":" + cfg.App.Port

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("%s running on http://localhost%s", cfg.App.Name, addr)
		serverErrors <- app.Listen(addr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("server stopped: %v", err)
		}
	case <-quit:
		log.Println("shutting down server...")
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}
}
