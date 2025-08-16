// Package postgresrepo provides PostgreSQL repository implementation.
package postgresrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-game-backend/services/profile-storage/internal/repository/postgres/sqlc"
)

// Repository provides access to player credentials stored in PostgreSQL.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// New creates a new Repository using the provided pgx pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

// Stop closes resources held by the repository. The underlying pool is managed externally.
func (r *Repository) Stop() error { return nil }

// AddUser inserts a new user with the specified login token and returns its ID.
func (r *Repository) AddUser(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	id, err := r.queries.AddUser(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

// FindUserByLoginToken retrieves a user ID by its login token.
func (r *Repository) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	id, err := r.queries.FindUserByLoginToken(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("find user: %w", err)
	}
	return id, nil
}
