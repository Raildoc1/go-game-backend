package postgresrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-game-backend/services/profile-storage/internal/repository/postgres/sqlc"
)

// Config holds PostgreSQL connection settings.
type Config struct {
	DSN string `yaml:"dsn"`
}

// Repository provides access to player credentials stored in PostgreSQL.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// New creates a new Repository with the given configuration.
func New(ctx context.Context, cfg *Config) (*Repository, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgx pool connect: %w", err)
	}
	return &Repository{pool: pool, queries: sqlc.New(pool)}, nil
}

// Stop closes the database connection pool.
func (r *Repository) Stop() error {
	r.pool.Close()
	return nil
}

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
