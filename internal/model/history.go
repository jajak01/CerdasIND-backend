package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type StatusUjian string

const (
	StatusBerlangsung     StatusUjian = "berlangsung"
	StatusMenungguKoreksi StatusUjian = "menunggu_koreksi"
	StatusSelesai         StatusUjian = "selesai"
)

type DetailJawaban struct {
	SoalID         int64  `json:"soal_id"`
	JawabanPeserta string `json:"jawaban_peserta"`
	SkorDidapat    float64 `json:"skor_didapat"`
	IsDinilai      bool   `json:"is_dinilai"`
}

type DetailJawabanList []DetailJawaban

func (l DetailJawabanList) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *DetailJawabanList) Scan(value interface{}) error {
	if value == nil {
		*l = DetailJawabanList{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, l)
}

type HistoryUjian struct {
	ID            int64             `json:"id" db:"id"`
	UserID        int64             `json:"user_id" db:"user_id"`
	BundleID      int64             `json:"bundle_id" db:"bundle_id"`
	WaktuMulai    time.Time         `json:"waktu_mulai" db:"waktu_mulai"`
	WaktuSelesai  *time.Time        `json:"waktu_selesai" db:"waktu_selesai"`
	SkorAkhir     float64           `json:"skor_akhir" db:"skor_akhir"`
	DetailJawaban DetailJawabanList `json:"detail_jawaban" db:"detail_jawaban"`
	Status        StatusUjian       `json:"status" db:"status"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
}

type SubmitJawaban struct {
	SoalID         int64  `json:"soal_id"`
	JawabanPeserta string `json:"jawaban_peserta"`
}

type SubmitRequest struct {
	Jawaban []SubmitJawaban `json:"jawaban"`
}

type HistoryResponse struct {
	HistoryID  int64     `json:"history_id"`
	NamaBundle string    `json:"nama_bundle"`
	WaktuMulai time.Time `json:"waktu_mulai"`
	SkorAkhir  float64   `json:"skor_akhir"`
	Status     StatusUjian `json:"status"`
}

type PenilaianManual struct {
	SoalID         int64   `json:"soal_id"`
	SkorDiberikan  float64 `json:"skor_diberikan"`
}

type GradeRequest struct {
	PenilaianManual []PenilaianManual `json:"penilaian_manual"`
}

type ReviewResponse struct {
	ID             int64              `json:"id"`
	TipeSoal       JenisSoal          `json:"tipe_soal"`
	TeksSoal       string             `json:"teks_soal"`
	PilihanJawaban PilihanJawabanList `json:"pilihan_jawaban"`
	Pembahasan     string             `json:"pembahasan"`
	JawabanPeserta string             `json:"jawaban_peserta"`
	KunciJawaban   string             `json:"kunci_jawaban"`
	IsBenar        bool               `json:"is_benar"`
}
