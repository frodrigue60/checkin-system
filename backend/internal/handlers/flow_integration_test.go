package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestEmployeeLifecycleFlow(t *testing.T) {
	app := fiber.New()

	// Auth mock middleware setting admin user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 999) // Admin ID
		return c.Next()
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	cCache := cache.New(5*time.Minute, 10*time.Minute)
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		R2PublicURL: "http://localhost/r2",
	}

	auditService := &services.AuditService{DB: sqlxDB}
	base := AdminBase{
		DB:           sqlxDB,
		Cfg:          cfg,
		AuditService: auditService,
		Cache:        cCache,
	}

	authHandler := &AuthHandler{DB: sqlxDB, Cfg: cfg}
	shiftHandler := &ShiftHandler{AdminBase: base}
	centerHandler := &CenterHandler{AdminBase: base}
	empHandler := &EmployeeAdminHandler{AdminBase: base}

	// Register Routes
	app.Post("/auth/register", authHandler.Register)
	app.Post("/admin/shifts", shiftHandler.CreateShift)
	app.Post("/admin/centers", centerHandler.CreateCenter)
	app.Post("/admin/employees", empHandler.CreateEmployee)
	app.Put("/admin/employees/:id", empHandler.UpdateEmployee)

	// =========================================================================
	// STEP 1: Register User (POST /auth/register)
	// =========================================================================
	t.Run("1. Register User", func(t *testing.T) {
		regReq := RegisterRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Phone:    "1234567890",
			Password: "securepassword",
		}
		body, _ := json.Marshal(regReq)

		// Expectations
		mock.ExpectQuery("SELECT id FROM roles WHERE slug = 'user'").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(regReq.Name, regReq.Email, regReq.Phone, sqlmock.AnyArg(), 5).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10)) // userID = 10

		mock.ExpectQuery("SELECT u.*, r.slug as role_slug").
			WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role_id", "role_slug"}).
				AddRow(10, regReq.Name, regReq.Email, 5, "user"))

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	// =========================================================================
	// STEP 2: Create Work Shift (POST /admin/shifts)
	// =========================================================================
	t.Run("2. Create Work Shift", func(t *testing.T) {
		daysRaw, _ := json.Marshal([]int{1, 2, 3, 4, 5})
		newShift := models.WorkShift{
			Name:             "Morning Fixed",
			ExpectedCheckIn:  "09:00:00",
			ExpectedCheckOut: "18:00:00",
			AllowedLunchTime: "01:00:00",
			ToleranceTime:    "00:15:00",
			IsNightShift:     false,
			ShiftType:        "fixed",
			WorkDays:         daysRaw,
		}
		body, _ := json.Marshal(newShift)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO work_shifts").
			WithArgs(newShift.Name, newShift.ExpectedCheckIn, newShift.ExpectedCheckOut, newShift.AllowedLunchTime, newShift.ToleranceTime, newShift.IsNightShift, false, false, false, newShift.ShiftType, newShift.WorkDays, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(20)) // shiftID = 20

		mock.ExpectExec("INSERT INTO audit_logs").
			WithArgs(999, string(models.AuditActionCreateShift), "work_shift", 20, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest("POST", "/admin/shifts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	// =========================================================================
	// STEP 3: Create Work Center (POST /admin/centers)
	// =========================================================================
	t.Run("3. Create Work Center", func(t *testing.T) {
		newCenter := models.WorkCenter{
			Name:                  "HQ Office",
			Address:               stringPtr("Financial District"),
			Latitude:              19.4326,
			Longitude:             -99.1332,
			ToleranceRadiusMeters: 100,
		}
		body, _ := json.Marshal(newCenter)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO work_centers").
			WithArgs(newCenter.Name, newCenter.Address, newCenter.Latitude, newCenter.Longitude, newCenter.ToleranceRadiusMeters, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(30)) // centerID = 30

		mock.ExpectExec("INSERT INTO audit_logs").
			WithArgs(999, string(models.AuditActionCreateWorkCenter), "work_center", 30, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest("POST", "/admin/centers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	// =========================================================================
	// STEP 4: Hire Employee (POST /admin/employees)
	// =========================================================================
	t.Run("4. Hire Employee", func(t *testing.T) {
		newEmp := models.Employee{
			UserID:       10,
			PositionID:   1,
			WorkCenterID: 30,
			WorkShiftID:  20,
		}
		body, _ := json.Marshal(newEmp)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO employees").
			WithArgs(newEmp.UserID, newEmp.WorkCenterID, newEmp.WorkShiftID, newEmp.PositionID, true, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(40)) // employeeID = 40

		// Promotion check queries
		mock.ExpectExec("UPDATE users SET role_id").
			WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("INSERT INTO audit_logs").
			WithArgs(999, string(models.AuditActionCreateEmployee), "employee", 40, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest("POST", "/admin/employees", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	// =========================================================================
	// STEP 5: Update Assignment of Employee (PUT /admin/employees/:id)
	// =========================================================================
	t.Run("5. Assign new shift and center to Employee", func(t *testing.T) {
		updateEmp := models.Employee{
			WorkCenterID: 35, // New Center
			WorkShiftID:  25, // New Shift
			PositionID:   1,
			IsActive:     true,
		}
		body, _ := json.Marshal(updateEmp)

		mock.ExpectBegin()
		// Get old state
		mock.ExpectQuery("SELECT \\* FROM employees WHERE id = \\$1").
			WithArgs(40).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "work_center_id", "work_shift_id", "position_id", "is_active"}).
				AddRow(40, 10, 30, 20, 1, true))

		// Update query
		mock.ExpectExec("UPDATE employees SET").
			WithArgs(35, 25, 1, true, sqlmock.AnyArg(), 40).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Audit log
		mock.ExpectExec("INSERT INTO audit_logs").
			WithArgs(999, string(models.AuditActionUpdateEmployee), "employee", 40, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		req := httptest.NewRequest("PUT", "/admin/employees/40", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
