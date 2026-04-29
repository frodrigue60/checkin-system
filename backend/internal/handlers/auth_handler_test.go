package handlers

import (
	"attendance-api/internal/config"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthHandler_Login(t *testing.T) {
	// Setup Fiber app
	app := fiber.New()

	// Setup SQL Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	handler := &AuthHandler{
		DB:  sqlxDB,
		Cfg: &config.Config{JWTSecret: "testsecret"},
	}
	app.Post("/login", handler.Login)

	// Hash a password for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("Successful Login", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    "test@test.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)

		// Expect query for user
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role_id", "role_slug"}).
			AddRow(1, "Test User", "test@test.com", string(hashedPassword), 1, "admin")
		
		mock.ExpectQuery("SELECT u.*, r.slug as role_slug").
			WithArgs(loginReq.Email).
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT id FROM employees WHERE user_id").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
		
		// Wait, the real query joins roles. Let's make it simpler or match exactly.
		// Re-reading auth_handler.go Login...
		
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		
		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    "test@test.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role_id", "role_slug"}).
			AddRow(1, "Test User", "test@test.com", string(hashedPassword), 1, "admin")
		
		mock.ExpectQuery("SELECT u.*, r.slug as role_slug").
			WithArgs(loginReq.Email).
			WillReturnRows(rows)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}




