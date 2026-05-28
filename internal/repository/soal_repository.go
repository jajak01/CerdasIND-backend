package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type SoalRepository interface {
	FindByBundleID(ctx context.Context, bundleID int64) ([]model.Soal, error)
	BulkCreate(ctx context.Context, soalList []model.Soal) error
	Update(ctx context.Context, soal *model.Soal) error
}

type soalRepository struct {
	db *sqlx.DB
}

func NewSoalRepository(db *sqlx.DB) SoalRepository {
	return &soalRepository{db: db}
}

func (r *soalRepository) FindByBundleID(ctx context.Context, bundleID int64) ([]model.Soal, error) {
	var soalList []model.Soal
	query := `SELECT * FROM soal WHERE bundle_id = $1`
	err := getRunner(ctx, r.db).SelectContext(ctx, &soalList, query, bundleID)
	return soalList, err
}

func (r *soalRepository) BulkCreate(ctx context.Context, soalList []model.Soal) error {
	if len(soalList) == 0 {
		return nil
	}

	query := `INSERT INTO soal (bundle_id, tipe_soal, teks_soal, pilihan_jawaban, kunci_jawaban, pembahasan, bobot_nilai) 
	          VALUES (:bundle_id, :tipe_soal, :teks_soal, :pilihan_jawaban, :kunci_jawaban, :pembahasan, :bobot_nilai)`
	_, err := getRunner(ctx, r.db).NamedExecContext(ctx, query, soalList)
	return err
}

func (r *soalRepository) Update(ctx context.Context, soal *model.Soal) error {
	query := `UPDATE soal SET tipe_soal = $1, teks_soal = $2, pilihan_jawaban = $3, kunci_jawaban = $4, pembahasan = $5, bobot_nilai = $6 WHERE id = $7`
	_, err := getRunner(ctx, r.db).ExecContext(ctx, query, soal.TipeSoal, soal.TeksSoal, soal.PilihanJawaban, soal.KunciJawaban, soal.Pembahasan, soal.BobotNilai, soal.ID)
	return err
}
