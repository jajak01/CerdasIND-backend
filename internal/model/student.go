package model

import "time"

type Student struct {
	ID        int64     `json:"id" db:"id"`
	PublicID  string    `json:"public_id" db:"public_id"`
	UserID    *int64    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	School    string    `json:"school" db:"school"`
	Grade     string    `json:"grade" db:"grade"`
	Contact   string    `json:"contact" db:"contact"`
	Address   string    `json:"address" db:"address"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type StudentRequest struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name" binding:"required"`
	School   string `json:"school"`
	Grade    string `json:"grade"`
	Contact  string `json:"contact"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}
