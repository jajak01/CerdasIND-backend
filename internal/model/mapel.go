package model

type Mapel struct {
	ID        int    `json:"id" db:"id"`
	JenjangID int    `json:"jenjang_id" db:"jenjang_id"`
	Nama      string `json:"nama" db:"nama"`
}
