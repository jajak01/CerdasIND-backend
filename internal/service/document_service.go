package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

type DocumentService interface {
	CreateInvoice(ctx context.Context, userID int64, req model.StudentDocumentRequest) (*model.StudentDocument, error)
	CreateReport(ctx context.Context, userID int64, req model.StudentDocumentRequest) (*model.StudentDocument, error)
	GetInvoices(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error)
	GetReports(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error)
	GetInvoiceByID(ctx context.Context, id int64) (*model.StudentDocument, error)
	GetReportByID(ctx context.Context, id int64) (*model.StudentDocument, error)
}

type documentService struct {
	db           *sqlx.DB
	documentRepo repository.DocumentRepository
	sessionRepo  repository.SessionRepository
	studentRepo  repository.StudentRepository
}

func NewDocumentService(db *sqlx.DB, documentRepo repository.DocumentRepository, sessionRepo repository.SessionRepository, studentRepo repository.StudentRepository) DocumentService {
	return &documentService{
		db:           db,
		documentRepo: documentRepo,
		sessionRepo:  sessionRepo,
		studentRepo:  studentRepo,
	}
}

func validateSessionOwnership(studentID int64, sessions []model.Session) error {
	for _, session := range sessions {
		if session.StudentID != studentID {
			return fmt.Errorf("sesi %d tidak sesuai dengan siswa yang dipilih", session.ID)
		}
	}
	return nil
}

func sortSessionsByDateTime(sessions []model.Session) {
	sort.SliceStable(sessions, func(i, j int) bool {
		leftDate, _ := time.Parse("2006-01-02", sessions[i].Date.Format("2006-01-02"))
		rightDate, _ := time.Parse("2006-01-02", sessions[j].Date.Format("2006-01-02"))
		if !leftDate.Equal(rightDate) {
			return leftDate.Before(rightDate)
		}
		return sessions[i].Time < sessions[j].Time
	})
}

func toSnapshotSessions(sessions []model.Session, includeNotes bool) []model.StudentDocumentSession {
	snapshots := make([]model.StudentDocumentSession, 0, len(sessions))
	for _, session := range sessions {
		note := ""
		if includeNotes {
			note = session.Notes
		}
		snapshots = append(snapshots, model.StudentDocumentSession{
			SessionID:     &session.ID,
			SessionDate:   session.Date,
			SessionTime:   session.Time,
			Subject:       session.Subject,
			Note:          note,
			Price:         session.Price,
			PaymentStatus: string(session.PaymentStatus),
		})
	}
	return snapshots
}

func buildDateRange(sessions []model.Session) (time.Time, time.Time) {
	if len(sessions) == 0 {
		now := time.Now()
		return now, now
	}

	start := sessions[0].Date
	end := sessions[0].Date
	for _, session := range sessions[1:] {
		if session.Date.Before(start) {
			start = session.Date
		}
		if session.Date.After(end) {
			end = session.Date
		}
	}
	return start, end
}

func buildDateRangeFromSnapshots(sessions []model.StudentDocumentSession) (time.Time, time.Time) {
	if len(sessions) == 0 {
		now := time.Now()
		return now, now
	}

	start := sessions[0].SessionDate
	end := sessions[0].SessionDate
	for _, session := range sessions[1:] {
		if session.SessionDate.Before(start) {
			start = session.SessionDate
		}
		if session.SessionDate.After(end) {
			end = session.SessionDate
		}
	}
	return start, end
}

func formatCurrencyIDR(value float64) string {
	return fmt.Sprintf("Rp%.0f", value)
}

func buildInvoiceMessage(studentName, documentNumber string, sessions []model.Session, totalAmount float64) string {
	if len(sessions) == 0 {
		return fmt.Sprintf("Nomor invoice: %s. Ini adalah invoice pembayaran untuk %s.", documentNumber, studentName)
	}

	firstDate := sessions[0].Date.Format("02 January 2006")
	lastDate := sessions[len(sessions)-1].Date.Format("02 January 2006")
	return strings.Join([]string{
		fmt.Sprintf("Nomor invoice: %s.", documentNumber),
		fmt.Sprintf("Ini adalah invoice pembayaran dari tanggal %s sampai %s.", firstDate, lastDate),
		fmt.Sprintf("Nama siswa: %s.", studentName),
		fmt.Sprintf("Total tagihan: %s.", formatCurrencyIDR(totalAmount)),
		"Mohon cek file invoice terlampir.",
	}, " ")
}

func buildReportMessage(studentName, documentNumber, summary string) string {
	parts := []string{
		fmt.Sprintf("Nomor report: %s.", documentNumber),
		fmt.Sprintf("Laporan perkembangan untuk %s.", studentName),
	}
	if summary != "" {
		parts = append(parts, fmt.Sprintf("Resume: %s.", summary))
	}
	parts = append(parts, "Mohon cek file report terlampir.")
	return strings.Join(parts, " ")
}

func (s *documentService) createDocument(ctx context.Context, kind model.StudentDocumentKind, userID int64, req model.StudentDocumentRequest) (*model.StudentDocument, error) {
	student, err := s.studentRepo.FindByID(ctx, req.StudentID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.New("siswa tidak ditemukan")
	}

	var sessions []model.Session
	var linkedInvoice *model.StudentDocument

	switch kind {
	case model.StudentDocumentInvoice:
		sessions, err = s.sessionRepo.FindByIDs(ctx, req.SessionIDs)
		if err != nil {
			return nil, err
		}
		if len(sessions) == 0 {
			return nil, errors.New("minimal satu sesi harus dipilih")
		}
		if err := validateSessionOwnership(student.ID, sessions); err != nil {
			return nil, err
		}
		for _, session := range sessions {
			if session.PaymentStatus != model.PaymentPaid {
				return nil, fmt.Errorf("sesi %d belum lunas", session.ID)
			}
		}
	case model.StudentDocumentReport:
		if req.LinkedInvoiceID == nil {
			return nil, errors.New("report harus terhubung ke invoice")
		}
		linkedInvoice, err = s.documentRepo.FindByIDWithSessions(ctx, *req.LinkedInvoiceID)
		if err != nil {
			return nil, err
		}
		if linkedInvoice.DocumentKind != model.StudentDocumentInvoice {
			return nil, errors.New("report hanya dapat dibuat dari invoice")
		}
		if linkedInvoice.StudentID != student.ID {
			return nil, errors.New("invoice tidak sesuai dengan siswa yang dipilih")
		}
		if len(linkedInvoice.Sessions) == 0 {
			return nil, errors.New("invoice belum memiliki sesi")
		}
		if len(req.SessionIDs) > 0 {
			expected := make(map[int64]struct{}, len(linkedInvoice.Sessions))
			for _, snapshot := range linkedInvoice.Sessions {
				if snapshot.SessionID != nil {
					expected[*snapshot.SessionID] = struct{}{}
				}
			}
			for _, id := range req.SessionIDs {
				if _, ok := expected[id]; !ok {
					return nil, fmt.Errorf("sesi %d tidak ada di invoice", id)
				}
			}
		}
	default:
		return nil, errors.New("jenis dokumen tidak valid")
	}

	var snapshotSessions []model.StudentDocumentSession
	var totalAmount float64
	var startDate, endDate time.Time

	switch kind {
	case model.StudentDocumentInvoice:
		sortSessionsByDateTime(sessions)
		startDate, endDate = buildDateRange(sessions)
		snapshotSessions = toSnapshotSessions(sessions, false)
		for _, session := range sessions {
			totalAmount += session.Price
		}
	case model.StudentDocumentReport:
		snapshotSessions = linkedInvoice.Sessions
		startDate, endDate = buildDateRangeFromSnapshots(snapshotSessions)
	}

	document := &model.StudentDocument{
		DocumentKind:    kind,
		StudentID:       student.ID,
		PeriodStart:     startDate,
		PeriodEnd:       endDate,
		TotalAmount:     totalAmount,
		Summary:         strings.TrimSpace(req.Summary),
		Message:         strings.TrimSpace(req.Message),
		CreatedBy:       &userID,
		LinkedInvoiceID: req.LinkedInvoiceID,
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txCtx := repository.InjectTx(ctx, tx)
	if err := s.documentRepo.Create(txCtx, document, snapshotSessions); err != nil {
		return nil, err
	}

	message := ""
	switch kind {
	case model.StudentDocumentInvoice:
		message = buildInvoiceMessage(student.Name, document.DocumentNumber, sessions, totalAmount)
	case model.StudentDocumentReport:
		message = buildReportMessage(student.Name, document.DocumentNumber, document.Summary)
	}
	if err := s.documentRepo.UpdateMessage(txCtx, document.ID, message); err != nil {
		return nil, err
	}
	document.Message = message

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return document, nil
}

func (s *documentService) CreateInvoice(ctx context.Context, userID int64, req model.StudentDocumentRequest) (*model.StudentDocument, error) {
	return s.createDocument(ctx, model.StudentDocumentInvoice, userID, req)
}

func (s *documentService) CreateReport(ctx context.Context, userID int64, req model.StudentDocumentRequest) (*model.StudentDocument, error) {
	return s.createDocument(ctx, model.StudentDocumentReport, userID, req)
}

func (s *documentService) GetInvoices(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error) {
	if filters == nil {
		filters = map[string]interface{}{}
	}
	filters["kind"] = model.StudentDocumentInvoice
	return s.documentRepo.FindAll(ctx, filters)
}

func (s *documentService) GetReports(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error) {
	if filters == nil {
		filters = map[string]interface{}{}
	}
	filters["kind"] = model.StudentDocumentReport
	return s.documentRepo.FindAll(ctx, filters)
}

func (s *documentService) GetInvoiceByID(ctx context.Context, id int64) (*model.StudentDocument, error) {
	return s.documentRepo.FindByIDWithSessions(ctx, id)
}

func (s *documentService) GetReportByID(ctx context.Context, id int64) (*model.StudentDocument, error) {
	return s.documentRepo.FindByIDWithSessions(ctx, id)
}
