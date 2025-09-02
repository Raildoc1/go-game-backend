package authsvc

import (
	"context"
	"time"

	"go-game-backend/services/auth/internal/dto"

	"github.com/google/uuid"
)

// UserRepository defines operations for managing users in persistent storage.
type UserRepository interface {
	AddUser(ctx context.Context, loginToken uuid.UUID) (int64, error)
	FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error)
}

// OutboxRepository defines operations for working with outbox events.
type OutboxRepository interface {
	AddJSON(ctx context.Context, topic string, payload any) error
}

// PostgresRepos aggregates repositories backed by PostgreSQL.
type PostgresRepos interface {
	User() UserRepository
	Outbox() OutboxRepository
}

// PostgresStore provides transactional access to PostgreSQL repositories.
type PostgresStore interface {
	DoTx(ctx context.Context, f func(ctx context.Context, r PostgresRepos) error) error
	Raw() PostgresRepos
}

// SessionRepository defines operations for managing authentication sessions.
type SessionRepository interface {
	SetSessionToken(ctx context.Context, userID int64, token uuid.UUID, expiresAt time.Time) error
	SetRefreshToken(ctx context.Context, token uuid.UUID, sessionInfo dto.SessionInfo, expiresAt time.Time) error
	RemoveRefreshToken(ctx context.Context, token uuid.UUID) error
	GetSessionInfo(ctx context.Context, refreshToken uuid.UUID) (dto.SessionInfo, error)
}

// RedisRepos aggregates repositories backed by Redis.
type RedisRepos interface {
	Session() SessionRepository
}

// RedisStore provides transactional access to Redis repositories.
type RedisStore interface {
	DoTx(ctx context.Context, f func(ctx context.Context, r RedisRepos) error) error
	Raw() RedisRepos
}
