package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type BundleRepository interface {
	FindByMapelID(ctx context.Context, mapelID int, onlyActive bool) ([]model.Bundle, error)
	FindAll(ctx context.Context) ([]model.Bundle, error)
	FindByID(ctx context.Context, id int64) (*model.Bundle, error)
	Create(ctx context.Context, bundle *model.Bundle) error
}

type bundleRepository struct {
	db *sqlx.DB
}

func NewBundleRepository(db *sqlx.DB) BundleRepository {
	return &bundleRepository{db: db}
}

func (r *bundleRepository) FindByMapelID(ctx context.Context, mapelID int, onlyActive bool) ([]model.Bundle, error) {
	var bundles []model.Bundle
	query := `SELECT * FROM bundles WHERE mapel_id = $1`
	if onlyActive {
		query += " AND is_active = true"
	}
	err := getRunner(ctx, r.db).SelectContext(ctx, &bundles, query, mapelID)
	return bundles, err
}

func (r *bundleRepository) FindAll(ctx context.Context) ([]model.Bundle, error) {
	var bundles []model.Bundle
	query := `SELECT * FROM bundles`
	err := getRunner(ctx, r.db).SelectContext(ctx, &bundles, query)
	return bundles, err
}

func (r *bundleRepository) FindByID(ctx context.Context, id int64) (*model.Bundle, error) {
	var bundle model.Bundle
	query := `SELECT * FROM bundles WHERE id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &bundle, query, id)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (r *bundleRepository) Create(ctx context.Context, bundle *model.Bundle) error {
	query := `INSERT INTO bundles (mapel_id, nama_bundle, deskripsi, waktu_menit, is_active, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, bundle.MapelID, bundle.NamaBundle, bundle.Deskripsi, bundle.WaktuMenit, bundle.IsActive, bundle.CreatedBy).
		Scan(&bundle.ID, &bundle.CreatedAt, &bundle.UpdatedAt)
}
