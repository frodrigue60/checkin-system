package models

import (
	"encoding/json"
	"time"
)

// MapUserToDTO converts a User entity to a clean DTO
func MapUserToDTO(u User) UserDTO {
	createdAt := ""
	if u.CreatedAt != nil {
		createdAt = u.CreatedAt.Format(time.RFC3339)
	}
	return UserDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		RoleID:    u.RoleID,
		CreatedAt: createdAt,
	}
}

// MapEmployeeToDTO converts an Employee entity to DTO
func MapEmployeeToDTO(e Employee) EmployeeDTO {
	return EmployeeDTO{
		ID:           e.ID,
		UserID:       e.UserID,
		PositionID:   e.PositionID,
		WorkCenterID: e.WorkCenterID,
		WorkShiftID:  e.WorkShiftID,
		IsActive:     e.IsActive,
	}
}

// MapAttendanceToDTO converts Attendance entity to DTO
func MapAttendanceToDTO(a Attendance) AttendanceDTO {
	dateStr := ""
	if a.CheckIn != nil && !a.CheckIn.IsZero() {
		dateStr = a.CheckIn.Format("2006-01-02")
	} else if a.CreatedAt != nil {
		dateStr = a.CreatedAt.Format("2006-01-02")
	}
	netHours := 0.0
	if a.NetHoursWorked != nil {
		netHours = *a.NetHoursWorked
	}
	earnings := 0.0
	if a.DailyEarnings != nil {
		earnings = *a.DailyEarnings
	}

	var evs []string
	if len(a.EvidenceURLs) > 0 {
		_ = json.Unmarshal(a.EvidenceURLs, &evs)
	}
	if evs == nil {
		evs = []string{}
	}

	return AttendanceDTO{
		ID:                a.ID,
		EmployeeID:        a.EmployeeID,
		Date:              dateStr,
		CheckIn:           a.CheckIn,
		LunchStart:        a.LunchStart,
		LunchEnd:          a.LunchEnd,
		CheckOut:          a.CheckOut,
		NetHoursWorked:    netHours,
		DailyEarnings:     earnings,
		CheckInLatitude:   a.CheckInLatitude,
		CheckInLongitude:  a.CheckInLongitude,
		CheckOutLatitude:  a.CheckOutLatitude,
		CheckOutLongitude: a.CheckOutLongitude,
		IsAbsence:         a.IsAbsence,
		AbsenceReason:     a.AbsenceReason,
		EvidenceURL:       a.EvidenceURL,
		CheckOutNote:      a.CheckOutNote,
		CheckInAddress:    a.CheckInAddress,
		CheckInNote:       a.CheckInNote,
		CheckOutAddress:   a.CheckOutAddress,
		IsFieldWork:       a.IsFieldWork,
		EvidenceURLs:      evs,
	}
}

// MapAttendanceToDetailDTO is used for rich administrative logs
func MapAttendanceToDetailDTO(a Attendance, empName, centerName, posName string, isLate bool) AttendanceDetailDTO {
	dto := MapAttendanceToDTO(a)
	return AttendanceDetailDTO{
		ID:                dto.ID,
		EmployeeID:        dto.EmployeeID,
		Date:              dto.Date,
		CheckIn:           dto.CheckIn,
		LunchStart:        dto.LunchStart,
		LunchEnd:          dto.LunchEnd,
		CheckOut:          dto.CheckOut,
		NetHoursWorked:    dto.NetHoursWorked,
		DailyEarnings:     dto.DailyEarnings,
		CheckInLatitude:   a.CheckInLatitude,
		CheckInLongitude:  a.CheckInLongitude,
		CheckOutLatitude:  a.CheckOutLatitude,
		CheckOutLongitude: a.CheckOutLongitude,
		IsAbsence:         dto.IsAbsence,
		AbsenceReason:     dto.AbsenceReason,
		EmployeeName:      empName,
		WorkCenterName:    centerName,
		PositionName:      posName,
		IsLate:            isLate,
		EvidenceURL:       a.EvidenceURL,
		CheckOutNote:      a.CheckOutNote,
		CheckInAddress:    a.CheckInAddress,
		CheckInNote:       a.CheckInNote,
		CheckOutAddress:   a.CheckOutAddress,
		IsFieldWork:       a.IsFieldWork,
		EvidenceURLs:      dto.EvidenceURLs,
	}
}

// MapReportToDTO converts Report entity to DTO (Decodes JSON breakdown)
func MapReportToDTO(r Report) ReportDTO {
	var breakdown []DailyBreakdownItem
	if len(r.DailyBreakdown) > 0 {
		_ = json.Unmarshal(r.DailyBreakdown, &breakdown)
	}

	return ReportDTO{
		ID:               r.ID,
		EmployeeID:       r.EmployeeID,
		StartDate:        r.StartDate.Format("2006-01-02"),
		EndDate:          r.EndDate.Format("2006-01-02"),
		TotalHoursWorked: r.TotalHoursWorked,
		TotalEarnings:    r.TotalEarnings,
		TotalDeductions:  r.TotalDeductions,
		TotalIncidents:   r.TotalIncidents,
		DaysWorked:       r.DaysWorked,
		DailyBreakdown:   breakdown,
		EmployeeName:     r.EmployeeName,
		Status:           r.Status,
	}
}

// MapWorkCenterToDTO converts WorkCenter entity to DTO
func MapWorkCenterToDTO(wc WorkCenter) WorkCenterDTO {
	return WorkCenterDTO{
		ID:                    wc.ID,
		Name:                  wc.Name,
		Address:               wc.Address,
		Latitude:              wc.Latitude,
		Longitude:             wc.Longitude,
		ToleranceRadiusMeters: wc.ToleranceRadiusMeters,
	}
}

// MapPositionToDTO converts Position entity to DTO
func MapPositionToDTO(p Position) PositionDTO {
	return PositionDTO{
		ID:             p.ID,
		Name:           p.Name,
		HourlyRate:     p.HourlyRate,
		LatePenaltyFee: p.LatePenaltyFee,
		OutOfRangeFee:  p.OutOfRangeFee,
		LunchOverstayFee: p.LunchOverstayFee,
		EmployeesCount: p.EmployeesCount,
	}
}

// MapWorkShiftToDTO converts WorkShift entity to DTO
func MapWorkShiftToDTO(ws WorkShift) WorkShiftDTO {
	var days []int
	if len(ws.WorkDays) > 0 {
		_ = json.Unmarshal(ws.WorkDays, &days)
	}

	return WorkShiftDTO{
		ID:               ws.ID,
		Name:             ws.Name,
		ExpectedCheckIn:  ws.ExpectedCheckIn,
		ExpectedCheckOut: ws.ExpectedCheckOut,
		AllowedLunchTime: ws.AllowedLunchTime,
		ToleranceTime:    ws.ToleranceTime,
		IsNightShift:     ws.IsNightShift,
		IsActive:          ws.IsActive,
		EnforceLateness:   ws.EnforceLateness,
		EnforceLunchLimit: ws.EnforceLunchLimit,
		EnforceGeofence:   ws.EnforceGeofence,
		ShiftType:         ws.ShiftType,
		WorkDays:          days,
	}
}

// MapHolidayToDTO converts Holiday entity to DTO
func MapHolidayToDTO(h Holiday) HolidayDTO {
	desc := ""
	if h.Description != nil {
		desc = *h.Description
	}
	return HolidayDTO{
		ID:          h.ID,
		Name:        h.Name,
		Date:        h.Date.Format("2006-01-02"),
		Description: desc,
		Type:        h.Type,
	}
}

// MapIncidentToDTO converts Incident entity to DTO
func MapIncidentToDTO(i Incident) IncidentDTO {
	return IncidentDTO{
		ID:           i.ID,
		Type:         i.Type,
		IsLate:       i.IsLate,
		DelayMinutes: i.DelayMinutes,
		IsOutOfRange: i.IsOutOfRange,
		Distance:     i.Distance,
		Status:       i.Status,
		ResolvedBy:   i.ResolvedBy,
		ResolutionNote: i.ResolutionNote,
		MetadataJSON: i.MetadataJSON,
		CreatedAt:    i.CreatedAt,
	}
}

func MapIncidentToRichDTO(i Incident, empName string, attDate string, centerName string) IncidentRichDTO {
	dto := MapIncidentToDTO(i)
	return IncidentRichDTO{
		IncidentDTO:    dto,
		EmployeeName:   empName,
		AttendanceDate: attDate,
		AttendanceID:   i.AttendanceID,
		WorkCenterName: centerName,
	}
}

func MapReportJobToDTO(j ReportJob) ReportJobDTO {
	return ReportJobDTO{
		ID:               j.ID,
		Status:           j.Status,
		Progress:         j.Progress,
		ProcessedRecords: j.ProcessedRecords,
		TotalRecords:     j.TotalRecords,
		StartDate:        j.StartDate.Format("2006-01-02"),
		EndDate:          j.EndDate.Format("2006-01-02"),
		CreatedAt:        j.CreatedAt.Format(time.RFC3339),
	}
}
