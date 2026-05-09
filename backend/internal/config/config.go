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
	SwaggerHost    string
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
		DBURL:          dbURL,
		Port:           getEnv("PORT", "3000"),
		JWTSecret:      requireEnv("JWT_SECRET"),
		AppName:        getEnv("APP_NAME", "Attendance System"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		EnableSwagger:  getEnv("ENABLE_SWAGGER", "true") == "true",
		SwaggerHost:    getEnv("SWAGGER_HOST", "localhost:3000"),
	}
}

// requireEnv panics if a required environment variable is not set.
// Fail-Fast Pattern: prevents the app from running with insecure defaults.
func requireEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("FATAL: Required environment variable %s is not set. Cannot start.", key)
	}
	return value
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
