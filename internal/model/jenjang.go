package model

type Jenjang struct {
	ID   int    `json:"id" db:"id"`
	Nama string `json:"nama" db:"nama"`
}
