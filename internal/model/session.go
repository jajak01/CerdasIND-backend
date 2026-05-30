package model

import (
	"time"
)

type SessionStatus string

const (
	SessionScheduled SessionStatus = "scheduled"
	SessionCompleted SessionStatus = "completed"
	SessionCancelled SessionStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
	PaymentOverdue PaymentStatus = "overdue"
)

type Session struct {
	ID            int64         `json:"id" db:"id"`
	StudentID     int64         `json:"student_id" db:"student_id"`
	StudentName   string        `json:"student_name,omitempty" db:"student_name"` // Join result
	Date          time.Time     `json:"date" db:"date"`
	Time          string        `json:"time" db:"time"`
	Subject       string        `json:"subject" db:"subject"`
	Notes         string        `json:"notes" db:"notes"`
	Price         float64       `json:"price" db:"price"`
	Status        SessionStatus `json:"status" db:"status"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentDate   *time.Time    `json:"payment_date,omitempty" db:"payment_date"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type SessionRequest struct {
	StudentID     int64         `json:"student_id" binding:"required"`
	Date          string        `json:"date" binding:"required"` // Format: YYYY-MM-DD
	Time          string        `json:"time" binding:"required"` // Format: HH:MM
	Subject       string        `json:"subject"`
	Notes         string        `json:"notes"`
	Price         float64       `json:"price"`
	Status        SessionStatus `json:"status"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	PaymentDate   *string       `json:"payment_date"` // Format: YYYY-MM-DD HH:MM:SS
}
