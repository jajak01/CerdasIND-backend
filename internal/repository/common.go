package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// SqlxRunner defines a common interface for both *sqlx.DB and *sqlx.Tx
type SqlxRunner interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type txKey struct{}

// InjectTx embeds a transaction (*sqlx.Tx) in context
func InjectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// getRunner returns active transaction from context if exists, otherwise falls back to db
func getRunner(ctx context.Context, db *sqlx.DB) SqlxRunner {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
