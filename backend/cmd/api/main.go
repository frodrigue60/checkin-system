package main

import (
	"attendance-api/internal/config"
	"attendance-api/internal/database"
	"attendance-api/internal/handlers"
	"attendance-api/internal/middleware"
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"log"
	_ "time/tzdata"

	_ "attendance-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/swagger"
	"github.com/patrickmn/go-cache"
	"time"
)

// @title JGC Attendance Management API
// @version 1.0
// @description Enterprise-grade Attendance and Workforce Management System API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// 0. Initialize Logger
	utils.InitLogger()
	defer utils.Logger.Sync()

	// 1. Initialize Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database
	database.InitDB(cfg)
	db := database.GetDB()

	// 2.1 Initialize Cache
	c := cache.New(5*time.Minute, 10*time.Minute)

	// 3. Initialize Services and Handlers
	attendanceService := &services.AttendanceService{}
	auditService := &services.AuditService{DB: db}
	pdfService := services.NewPDFService(cfg.AppName)
	reportService := &services.ReportService{DB: db}
	authHandler := &handlers.AuthHandler{DB: db, Cfg: cfg}
	alertService := &services.AlertService{DB: db}
	justificationService := &services.JustificationService{
		DB:                db,
		AlertService:      alertService,
		AttendanceService: attendanceService,
	}

	adminHandler := &handlers.AdminHandler{
		DB:                   db,
		PDFService:           pdfService,
		AttendanceService:    attendanceService,
		AuditService:         auditService,
		ReportService:        reportService,
		AlertService:         alertService,
		JustificationService: justificationService,
		Cache:                c,
	}
	attendanceHandler := &handlers.AttendanceHandler{
		DB:                   db,
		Service:              attendanceService,
		JustificationService: justificationService,
	}
	managerHandler := &handlers.ManagerHandler{DB: db}
	reportHandler := &handlers.ReportHandler{DB: db, PDFService: pdfService, AuditService: auditService}
	userHandler := &handlers.UserHandler{DB: db}
	utilsHandler := &handlers.UtilsHandler{}

	// 3.1 Initialize Background Workers
	workerService := &services.WorkerService{
		DB:                db, 
		AttendanceService: attendanceService,
		AlertService:      alertService,
	}
	workerService.StartGhostSessionCleaner()

	// 4. Initialize Fiber App
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Swagger
	if cfg.EnableSwagger {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// API Routes
	api := app.Group("/api/v1")

	// Health Check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "up",
			"database": "connected",
		})
	})

	// Auth Endpoints (Rate limited)
	auth := api.Group("/auth", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Demasiados intentos. Por favor, intenta de nuevo en un minuto.",
			})
		},
	}))
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)

	// Attendance Endpoints
	attendance := api.Group("/attendance", middleware.JWTAuth(cfg))
	attendance.Get("/today", attendanceHandler.GetTodayStatus)
	attendance.Get("/centers", attendanceHandler.ListCenters)
	attendance.Post("/check-in", attendanceHandler.CheckIn)
	attendance.Post("/check-out", attendanceHandler.CheckOutNoID)
	attendance.Post("/check-out/:id", attendanceHandler.CheckOut)
	attendance.Post("/lunch-start", attendanceHandler.LunchStart)
	attendance.Post("/lunch-end", attendanceHandler.LunchEnd)
	attendance.Post("/report-absence", attendanceHandler.ReportAbsence)
	attendance.Post("/justify", attendanceHandler.SubmitJustification)


	// Admin Root Group
	admin := api.Group("/admin", middleware.JWTAuth(cfg))
	
	// Read-Only (Admin, Manager, Supervisor)
	ro := admin.Group("/", middleware.RoleCheck("admin", "manager", "supervisor"))
	ro.Get("/centers", adminHandler.ListCenters)
	ro.Get("/centers/:id/details", adminHandler.GetCenterDetails)
	ro.Get("/shifts", adminHandler.ListShifts)
	ro.Get("/shifts/:id/details", adminHandler.GetShiftDetails)
	ro.Get("/positions", adminHandler.ListPositions)
	ro.Get("/positions/:id/details", adminHandler.GetPositionDetails)
	ro.Get("/employees", adminHandler.ListEmployees)
	ro.Get("/employees/:id/details", adminHandler.GetEmployeeDetails)
	ro.Get("/holidays", adminHandler.ListHolidays)
	ro.Get("/managers", adminHandler.ListManagers)
	ro.Get("/users/unassigned", adminHandler.ListUnassignedUsers)
	ro.Get("/attendances", adminHandler.ListAttendances)
	ro.Get("/attendances/:id/details", adminHandler.GetAttendanceDetails)
	ro.Get("/attendances/export/csv", adminHandler.ExportAttendancesCSV)
	ro.Get("/attendances/export/pdf", adminHandler.ExportAttendancesPDF)
	ro.Post("/attendances/:id/recalculate", adminHandler.RecalculateIncidents)
	ro.Delete("/attendances/:id", adminHandler.DeleteAttendance)
	ro.Get("/reports", reportHandler.ListReports)
	ro.Get("/reports/details", reportHandler.GetReportDetails)
	ro.Get("/reports/jobs", reportHandler.ListReportJobs)
	ro.Get("/reports/jobs/:id", reportHandler.GetReportJob)
	ro.Get("/alerts", adminHandler.ListAlerts)
	ro.Post("/alerts/:id/read", adminHandler.MarkAlertAsRead)
	ro.Get("/justifications", adminHandler.ListJustifications)
	ro.Post("/justifications/:id/resolve", adminHandler.ResolveJustification)
	ro.Get("/reports/:id/export", reportHandler.DownloadIndividualReport)
	ro.Get("/reports/export", reportHandler.DownloadBatchReport)
	ro.Get("/stats", adminHandler.GetDashboardStats)
	ro.Get("/dashboard/compliance", adminHandler.GetComplianceDashboard)
	ro.Get("/audit-logs", adminHandler.ListAuditLogs)
	ro.Get("/incidents", adminHandler.ListIncidents)

	// Range deletion for reports (Admin/Manager check depending on logic, but currently in ro if it's read-only? No, deletion is RW)

	// Write Actions (Admin Only)
	rw := admin.Group("/", middleware.RoleCheck("admin"))
	rw.Post("/centers", adminHandler.CreateCenter)
	rw.Put("/centers/:id", adminHandler.UpdateCenter)
	rw.Delete("/centers/:id", adminHandler.DeleteCenter)

	rw.Post("/shifts", adminHandler.CreateShift)
	rw.Put("/shifts/:id", adminHandler.UpdateShift)
	rw.Delete("/shifts/:id", adminHandler.DeleteShift)

	rw.Post("/positions", adminHandler.CreatePosition)
	rw.Put("/positions/:id", adminHandler.UpdatePosition)
	rw.Delete("/positions/:id", adminHandler.DeletePosition)

	rw.Post("/employees", adminHandler.CreateEmployee)
	rw.Put("/employees/:id", adminHandler.UpdateEmployee)
	rw.Delete("/employees/:id", adminHandler.DeleteEmployee)

	rw.Post("/holidays", adminHandler.CreateHoliday)
	rw.Put("/holidays/:id", adminHandler.UpdateHoliday)
	rw.Delete("/holidays/:id", adminHandler.DeleteHoliday)

	rw.Post("/reports/generate", reportHandler.GenerateFullReport)
	rw.Delete("/reports", reportHandler.DeleteReports)
	rw.Delete("/reports/:id", reportHandler.DeleteReportByID)
	rw.Post("/utils/parse-maps-url", utilsHandler.ParseMapsURL)
	rw.Get("/utils/detect-timezone", utilsHandler.DetectTimezone)

	rw.Put("/attendances/:id", adminHandler.UpdateAttendance)
	rw.Patch("/incidents/:id", adminHandler.UpdateIncidentStatus)

	// Bulk Actions
	bulk := admin.Group("/bulk", middleware.RoleCheck("admin"))
	bulk.Post("/employees/update", adminHandler.BulkUpdateEmployees)
	bulk.Post("/employees/delete", adminHandler.BulkDeleteEmployees)
	bulk.Post("/attendances/justify", adminHandler.BulkJustifyAttendances)
	bulk.Post("/incidents/justify", adminHandler.BulkJustifyIncidents)

	// Manager Endpoints (admin & manager)
	manager := api.Group("/manager", middleware.JWTAuth(cfg), middleware.RoleCheck("admin", "manager"))
	manager.Get("/centers", managerHandler.ListManagedCenters)
	manager.Get("/employees", managerHandler.ListManagedEmployees)
	manager.Post("/assign/:id", managerHandler.AssignEmployee)

	// User/Employee Endpoints
	user := api.Group("/user", middleware.JWTAuth(cfg))
	user.Get("/stats", userHandler.GetEmployeeStats)
	user.Get("/history", userHandler.GetAttendanceHistory)
	user.Get("/profile", userHandler.GetProfile)
	user.Put("/profile", userHandler.UpdateProfile)

	// Start Server
	log.Fatal(app.Listen(":" + cfg.Port))
}
