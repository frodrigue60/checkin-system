package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// @Summary Authenticate user and return JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string "Success: {token: '...', user: {...}}"
// @Failure 401 {object} map[string]string "Error: Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user struct {
		models.User
		RoleSlug string `db:"role_slug"`
	}
	
	err := h.DB.Get(&user, `
		SELECT u.*, r.slug as role_slug 
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE LOWER(u.email) = LOWER($1)`, req.Email)
	
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid email or password", err)
	}

	// Bcrypt comparison
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid email or password", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"role_id":   user.RoleID,
		"role_slug": user.RoleSlug,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(h.Cfg.JWTSecret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var employeeID *int
	h.DB.Get(&employeeID, "SELECT id FROM employees WHERE user_id = $1", user.ID)

	return c.JSON(fiber.Map{
		"token": t,
		"user": fiber.Map{
			"profile":     models.MapUserToDTO(user.User, h.Cfg.R2PublicURL),
			"role_slug":   user.RoleSlug,
			"employee_id": employeeID,
		},
	})
}

// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register data"
// @Success 201 {object} map[string]string "Success: {message: '...'}"
// @Failure 400 {object} map[string]string "Error: Invalid request"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Error hashing password", err)
	}

	// Get Dynamic Role ID for 'user'
	var roleID int
	err = h.DB.Get(&roleID, "SELECT id FROM roles WHERE slug = 'user'")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server configuration error", err)
	}

	// Insert User
	query := `INSERT INTO users (name, email, phone, password, role_id, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`
	
	var userID int
	err = h.DB.QueryRow(query, req.Name, req.Email, req.Phone, string(hashedPassword), roleID).Scan(&userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusConflict, "Email already exists or database error", err)
	}

	// Get user and role slug for token generation (Automated Login)
	var user struct {
		models.User
		RoleSlug string `db:"role_slug"`
	}
	
	err = h.DB.Get(&user, `
		SELECT u.*, r.slug as role_slug 
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.id = $1`, userID)
	
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "User created but failed to generate session", err)
	}

	// Generate Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"role_id":   user.RoleID,
		"role_slug": user.RoleSlug,
		"exp":       time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(h.Cfg.JWTSecret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"token":   t,
		"user": fiber.Map{
			"profile":     models.MapUserToDTO(user.User, h.Cfg.R2PublicURL),
			"role_slug":   user.RoleSlug,
			"employee_id": nil, // New users don't have an employee profile yet
		},
	})
}




