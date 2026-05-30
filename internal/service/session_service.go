package service

import (
	"context"
	"time"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
)

type SessionService interface {
	CreateSession(ctx context.Context, req model.SessionRequest) error
	UpdateSession(ctx context.Context, id int64, req model.SessionRequest) error
	DeleteSession(ctx context.Context, id int64) error
	GetSessionByID(ctx context.Context, id int64) (*model.Session, error)
	GetAllSessions(ctx context.Context, filters map[string]interface{}) ([]model.Session, error)
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
}

type sessionService struct {
	sessionRepo repository.SessionRepository
}

func NewSessionService(sessionRepo repository.SessionRepository) SessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) CreateSession(ctx context.Context, req model.SessionRequest) error {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return err
	}

	if req.Subject == "" {
		req.Subject = "isikan mapel"
	}
	if req.Price == 0 {
		req.Price = 20000.00
	}
	if req.Status == "" {
		req.Status = model.SessionScheduled
	}
	if req.PaymentStatus == "" {
		req.PaymentStatus = model.PaymentPending
	}

	var paymentDate *time.Time
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		pd, err := time.Parse("2006-01-02 15:04:05", *req.PaymentDate)
		if err == nil {
			paymentDate = &pd
		}
	}

	session := &model.Session{
		StudentID:     req.StudentID,
		Date:          date,
		Time:          req.Time,
		Subject:       req.Subject,
		Notes:         req.Notes,
		Price:         req.Price,
		Status:        req.Status,
		PaymentStatus: req.PaymentStatus,
		PaymentDate:   paymentDate,
	}

	return s.sessionRepo.Create(ctx, session)
}

func (s *sessionService) UpdateSession(ctx context.Context, id int64, req model.SessionRequest) error {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return err
	}

	session.StudentID = req.StudentID
	session.Date = date
	session.Time = req.Time
	session.Subject = req.Subject
	session.Notes = req.Notes
	session.Price = req.Price
	session.Status = req.Status
	session.PaymentStatus = req.PaymentStatus

	// Logic for payment_date: clears it if status is not 'paid'
	if req.PaymentStatus == model.PaymentPaid {
		if req.PaymentDate != nil && *req.PaymentDate != "" {
			pd, err := time.Parse("2006-01-02 15:04:05", *req.PaymentDate)
			if err == nil {
				session.PaymentDate = &pd
			}
		} else if session.PaymentDate == nil {
			now := time.Now()
			session.PaymentDate = &now
		}
	} else {
		session.PaymentDate = nil
	}

	return s.sessionRepo.Update(ctx, session)
}

func (s *sessionService) DeleteSession(ctx context.Context, id int64) error {
	return s.sessionRepo.Delete(ctx, id)
}

func (s *sessionService) GetSessionByID(ctx context.Context, id int64) (*model.Session, error) {
	return s.sessionRepo.FindByID(ctx, id)
}

func (s *sessionService) GetAllSessions(ctx context.Context, filters map[string]interface{}) ([]model.Session, error) {
	// Process dates in filters
	if sd, ok := filters["start_date"]; ok && sd != "" {
		if t, err := time.Parse("2006-01-02", sd.(string)); err == nil {
			filters["start_date"] = t
		}
	}
	if ed, ok := filters["end_date"]; ok && ed != "" {
		if t, err := time.Parse("2006-01-02", ed.(string)); err == nil {
			filters["end_date"] = t
		}
	}

	return s.sessionRepo.FindAll(ctx, filters)
}

func (s *sessionService) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	now := time.Now()
	
	// Start of month
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// We'll use startOfMonth and endOfMonth for the main stats (revenue etc)
	return s.sessionRepo.GetStats(ctx, startOfMonth, endOfMonth)
}
