package database

import (
	"log"

	"attendance-api/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

var DB *sqlx.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = sqlx.Connect("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	log.Println("Database connection established")
	
	// Connection Pool Settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	RunMigrations(cfg.DBURL)
}

func GetDB() *sqlx.DB {
	return DB
}
