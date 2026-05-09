package handlers

import (
	"fmt"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	var stats struct {
		TotalEmployees int `db:"total_employees" json:"total_employees"`
		TotalCenters   int `db:"total_centers" json:"total_centers"`
	}

	// 1. Get Basic Counts
	err := h.DB.Get(&stats, "SELECT (SELECT COUNT(*) FROM employees WHERE is_active = true) as total_employees, (SELECT COUNT(*) FROM work_centers) as total_centers")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 2. Get Recent Incidents with Context
	type IncidentWithContext struct {
		ID           int       `db:"id" json:"id"`
		EmployeeName string    `db:"employee_name" json:"employee_name"`
		CenterName   string    `db:"center_name" json:"center_name"`
		Type         string    `db:"type" json:"type"`
		CreatedAt    time.Time `db:"created_at" json:"created_at"`
	}

	var incidents []IncidentWithContext
	incidentQuery := `
		SELECT i.id, u.name as employee_name, COALESCE(wc.name, 'Sede no identificada') as center_name, i.type, i.created_at
		FROM incidents i
		JOIN attendances a ON i.attendance_id = a.id
		JOIN employees e ON a.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		ORDER BY i.created_at DESC
		LIMIT 5
	`
	_ = h.DB.Select(&incidents, incidentQuery)

	// 3. System Alerts
	var alerts []models.SystemAlert
	h.DB.Select(&alerts, "SELECT * FROM system_alerts WHERE is_read = false ORDER BY created_at DESC LIMIT 10")

	// 4. Pending Justifications
	var justifications []models.Justification
	h.DB.Select(&justifications, "SELECT * FROM justifications WHERE status = 'pending' ORDER BY created_at DESC LIMIT 10")

	// 5. Compliance Trend (last 7 days)
	type TrendPoint struct {
		Date       string  `json:"date"`
		Attendance int     `json:"attendance"`
		Incidents  int     `json:"incidents"`
		Compliance float64 `json:"compliance"`
	}

	var trend []TrendPoint
	trendQuery := `
		SELECT 
			d::date::text as date,
			(SELECT COUNT(*) FROM attendances a WHERE a.check_in::date = d::date) as attendance,
			(SELECT COUNT(*) FROM incidents i WHERE i.created_at::date = d::date) as incidents
		FROM generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day'::interval) d
		ORDER BY d ASC
	`
	_ = h.DB.Select(&trend, trendQuery)

	// Calculate compliance for each point
	var totalComp float64
	for i, p := range trend {
		if p.Attendance > 0 {
			trend[i].Compliance = 100.0 - (float64(p.Incidents) / float64(p.Attendance) * 100.0)
			if trend[i].Compliance < 0 { trend[i].Compliance = 0 }
		} else {
			trend[i].Compliance = 100.0
		}
		totalComp += trend[i].Compliance
	}

	complianceRate := totalComp / 7.0

	return c.JSON(fiber.Map{
		"total_employees":   stats.TotalEmployees,
		"total_centers":     stats.TotalCenters,
		"recent_incidents":  incidents,
		"compliance_rate":   fmt.Sprintf("%.1f", complianceRate),
		"compliance_trend":  trend,
		"alerts":            alerts,
		"justifications":    justifications,
	})
}


func (h *DashboardHandler) GetComplianceDashboard(c *fiber.Ctx) error {
	type CenterCompliance struct {
		CenterID        int    `json:"center_id"`
		CenterName      string `json:"center_name"`
		TotalExpected   int    `json:"total_expected"`
		PresentCount    int    `json:"present_count"`
		LateCount       int    `json:"late_count"`
		OutOfRangeCount int    `json:"out_of_range_count"`
	}

	var stats []CenterCompliance
	query := `
		SELECT 
			wc.id as center_id, 
			wc.name as center_name,
			(SELECT COUNT(*) FROM employees e WHERE e.work_center_id = wc.id AND e.is_active = true) as total_expected,
			(SELECT COUNT(*) FROM attendances a WHERE a.work_center_id = wc.id AND a.check_in::date = CURRENT_DATE AND a.check_out IS NULL) as present_count,
			(SELECT COUNT(*) FROM incidents i WHERE i.work_center_id = wc.id AND i.created_at::date = CURRENT_DATE AND i.type = 'late') as late_count,
			(SELECT COUNT(*) FROM incidents i WHERE i.work_center_id = wc.id AND i.created_at::date = CURRENT_DATE AND i.type = 'out_of_range') as out_of_range_count
		FROM work_centers wc
		ORDER BY wc.name
	`
	
	if err := h.DB.Select(&stats, query); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(stats)
}

// RecalculateIncidents triggers an automatic scan for infractions on an existing record

