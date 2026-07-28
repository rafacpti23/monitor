package main

import (
	"log"

	"p-mon-backend/internal/config"
	"p-mon-backend/internal/db"
)

func main() {
	if err := config.Load("config.json"); err != nil {
		log.Println("Note: Using default config because config.json could not be loaded:", err)
	}

	if err := db.InitDB(config.Cfg.DBPath); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.DB.Close()

	if err := db.Migrate("internal/db/schema.sql"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
