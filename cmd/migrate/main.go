package main

import (
	"log"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/database"
)

// Migrasi dipisahkan dari startup API agar akun database runtime dapat diberi
// hak DML minimum. Jalankan command ini dengan akun migrasi berwenang sebelum
// binary API baru dirilis, atau terapkan DDL history.sql secara terkontrol.
func main() {
	cfg := config.Load()
	if err := cfg.ValidateDatabase(); err != nil {
		log.Fatalf("invalid database configuration: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	if err = database.EnsureAuthSchema(db, cfg.Database.Name); err != nil {
		log.Fatalf("failed to apply auth migration: %v", err)
	}
	if err = database.EnsureHolidaySchema(db, cfg.Database.Name); err != nil {
		log.Fatalf("failed to apply holiday migration: %v", err)
	}
	if err = database.ValidateAuthSchema(db, cfg.Database.Name); err != nil {
		log.Fatalf("auth schema validation failed after migration: %v", err)
	}

	log.Println("database migration completed")
}
