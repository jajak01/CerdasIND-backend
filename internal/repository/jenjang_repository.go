package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type JenjangRepository interface {
	FindAll(ctx context.Context) ([]model.Jenjang, error)
	FindByNama(ctx context.Context, nama string) (*model.Jenjang, error)
}

type jenjangRepository struct {
	db *sqlx.DB
}

func NewJenjangRepository(db *sqlx.DB) JenjangRepository {
	return &jenjangRepository{db: db}
}

func (r *jenjangRepository) FindAll(ctx context.Context) ([]model.Jenjang, error) {
	var jenjangs []model.Jenjang
	query := `SELECT * FROM jenjang`
	err := getRunner(ctx, r.db).SelectContext(ctx, &jenjangs, query)
	return jenjangs, err
}

func (r *jenjangRepository) FindByNama(ctx context.Context, nama string) (*model.Jenjang, error) {
	var jenjang model.Jenjang
	query := `SELECT * FROM jenjang WHERE nama = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &jenjang, query, nama)
	if err != nil {
		return nil, err
	}
	return &jenjang, nil
}
