package main

import (
	"log"

	"enterprise-go-service/internal/config"
	"enterprise-go-service/internal/infrastructure/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	if err := database.RunMigrations(cfg); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("Database migrations applied successfully!")
}