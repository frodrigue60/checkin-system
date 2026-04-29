package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL          string
	Port           string
	JWTSecret      string
	AppName        string
	AllowedOrigins string
	EnableSwagger  bool
}

func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USERNAME", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_DATABASE", "jgc_checkin")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dbURL := "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode

	return &Config{
		DBURL:     dbURL,
		Port:           getEnv("PORT", "3000"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key-change-me-in-production"),
		AppName:        getEnv("APP_NAME", "Attendance System"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		EnableSwagger:  getEnv("ENABLE_SWAGGER", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
