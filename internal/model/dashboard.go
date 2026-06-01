package model

type DashboardStats struct {
	TotalStudents      int     `json:"total_students" db:"total_students"`
	TodaySessions      int     `json:"today_sessions" db:"today_sessions"`
	ThisWeekSessions   int     `json:"this_week_sessions" db:"this_week_sessions"`
	ThisMonthRevenue   float64 `json:"this_month_revenue" db:"this_month_revenue"`
	PendingPayments    float64 `json:"pending_payments" db:"pending_payments"`
	TotalOmzet         float64 `json:"total_omzet" db:"total_omzet"`
}
