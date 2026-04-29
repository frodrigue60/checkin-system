package main

import (
	"attendance-api/internal/config"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	
	db, err := sqlx.Connect("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Applying schema fixes...")
	_, err = db.Exec(`ALTER TABLE incidents ADD COLUMN IF NOT EXISTS metadata_json JSONB;`)
	if err != nil {
		log.Fatalf("Error applying schema fix: %v", err)
	}

	fmt.Println("Reading seed file...")
	content, err := ioutil.ReadFile("database/seeders/large_seed.sql")
	if err != nil {
		log.Fatalf("Error reading seed file: %v", err)
	}

	fmt.Println("Executing seed SQL...")
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Error executing seed SQL: %v", err)
	}

	fmt.Println("Database seeded successfully!")
}
