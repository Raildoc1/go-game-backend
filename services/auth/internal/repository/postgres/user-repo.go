// Package postgresrepo provides PostgreSQL repositories for the auth service.
package postgresrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	postgresstore "go-game-backend/pkg/postgres"
	"go-game-backend/services/auth/internal/repository/postgres/sqlc"
)

// UserRepo provides access to users stored in PostgreSQL.
type UserRepo struct {
	postgresstore.BaseRepo[*sqlc.Queries]
}

// NewUserRepo creates a new UserRepo instance bound to the given pool.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{BaseRepo: postgresstore.NewBaseRepo(sqlc.New(pool))}
}

// AddUser inserts a new user with the specified login token and returns its ID.
func (r *UserRepo) AddUser(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	id, err := r.Q(ctx).AddUser(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("insert user query: %w", err)
	}
	return id, nil
}

// FindUserByLoginToken retrieves a user ID by its login token.
func (r *UserRepo) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	id, err := r.Q(ctx).FindUserByLoginToken(ctx, pgtype.UUID{Bytes: loginToken, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("find user query: %w", err)
	}
	return id, nil
}
