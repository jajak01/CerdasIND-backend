package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type HistoryRepository interface {
	Create(ctx context.Context, history *model.HistoryUjian) error
	Update(ctx context.Context, history *model.HistoryUjian) error
	FindByID(ctx context.Context, id int64) (*model.HistoryUjian, error)
	FindByUserID(ctx context.Context, userID int64) ([]model.HistoryUjian, error)
	FindByStatus(ctx context.Context, status model.StatusUjian) ([]model.HistoryUjian, error)
	FindUserHistoryByBundleID(ctx context.Context, userID int64, bundleID int64) (*model.HistoryUjian, error)
	FindOngoingHistoryByBundleID(ctx context.Context, userID int64, bundleID int64) (*model.HistoryUjian, error)
}

type historyRepository struct {
	db *sqlx.DB
}

func NewHistoryRepository(db *sqlx.DB) HistoryRepository {
	return &historyRepository{db: db}
}

func (r *historyRepository) Create(ctx context.Context, history *model.HistoryUjian) error {
	query := `INSERT INTO history_ujian (user_id, bundle_id, status) VALUES ($1, $2, $3) RETURNING id, waktu_mulai, created_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, history.UserID, history.BundleID, history.Status).Scan(&history.ID, &history.WaktuMulai, &history.CreatedAt)
}

func (r *historyRepository) Update(ctx context.Context, history *model.HistoryUjian) error {
	query := `UPDATE history_ujian SET waktu_selesai = $1, skor_akhir = $2, detail_jawaban = $3, status = $4 WHERE id = $5`
	_, err := getRunner(ctx, r.db).ExecContext(ctx, query, history.WaktuSelesai, history.SkorAkhir, history.DetailJawaban, history.Status, history.ID)
	return err
}

func (r *historyRepository) FindByID(ctx context.Context, id int64) (*model.HistoryUjian, error) {
	var history model.HistoryUjian
	query := `SELECT * FROM history_ujian WHERE id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &history, query, id)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *historyRepository) FindByUserID(ctx context.Context, userID int64) ([]model.HistoryUjian, error) {
	var histories []model.HistoryUjian
	query := `SELECT * FROM history_ujian WHERE user_id = $1 ORDER BY created_at DESC`
	err := getRunner(ctx, r.db).SelectContext(ctx, &histories, query, userID)
	return histories, err
}

func (r *historyRepository) FindByStatus(ctx context.Context, status model.StatusUjian) ([]model.HistoryUjian, error) {
	var histories []model.HistoryUjian
	query := `SELECT * FROM history_ujian WHERE status = $1 ORDER BY created_at DESC`
	err := getRunner(ctx, r.db).SelectContext(ctx, &histories, query, status)
	return histories, err
}

func (r *historyRepository) FindUserHistoryByBundleID(ctx context.Context, userID int64, bundleID int64) (*model.HistoryUjian, error) {
	var history model.HistoryUjian
	query := `SELECT * FROM history_ujian WHERE user_id = $1 AND bundle_id = $2 ORDER BY created_at DESC LIMIT 1`
	err := getRunner(ctx, r.db).GetContext(ctx, &history, query, userID, bundleID)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *historyRepository) FindOngoingHistoryByBundleID(ctx context.Context, userID int64, bundleID int64) (*model.HistoryUjian, error) {
	var history model.HistoryUjian
	query := `SELECT * FROM history_ujian WHERE user_id = $1 AND bundle_id = $2 AND status = 'berlangsung' ORDER BY created_at DESC LIMIT 1`
	err := getRunner(ctx, r.db).GetContext(ctx, &history, query, userID, bundleID)
	if err != nil {
		return nil, err
	}
	return &history, nil
}
