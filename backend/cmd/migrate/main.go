package main

import (
	"attendance-api/internal/config"
	"attendance-api/internal/database"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	database.RunMigrations(cfg.DBURL)
	log.Println("Migration runner finished.")
}
