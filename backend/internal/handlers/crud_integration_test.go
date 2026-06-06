package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func setupTestCenterHandler(t *testing.T) (*fiber.App, sqlmock.Sqlmock, *cache.Cache, *CenterHandler) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if len(c.Response().Body()) > 0 {
				return nil
			}
			return fiber.DefaultErrorHandler(c, err)
		},
	})
	// Mock auth middleware setting user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 42)
		return c.Next()
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")

	cCache := cache.New(5*time.Minute, 10*time.Minute)
	
	cfg := &config.Config{
		R2PublicURL: "http://localhost/r2",
	}

	auditService := &services.AuditService{DB: sqlxDB}
	
	handler := &CenterHandler{
		AdminBase: AdminBase{
			DB:           sqlxDB,
			Cfg:          cfg,
			AuditService: auditService,
			Cache:        cCache,
		},
	}

	return app, mock, cCache, handler
}

// ==========================================
// 1. GET /centers (ListCenters)
// ==========================================

func TestCenterHandler_ListCenters_Success_DBFetch(t *testing.T) {
	app, mock, cCache, handler := setupTestCenterHandler(t)
	app.Get("/centers", handler.ListCenters)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "address", "latitude", "longitude", "tolerance_radius_meters", "manager_id", "created_at", "updated_at"}).
		AddRow(1, "Center A", "123 Main St", 19.4326, -99.1332, 100, nil, now, now)

	mock.ExpectQuery("SELECT \\* FROM work_centers").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/centers", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify cache was populated
	var cachedData []models.WorkCenterDTO
	val, found := cCache.Get("centers")
	assert.True(t, found)
	cachedData = val.([]models.WorkCenterDTO)
	assert.Len(t, cachedData, 1)
	assert.Equal(t, "Center A", cachedData[0].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_ListCenters_Success_CacheHit(t *testing.T) {
	app, mock, cCache, handler := setupTestCenterHandler(t)
	app.Get("/centers", handler.ListCenters)

	// Populate cache beforehand
	cachedDTOs := []models.WorkCenterDTO{
		{ID: 2, Name: "Cached Center", Address: stringPtr("456 Main St")},
	}
	cCache.Set("centers", cachedDTOs, 5*time.Minute)

	// Expect NO DB queries because cache hit
	req := httptest.NewRequest("GET", "/centers", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody []models.WorkCenterDTO
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody, 1)
	assert.Equal(t, "Cached Center", respBody[0].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_ListCenters_DBError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Get("/centers", handler.ListCenters)

	mock.ExpectQuery("SELECT \\* FROM work_centers").WillReturnError(fmt.Errorf("db connection failure"))

	req := httptest.NewRequest("GET", "/centers", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 2. POST /centers (CreateCenter)
// ==========================================

func TestCenterHandler_CreateCenter_Success(t *testing.T) {
	app, mock, cCache, handler := setupTestCenterHandler(t)
	app.Post("/centers", handler.CreateCenter)

	// Pre-fill cache to verify it is cleared
	cCache.Set("centers", []models.WorkCenterDTO{}, 5*time.Minute)

	newCenter := models.WorkCenter{
		Name:                  "New Center",
		Address:               stringPtr("789 Oak Ave"),
		Latitude:              19.5,
		Longitude:             -99.2,
		ToleranceRadiusMeters: 150,
	}
	body, _ := json.Marshal(newCenter)

	mock.ExpectBegin()
	// Insert query named query compilation
	mock.ExpectQuery("INSERT INTO work_centers").
		WithArgs(newCenter.Name, newCenter.Address, newCenter.Latitude, newCenter.Longitude, newCenter.ToleranceRadiusMeters, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	// Audit log query
	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(42, string(models.AuditActionCreateWorkCenter), "work_center", 10, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	req := httptest.NewRequest("POST", "/centers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// Cache should be deleted
	_, found := cCache.Get("centers")
	assert.False(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_CreateCenter_ValidationError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Post("/centers", handler.CreateCenter)

	// Send malformed JSON
	req := httptest.NewRequest("POST", "/centers", bytes.NewBufferString("{invalidjson"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_CreateCenter_DBInsertError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Post("/centers", handler.CreateCenter)

	newCenter := models.WorkCenter{
		Name: "Fail Center",
	}
	body, _ := json.Marshal(newCenter)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO work_centers").WillReturnError(fmt.Errorf("unique constraint violation"))
	mock.ExpectRollback()

	req := httptest.NewRequest("POST", "/centers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_CreateCenter_AuditError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Post("/centers", handler.CreateCenter)

	newCenter := models.WorkCenter{
		Name: "Fail Audit Center",
	}
	body, _ := json.Marshal(newCenter)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO work_centers").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectExec("INSERT INTO audit_logs").WillReturnError(fmt.Errorf("audit db unavailable"))
	mock.ExpectRollback()

	req := httptest.NewRequest("POST", "/centers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 3. PUT /centers/:id (UpdateCenter)
// ==========================================

func TestCenterHandler_UpdateCenter_Success(t *testing.T) {
	app, mock, cCache, handler := setupTestCenterHandler(t)
	app.Put("/centers/:id", handler.UpdateCenter)

	cCache.Set("centers", []models.WorkCenterDTO{}, 5*time.Minute)

	updateData := models.WorkCenter{
		Name:                  "Updated Name",
		Address:               stringPtr("Updated Address"),
		Latitude:              20.0,
		Longitude:             -100.0,
		ToleranceRadiusMeters: 200,
	}
	body, _ := json.Marshal(updateData)

	mock.ExpectBegin()
	// Get old entity
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(101).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "address"}).AddRow(101, "Old Name", "Old Address"))

	// Update query
	mock.ExpectExec("UPDATE work_centers SET").
		WithArgs(updateData.Name, updateData.Address, updateData.Latitude, updateData.Longitude, updateData.ToleranceRadiusMeters, nil, sqlmock.AnyArg(), 101).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Audit log
	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(42, string(models.AuditActionUpdateWorkCenter), "work_center", 101, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	req := httptest.NewRequest("PUT", "/centers/101", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	_, found := cCache.Get("centers")
	assert.False(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_UpdateCenter_InvalidID(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Put("/centers/:id", handler.UpdateCenter)

	req := httptest.NewRequest("PUT", "/centers/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_UpdateCenter_NotFound(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Put("/centers/:id", handler.UpdateCenter)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(999).
		WillReturnError(fmt.Errorf("sql: no rows in result set"))
	mock.ExpectRollback()

	req := httptest.NewRequest("PUT", "/centers/999", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_UpdateCenter_DBUpdateError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Put("/centers/:id", handler.UpdateCenter)

	body, _ := json.Marshal(models.WorkCenter{Name: "Fail DB"})

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(102).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(102, "Old Center"))

	mock.ExpectExec("UPDATE work_centers SET").WillReturnError(fmt.Errorf("database transaction error"))
	mock.ExpectRollback()

	req := httptest.NewRequest("PUT", "/centers/102", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 4. DELETE /centers/:id (DeleteCenter)
// ==========================================

func TestCenterHandler_DeleteCenter_Success(t *testing.T) {
	app, mock, cCache, handler := setupTestCenterHandler(t)
	app.Delete("/centers/:id", handler.DeleteCenter)

	cCache.Set("centers", []models.WorkCenterDTO{}, 5*time.Minute)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(201).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(201, "ToDelete"))

	mock.ExpectExec("DELETE FROM work_centers WHERE id = \\$1").
		WithArgs("201").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(42, string(models.AuditActionDeleteWorkCenter), "work_center", 201, sqlmock.AnyArg(), nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	req := httptest.NewRequest("DELETE", "/centers/201", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	_, found := cCache.Get("centers")
	assert.False(t, found)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_DeleteCenter_NotFound(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Delete("/centers/:id", handler.DeleteCenter)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(999).
		WillReturnError(fmt.Errorf("sql: no rows in result set"))
	mock.ExpectRollback()

	req := httptest.NewRequest("DELETE", "/centers/999", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCenterHandler_DeleteCenter_DBError(t *testing.T) {
	app, mock, _, handler := setupTestCenterHandler(t)
	app.Delete("/centers/:id", handler.DeleteCenter)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM work_centers WHERE id = \\$1").
		WithArgs(202).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(202, "FailDelete"))

	mock.ExpectExec("DELETE FROM work_centers WHERE id = \\$1").
		WillReturnError(fmt.Errorf("foreign key violation: linked employees exist"))
	mock.ExpectRollback()

	req := httptest.NewRequest("DELETE", "/centers/202", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func stringPtr(s string) *string {
	return &s
}

