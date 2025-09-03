package playersvc

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Player represents stored player profile.
type Player struct {
	UserID   int64
	Nickname string
}

// PostgresRepos groups Postgres repositories used by the service.
type PostgresRepos interface {
	Player() PlayerRepository
}

// PostgresStore wraps Postgres storage with helpers.
type PostgresStore interface {
	DoTx(ctx context.Context, f func(context.Context, PostgresRepos) error) error
	Raw() PostgresRepos
}

// PlayerRepository defines database operations on players.
type PlayerRepository interface {
	CreateOrUpdate(ctx context.Context, userID int64, nickname string) error
	Get(ctx context.Context, userID int64) (Player, error)
}

// RedisRepos groups Redis repositories used by the service.
type RedisRepos interface {
	Session() SessionRepository
}

// RedisStore wraps Redis storage with helpers.
type RedisStore interface {
	DoTx(ctx context.Context, f func(context.Context, RedisRepos) error) error
	Raw() RedisRepos
}

// SessionRepository validates sessions stored in Redis.
type SessionRepository interface {
	GetSessionToken(ctx context.Context, userID int64) (uuid.UUID, error)
}

// Service exposes player related operations.
type Service struct {
	cfg     *Config
	pgStore PostgresStore
}

// Config holds player service configuration.
type Config struct {
	PlayerLockTTL time.Duration `yaml:"player-lock-ttl"`
}
