package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JenisSoal string

const (
	SoalPG           JenisSoal = "pilihan_ganda"
	SoalIsianSingkat JenisSoal = "isian_singkat"
)

type PilihanJawaban struct {
	Opsi string `json:"opsi"`
	Teks string `json:"teks"`
}

type PilihanJawabanList []PilihanJawaban

func (l PilihanJawabanList) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *PilihanJawabanList) Scan(value interface{}) error {
	if value == nil {
		*l = PilihanJawabanList{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, l)
}

type Soal struct {
	ID             int64              `json:"id" db:"id"`
	BundleID       int64              `json:"bundle_id" db:"bundle_id"`
	TipeSoal       JenisSoal          `json:"tipe_soal" db:"tipe_soal"`
	TeksSoal       string             `json:"teks_soal" db:"teks_soal"`
	PilihanJawaban PilihanJawabanList `json:"pilihan_jawaban" db:"pilihan_jawaban"`
	KunciJawaban   string             `json:"kunci_jawaban,omitempty" db:"kunci_jawaban"`
	Pembahasan     string             `json:"pembahasan,omitempty" db:"pembahasan"`
	BobotNilai     int                `json:"bobot_nilai" db:"bobot_nilai"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
}

type SoalPublic struct {
	ID             int64              `json:"id"`
	TipeSoal       JenisSoal          `json:"tipe_soal"`
	TeksSoal       string             `json:"teks_soal"`
	PilihanJawaban PilihanJawabanList `json:"pilihan_jawaban"`
	BobotNilai     int                `json:"bobot_nilai"`
}
