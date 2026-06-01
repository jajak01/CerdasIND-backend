package model

import "time"

type Bundle struct {
	ID          int64     `json:"id" db:"id"`
	PublicID    string    `json:"public_id" db:"public_id"`
	MapelID     int       `json:"mapel_id" db:"mapel_id"`
	NamaBundle  string    `json:"nama_bundle" db:"nama_bundle"`
	Deskripsi   string    `json:"deskripsi,omitempty" db:"deskripsi"`
	WaktuMenit  int       `json:"waktu_menit" db:"waktu_menit"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   *int64    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
