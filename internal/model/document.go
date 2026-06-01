package model

import "time"

type StudentDocumentKind string

const (
	StudentDocumentInvoice StudentDocumentKind = "invoice"
	StudentDocumentReport  StudentDocumentKind = "report"
)

type StudentDocument struct {
	ID                 int64                 `json:"id" db:"id"`
	PublicID           string                `json:"public_id" db:"public_id"`
	DocumentKind       StudentDocumentKind   `json:"document_kind" db:"document_kind"`
	DocumentNumber     string                `json:"document_number" db:"document_number"`
	StudentID          int64                 `json:"student_id" db:"student_id"`
	StudentName        string                `json:"student_name,omitempty" db:"student_name"`
	LinkedInvoiceID    *int64                `json:"linked_invoice_id,omitempty" db:"linked_invoice_id"`
	LinkedInvoiceNumber *string              `json:"linked_invoice_number,omitempty" db:"linked_invoice_number"`
	PeriodStart        time.Time             `json:"period_start" db:"period_start"`
	PeriodEnd          time.Time             `json:"period_end" db:"period_end"`
	TotalAmount        float64               `json:"total_amount" db:"total_amount"`
	Summary            string                `json:"summary,omitempty" db:"summary"`
	Message            string                `json:"message,omitempty" db:"message"`
	CreatedBy          *int64                `json:"created_by,omitempty" db:"created_by"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at" db:"updated_at"`
	SessionCount       int                   `json:"session_count" db:"session_count"`
	Sessions           []StudentDocumentSession `json:"sessions,omitempty"`
}

type StudentDocumentSession struct {
	ID           int64     `json:"id" db:"id"`
	DocumentID   int64     `json:"document_id" db:"document_id"`
	SessionID    *int64    `json:"session_id,omitempty" db:"session_id"`
	SessionDate  time.Time `json:"session_date" db:"session_date"`
	SessionTime  string    `json:"session_time" db:"session_time"`
	Subject      string    `json:"subject" db:"subject"`
	Note         string    `json:"note,omitempty" db:"note"`
	Price        float64   `json:"price" db:"price"`
	PaymentStatus string   `json:"payment_status" db:"payment_status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type StudentDocumentRequest struct {
	StudentID       int64   `json:"student_id" binding:"required"`
	LinkedInvoiceID *int64  `json:"linked_invoice_id"`
	SessionIDs      []int64 `json:"session_ids" binding:"required"`
	Summary         string  `json:"summary"`
	Message         string  `json:"message"`
}
