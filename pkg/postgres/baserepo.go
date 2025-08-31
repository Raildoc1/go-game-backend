package postgresstore

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/outbox/sqlc"
)

type ctxField string

var txCtxField ctxField = "tx"

type BaseRepo struct {
	pool            *pgxpool.Pool
	queries         *sqlc.Queries
	logger          *logging.ZapLogger
	defaultIsoLevel pgx.TxIsoLevel
}

func NewBaseRepo(
	pool *pgxpool.Pool,
	queries *sqlc.Queries,
	logger *logging.ZapLogger,
	defaultIsoLevel pgx.TxIsoLevel,
) *BaseRepo {
	return &BaseRepo{
		pool:            pool,
		queries:         queries,
		logger:          logger,
		defaultIsoLevel: defaultIsoLevel,
	}
}

func (r *BaseRepo) DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       r.defaultIsoLevel,
		AccessMode:     pgx.ReadWrite,
	})
	q := r.queries.WithTx(tx)
}
