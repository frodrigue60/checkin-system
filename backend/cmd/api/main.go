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

	docs "attendance-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/swagger"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
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

	// Initialize Storage Service (R2)
	var storageService services.StorageService
	if cfg.R2AccountID != "" {
		r2, err := services.NewR2StorageService(cfg)
		if err != nil {
			utils.Logger.Error("Failed to initialize R2 storage service", zap.Error(err))
		} else {
			storageService = r2
			utils.Logger.Info("R2 Storage Service initialized successfully")
		}
	}

	uploadHandler := &handlers.UploadHandler{Storage: storageService}



	// Domain-specific handlers — SOLID (S): each handles one entity type
	base := handlers.AdminBase{
		DB:                   db,
		Cfg:                  cfg,
		PDFService:           pdfService,
		AttendanceService:    attendanceService,
		AuditService:         auditService,
		ReportService:        reportService,
		AlertService:         alertService,
		JustificationService: justificationService,
		Cache:                c,
	}
	centerHandler := &handlers.CenterHandler{AdminBase: base}
	shiftHandler := &handlers.ShiftHandler{AdminBase: base}
	positionHandler := &handlers.PositionHandler{AdminBase: base}
	employeeHandler := &handlers.EmployeeAdminHandler{AdminBase: base}
	holidayHandler := &handlers.HolidayHandler{AdminBase: base}
	attAdminHandler := &handlers.AttendanceAdminHandler{AdminBase: base}
	incidentHandler := &handlers.IncidentHandler{AdminBase: base}
	dashboardHandler := &handlers.DashboardHandler{AdminBase: base}
	exportHandler := &handlers.ExportHandler{AdminBase: base}
	bulkHandler := &handlers.BulkHandler{AdminBase: base}
	miscHandler := &handlers.MiscAdminHandler{AdminBase: base}

	attendanceHandler := &handlers.AttendanceHandler{
		DB:                   db,
		Cfg:                  cfg,
		Service:              attendanceService,
		JustificationService: justificationService,
	}
	excelService := services.NewExcelService()
	managerHandler := &handlers.ManagerHandler{DB: db}
	reportHandler := &handlers.ReportHandler{
		DB:           db,
		Cfg:          cfg,
		PDFService:   pdfService,
		ExcelService: excelService,
		Storage:      storageService,
		AuditService: auditService,
	}
	userHandler := &handlers.UserHandler{DB: db, Cfg: cfg, Storage: storageService}
	utilsHandler := &handlers.UtilsHandler{}

	// 3.1 Initialize Background Workers
	workerService := &services.WorkerService{
		DB:                db, 
		AttendanceService: attendanceService,
		AlertService:      alertService,
	}
	workerService.StartGhostSessionCleaner()

	// 4. Initialize Fiber App
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if len(c.Response().Body()) > 0 {
				return nil
			}
			return fiber.DefaultErrorHandler(c, err)
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Swagger
	if cfg.EnableSwagger {
		docs.SwaggerInfo.Host = cfg.SwaggerHost
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// API Routes
	api := app.Group("/api/v1")

	// Health Check
	api.Get("/health", func(c *fiber.Ctx) error {
		dbStatus := "connected"
		if err := db.Ping(); err != nil {
			dbStatus = "disconnected"
		}
		return c.JSON(fiber.Map{
			"status":   "up",
			"database": dbStatus,
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
	// Registration disabled — user creation is admin-only via POST /api/v1/admin/users

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
	attendance.Post("/upload", uploadHandler.UploadAttendanceEvidence)


	// Admin Root Group
	admin := api.Group("/admin", middleware.JWTAuth(cfg))
	
	// Read-Only (Admin, Manager, Supervisor)
	ro := admin.Group("/", middleware.RoleCheck("admin", "manager", "supervisor"))
	ro.Get("/centers", centerHandler.ListCenters)
	ro.Get("/centers/:id/details", centerHandler.GetCenterDetails)
	ro.Get("/shifts", shiftHandler.ListShifts)
	ro.Get("/shifts/:id/details", shiftHandler.GetShiftDetails)
	ro.Get("/positions", positionHandler.ListPositions)
	ro.Get("/positions/:id/details", positionHandler.GetPositionDetails)
	ro.Get("/employees", employeeHandler.ListEmployees)
	ro.Get("/employees/:id/details", employeeHandler.GetEmployeeDetails)
	ro.Get("/holidays", holidayHandler.ListHolidays)
	ro.Get("/managers", employeeHandler.ListManagers)
	ro.Get("/users/unassigned", employeeHandler.ListUnassignedUsers)
	ro.Get("/attendances", attAdminHandler.ListAttendances)
	ro.Get("/attendances/:id/details", attAdminHandler.GetAttendanceDetails)
	ro.Get("/attendances/export/csv", exportHandler.ExportAttendancesCSV)
	ro.Get("/attendances/export/pdf", exportHandler.ExportAttendancesPDF)
	ro.Post("/attendances/:id/recalculate", attAdminHandler.RecalculateIncidents)
	ro.Delete("/attendances/:id", attAdminHandler.DeleteAttendance)
	ro.Get("/reports", reportHandler.ListReports)
	ro.Get("/reports/details", reportHandler.GetReportDetails)
	ro.Get("/reports/jobs", reportHandler.ListReportJobs)
	ro.Get("/reports/jobs/:id", reportHandler.GetReportJob)
	ro.Get("/alerts", miscHandler.ListAlerts)
	ro.Post("/alerts/:id/read", miscHandler.MarkAlertAsRead)
	ro.Get("/justifications", miscHandler.ListJustifications)
	ro.Post("/justifications/:id/resolve", miscHandler.ResolveJustification)
	ro.Get("/reports/:id/export", reportHandler.DownloadIndividualReport)
	ro.Get("/reports/export", reportHandler.DownloadBatchReport)
	ro.Get("/stats", dashboardHandler.GetDashboardStats)
	ro.Get("/dashboard/compliance", dashboardHandler.GetComplianceDashboard)
	ro.Get("/audit-logs", miscHandler.ListAuditLogs)
	ro.Get("/incidents", incidentHandler.ListIncidents)

	// Range deletion for reports (Admin/Manager check depending on logic, but currently in ro if it's read-only? No, deletion is RW)

	// Write Actions (Admin Only)
	rw := admin.Group("/", middleware.RoleCheck("admin"))
	rw.Post("/centers", centerHandler.CreateCenter)
	rw.Put("/centers/:id", centerHandler.UpdateCenter)
	rw.Delete("/centers/:id", centerHandler.DeleteCenter)

	rw.Post("/shifts", shiftHandler.CreateShift)
	rw.Put("/shifts/:id", shiftHandler.UpdateShift)
	rw.Delete("/shifts/:id", shiftHandler.DeleteShift)

	rw.Post("/positions", positionHandler.CreatePosition)
	rw.Put("/positions/:id", positionHandler.UpdatePosition)
	rw.Delete("/positions/:id", positionHandler.DeletePosition)

	rw.Post("/employees", employeeHandler.CreateEmployee)
	rw.Put("/employees/:id", employeeHandler.UpdateEmployee)
	rw.Delete("/employees/:id", employeeHandler.DeleteEmployee)

	rw.Post("/holidays", holidayHandler.CreateHoliday)
	rw.Put("/holidays/:id", holidayHandler.UpdateHoliday)
	rw.Delete("/holidays/:id", holidayHandler.DeleteHoliday)

	rw.Post("/reports/generate", reportHandler.GenerateFullReport)
	rw.Delete("/reports", reportHandler.DeleteReports)
	rw.Delete("/reports/:id", reportHandler.DeleteReportByID)
	rw.Post("/utils/parse-maps-url", utilsHandler.ParseMapsURL)
	rw.Get("/utils/detect-timezone", utilsHandler.DetectTimezone)

	rw.Put("/attendances/:id", attAdminHandler.UpdateAttendance)
	rw.Patch("/incidents/:id", incidentHandler.UpdateIncidentStatus)

	// Bulk Actions
	bulk := admin.Group("/bulk", middleware.RoleCheck("admin"))
	bulk.Post("/employees/update", bulkHandler.BulkUpdateEmployees)
	bulk.Post("/employees/delete", bulkHandler.BulkDeleteEmployees)
	bulk.Post("/attendances/justify", bulkHandler.BulkJustifyAttendances)
	bulk.Post("/incidents/justify", bulkHandler.BulkJustifyIncidents)

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
	user.Get("/avatar/upload-url", userHandler.GetAvatarUploadURL)
	user.Post("/avatar/confirm", userHandler.ConfirmAvatarUpdate)

	// Start Server
	log.Fatal(app.Listen(":" + cfg.Port))
}
 
 
 
 
