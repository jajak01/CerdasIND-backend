package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/pkg/utils"
	"github.com/jmoiron/sqlx"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	Update(ctx context.Context, session *model.Session) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*model.Session, error)
	FindByPublicID(ctx context.Context, publicID string) (*model.Session, error)
	FindByIDs(ctx context.Context, ids []int64) ([]model.Session, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]model.Session, error)
	GetStats(ctx context.Context, startDate, endDate time.Time) (*model.DashboardStats, error)
}

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	if session.PublicID == "" {
		publicID, err := utils.GenerateUUID()
		if err != nil {
			return err
		}
		session.PublicID = publicID
	}

	query := `INSERT INTO sessions (public_id, student_id, date, time, subject, notes, price, status, payment_status, payment_date) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query,
		session.PublicID, session.StudentID, session.Date, session.Time, session.Subject, session.Notes, session.Price, session.Status, session.PaymentStatus, session.PaymentDate).
		Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
}

func (r *sessionRepository) Update(ctx context.Context, session *model.Session) error {
	query := `UPDATE sessions SET student_id = $1, date = $2, time = $3, subject = $4, notes = $5, price = $6, status = $7, payment_status = $8, payment_date = $9, updated_at = CURRENT_TIMESTAMP 
              WHERE id = $10 RETURNING updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query,
		session.StudentID, session.Date, session.Time, session.Subject, session.Notes, session.Price, session.Status, session.PaymentStatus, session.PaymentDate, session.ID).
		Scan(&session.UpdatedAt)
}

func (r *sessionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := getRunner(ctx, r.db).ExecContext(ctx, query, id)
	return err
}

func (r *sessionRepository) FindByID(ctx context.Context, id int64) (*model.Session, error) {
	var session model.Session
	query := `SELECT s.id, s.public_id, s.student_id, st.name as student_name, s.date, s.time::text as time, s.subject, s.notes, s.price, s.status, s.payment_status, s.payment_date, s.created_at, s.updated_at FROM sessions s 
              JOIN students st ON s.student_id = st.id 
              WHERE s.id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &session, query, id)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindByPublicID(ctx context.Context, publicID string) (*model.Session, error) {
	var session model.Session
	query := `SELECT s.id, s.public_id, s.student_id, st.name as student_name, s.date, s.time::text as time, s.subject, s.notes, s.price, s.status, s.payment_status, s.payment_date, s.created_at, s.updated_at FROM sessions s 
              JOIN students st ON s.student_id = st.id 
              WHERE s.public_id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &session, query, publicID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindByIDs(ctx context.Context, ids []int64) ([]model.Session, error) {
	sessions := make([]model.Session, 0)
	if len(ids) == 0 {
		return sessions, nil
	}

	query := `SELECT s.id, s.public_id, s.student_id, st.name as student_name, s.date, s.time::text as time, s.subject, s.notes, s.price, s.status, s.payment_status, s.payment_date, s.created_at, s.updated_at FROM sessions s 
              JOIN students st ON s.student_id = st.id 
              WHERE s.id IN (?) ORDER BY s.date ASC, s.time ASC`

	inQuery, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}

	inQuery = sqlx.Rebind(sqlx.DOLLAR, inQuery)
	err = getRunner(ctx, r.db).SelectContext(ctx, &sessions, inQuery, args...)
	return sessions, err
}

func (r *sessionRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]model.Session, error) {
	var sessions []model.Session
	var args []interface{}
	var conditions []string
	argIdx := 1

	query := `SELECT s.id, s.public_id, s.student_id, st.name as student_name, s.date, s.time::text as time, s.subject, s.notes, s.price, s.status, s.payment_status, s.payment_date, s.created_at, s.updated_at FROM sessions s 
              JOIN students st ON s.student_id = st.id`

	if val, ok := filters["student_id"]; ok && val != nil {
		conditions = append(conditions, fmt.Sprintf("s.student_id = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["start_date"]; ok && val != nil {
		conditions = append(conditions, fmt.Sprintf("s.date >= $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["end_date"]; ok && val != nil {
		conditions = append(conditions, fmt.Sprintf("s.date <= $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["status"]; ok && val != "" {
		conditions = append(conditions, fmt.Sprintf("s.status = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["payment_status"]; ok && val != "" {
		conditions = append(conditions, fmt.Sprintf("s.payment_status = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["search"]; ok && val != "" {
		conditions = append(conditions, fmt.Sprintf("(st.name ILIKE $%d OR st.contact ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+val.(string)+"%")
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY s.date DESC, s.time DESC"

	err := getRunner(ctx, r.db).SelectContext(ctx, &sessions, query, args...)
	return sessions, err
}

func (r *sessionRepository) GetStats(ctx context.Context, startDate, endDate time.Time) (*model.DashboardStats, error) {
	var stats model.DashboardStats

	// Total Students
	err := getRunner(ctx, r.db).GetContext(ctx, &stats.TotalStudents, "SELECT COUNT(*) FROM students")
	if err != nil {
		return nil, err
	}

	// Today Sessions
	today := time.Now().Format("2006-01-02")
	err = getRunner(ctx, r.db).GetContext(ctx, &stats.TodaySessions, "SELECT COUNT(*) FROM sessions WHERE date = $1", today)
	if err != nil {
		return nil, err
	}

	// This Week Sessions (non-cancelled)
	// We use the provided startDate and endDate which should be for the week or month as per user request
	err = getRunner(ctx, r.db).GetContext(ctx, &stats.ThisWeekSessions, 
		"SELECT COUNT(*) FROM sessions WHERE date BETWEEN $1 AND $2 AND status != 'cancelled'", startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Revenue & Omzet
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN payment_status = 'paid' THEN price ELSE 0 END), 0) as this_month_revenue,
			COALESCE(SUM(CASE WHEN payment_status = 'pending' THEN price ELSE 0 END), 0) as pending_payments,
			COALESCE(SUM(price), 0) as total_omzet
		FROM sessions 
		WHERE date BETWEEN $1 AND $2 AND status != 'cancelled'`
	
	err = getRunner(ctx, r.db).GetContext(ctx, &stats, query, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
