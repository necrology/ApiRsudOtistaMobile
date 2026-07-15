package main

import (
	"fmt"
	"log"

	"apirusdotistamobile/internal/config"
	"apirusdotistamobile/internal/database"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	if err = database.EnsureAuthSchema(db, cfg.Database.Name); err != nil {
		log.Fatalf("ensure auth schema: %v", err)
	}

	fmt.Println("auth mobile migration applied")
}
