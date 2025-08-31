// Package postgresstore provides PostgreSQL storage with transaction helpers.
package postgresstore

import (
	"context"
	"fmt"

	"go-game-backend/pkg/futils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection settings.
type Config struct {
	DSN string `yaml:"dsn"`
}

// Storage manages PostgreSQL connection pool and repositories.
type Storage[TRepos any] struct {
	cfg   *Config
	pool  *pgxpool.Pool
	repos *TRepos
}

// New creates a new Storage with the given configuration and repository factory.
func New[TRepos any](ctx context.Context, cfg *Config, factory func(*pgxpool.Pool) *TRepos) (*Storage[TRepos], error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgx pool connect: %w", err)
	}
	return &Storage[TRepos]{cfg: cfg, pool: pool, repos: factory(pool)}, nil
}

// Pool exposes underlying pgx pool.
func (s *Storage[TRepos]) Pool() *pgxpool.Pool {
	return s.pool
}

// Stop closes the database connection pool.
func (s *Storage[TRepos]) Stop() error {
	s.pool.Close()
	return nil
}

// DoTx executes the provided function within a database transaction.
func (s *Storage[TRepos]) DoTx(ctx context.Context, f futils.CtxFT[*TRepos]) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	ctxWithTx := ctxWithTx(ctx, tx)
	if err := f(ctxWithTx, s.repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// Raw returns underlying repositories instance.
func (s *Storage[TRepos]) Raw() *TRepos {
	return s.repos
}
