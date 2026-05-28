package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type MapelRepository interface {
	FindByJenjangID(ctx context.Context, jenjangID int) ([]model.Mapel, error)
	FindByNamaAndJenjang(ctx context.Context, nama string, jenjangID int) (*model.Mapel, error)
}

type mapelRepository struct {
	db *sqlx.DB
}

func NewMapelRepository(db *sqlx.DB) MapelRepository {
	return &mapelRepository{db: db}
}

func (r *mapelRepository) FindByJenjangID(ctx context.Context, jenjangID int) ([]model.Mapel, error) {
	var mapels []model.Mapel
	query := `SELECT * FROM mapel WHERE jenjang_id = $1`
	err := getRunner(ctx, r.db).SelectContext(ctx, &mapels, query, jenjangID)
	return mapels, err
}

func (r *mapelRepository) FindByNamaAndJenjang(ctx context.Context, nama string, jenjangID int) (*model.Mapel, error) {
	var mapel model.Mapel
	query := `SELECT * FROM mapel WHERE nama = $1 AND jenjang_id = $2`
	err := getRunner(ctx, r.db).GetContext(ctx, &mapel, query, nama, jenjangID)
	if err != nil {
		return nil, err
	}
	return &mapel, nil
}
