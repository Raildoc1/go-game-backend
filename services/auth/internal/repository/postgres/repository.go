// Package postgresrepo provides PostgreSQL repository implementation.
package postgresrepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/auth/internal/repository/postgres/sqlc"
	"go.uber.org/zap"
)

// Repository provides access to player credentials stored in PostgreSQL.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	logger  *logging.ZapLogger
}

// New creates a new Repository using the provided pgx pool.
func New(pool *pgxpool.Pool, logger *logging.ZapLogger) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool), logger: logger}
}

// Stop closes resources held by the repository. The underlying pool is managed externally.
func (r *Repository) Stop() error { return nil }

// DoWithTransaction executes the provided function within a database transaction.
func (r *Repository) DoWithTransaction(ctx context.Context, f func(context.Context, pgx.Tx) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			if !errors.Is(pgx.ErrTxClosed, err) {
				r.logger.ErrorCtx(ctx, "rollback", zap.Error(err))
			}
		}
	}()

	if err := f(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// AddUser inserts a new user with the specified login token inside the transaction and returns its ID.
func (r *Repository) AddUser(ctx context.Context, tx pgx.Tx, loginToken uuid.UUID) (int64, error) {
	q := r.queries.WithTx(tx)
	id, err := q.AddUser(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("insert user query: %w", err)
	}
	return id, nil
}

// FindUserByLoginToken retrieves a user ID by its login token.
func (r *Repository) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	id, err := r.queries.FindUserByLoginToken(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("find user query: %w", err)
	}
	return id, nil
}
