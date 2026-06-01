package repository

import (
	"context"
	"fmt"
	"strings"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/pkg/utils"
	"github.com/jmoiron/sqlx"
)

type DocumentRepository interface {
	Create(ctx context.Context, document *model.StudentDocument, sessions []model.StudentDocumentSession) error
	UpdateMessage(ctx context.Context, id int64, message string) error
	FindByID(ctx context.Context, id int64) (*model.StudentDocument, error)
	FindByPublicID(ctx context.Context, publicID string) (*model.StudentDocument, error)
	FindByIDWithSessions(ctx context.Context, id int64) (*model.StudentDocument, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error)
	FindSessionsByDocumentID(ctx context.Context, documentID int64) ([]model.StudentDocumentSession, error)
}

type documentRepository struct {
	db *sqlx.DB
}

func NewDocumentRepository(db *sqlx.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func documentPrefix(kind model.StudentDocumentKind) string {
	if kind == model.StudentDocumentReport {
		return "RPT"
	}
	return "INV"
}

func buildDocumentNumber(kind model.StudentDocumentKind, createdAt string, id int64) string {
	prefix := documentPrefix(kind)
	datePart := createdAt[:10]
	datePart = strings.ReplaceAll(datePart, "-", "")
	return fmt.Sprintf("%s-%s-%04d", prefix, datePart, id)
}

func (r *documentRepository) Create(ctx context.Context, document *model.StudentDocument, sessions []model.StudentDocumentSession) error {
	runner := getRunner(ctx, r.db)

	if document.PublicID == "" {
		publicID, err := utils.GenerateUUID()
		if err != nil {
			return err
		}
		document.PublicID = publicID
	}

	if document.DocumentKind == "" {
		document.DocumentKind = model.StudentDocumentInvoice
	}

	tempNumber := fmt.Sprintf("%s-TEMP-%s", documentPrefix(document.DocumentKind), document.PublicID[:8])
	query := `INSERT INTO student_documents (public_id, document_kind, document_number, student_id, linked_invoice_id, period_start, period_end, total_amount, summary, message, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	          RETURNING id, created_at, updated_at`
	err := runner.QueryRowContext(ctx, query,
		document.PublicID,
		document.DocumentKind,
		tempNumber,
		document.StudentID,
		document.LinkedInvoiceID,
		document.PeriodStart,
		document.PeriodEnd,
		document.TotalAmount,
		document.Summary,
		document.Message,
		document.CreatedBy,
	).Scan(&document.ID, &document.CreatedAt, &document.UpdatedAt)
	if err != nil {
		return err
	}

	document.DocumentNumber = buildDocumentNumber(document.DocumentKind, document.CreatedAt.Format("2006-01-02 15:04:05"), document.ID)
	_, err = runner.ExecContext(ctx, `UPDATE student_documents SET document_number = $1 WHERE id = $2`, document.DocumentNumber, document.ID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		_, err = runner.ExecContext(ctx, `
			INSERT INTO student_document_sessions (document_id, session_id, session_date, session_time, subject, note, price, payment_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, document.ID, session.SessionID, session.SessionDate, session.SessionTime, session.Subject, session.Note, session.Price, session.PaymentStatus)
		if err != nil {
			return err
		}
	}

	document.SessionCount = len(sessions)
	document.Sessions = sessions
	return nil
}

func (r *documentRepository) UpdateMessage(ctx context.Context, id int64, message string) error {
	_, err := getRunner(ctx, r.db).ExecContext(ctx, `UPDATE student_documents SET message = $1 WHERE id = $2`, message, id)
	return err
}

func (r *documentRepository) FindByID(ctx context.Context, id int64) (*model.StudentDocument, error) {
	var document model.StudentDocument
	query := `
		SELECT d.id, d.public_id, d.document_kind, d.document_number, d.student_id, st.name AS student_name,
		       d.linked_invoice_id, linked.document_number AS linked_invoice_number,
		       d.period_start, d.period_end, d.total_amount, d.summary, d.message, d.created_by,
		       d.created_at, d.updated_at,
		       COALESCE((SELECT COUNT(*) FROM student_document_sessions ds WHERE ds.document_id = d.id), 0) AS session_count
		FROM student_documents d
		JOIN students st ON d.student_id = st.id
		LEFT JOIN student_documents linked ON d.linked_invoice_id = linked.id
		WHERE d.id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &document, query, id)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) FindByPublicID(ctx context.Context, publicID string) (*model.StudentDocument, error) {
	var document model.StudentDocument
	query := `
		SELECT d.id, d.public_id, d.document_kind, d.document_number, d.student_id, st.name AS student_name,
		       d.linked_invoice_id, linked.document_number AS linked_invoice_number,
		       d.period_start, d.period_end, d.total_amount, d.summary, d.message, d.created_by,
		       d.created_at, d.updated_at,
		       COALESCE((SELECT COUNT(*) FROM student_document_sessions ds WHERE ds.document_id = d.id), 0) AS session_count
		FROM student_documents d
		JOIN students st ON d.student_id = st.id
		LEFT JOIN student_documents linked ON d.linked_invoice_id = linked.id
		WHERE d.public_id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &document, query, publicID)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) FindByIDWithSessions(ctx context.Context, id int64) (*model.StudentDocument, error) {
	document, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sessions, err := r.FindSessionsByDocumentID(ctx, id)
	if err != nil {
		return nil, err
	}

	document.Sessions = sessions
	return document, nil
}

func (r *documentRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]model.StudentDocument, error) {
	var documents []model.StudentDocument
	var args []interface{}
	var conditions []string
	argIdx := 1

	query := `
		SELECT d.id, d.public_id, d.document_kind, d.document_number, d.student_id, st.name AS student_name,
		       d.linked_invoice_id, linked.document_number AS linked_invoice_number,
		       d.period_start, d.period_end, d.total_amount, d.summary, d.message, d.created_by,
		       d.created_at, d.updated_at,
		       COALESCE((SELECT COUNT(*) FROM student_document_sessions ds WHERE ds.document_id = d.id), 0) AS session_count
		FROM student_documents d
		JOIN students st ON d.student_id = st.id
		LEFT JOIN student_documents linked ON d.linked_invoice_id = linked.id`

	if val, ok := filters["kind"]; ok && val != "" {
		conditions = append(conditions, fmt.Sprintf("d.document_kind = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := filters["student_id"]; ok && val != nil {
		conditions = append(conditions, fmt.Sprintf("d.student_id = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := filters["linked_invoice_id"]; ok && val != nil {
		conditions = append(conditions, fmt.Sprintf("d.linked_invoice_id = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY d.created_at DESC"

	err := getRunner(ctx, r.db).SelectContext(ctx, &documents, query, args...)
	return documents, err
}

func (r *documentRepository) FindSessionsByDocumentID(ctx context.Context, documentID int64) ([]model.StudentDocumentSession, error) {
	sessions := make([]model.StudentDocumentSession, 0)
	query := `
		SELECT id, document_id, session_id, session_date, session_time::text AS session_time, subject, note, price, payment_status, created_at
		FROM student_document_sessions
		WHERE document_id = $1
		ORDER BY session_date ASC, session_time ASC`
	err := getRunner(ctx, r.db).SelectContext(ctx, &sessions, query, documentID)
	return sessions, err
}
