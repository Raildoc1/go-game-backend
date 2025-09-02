package redisrepo

import (
	"context"

	redisstore "go-game-backend/pkg/redis"
	authsvc "go-game-backend/services/auth/internal/services/auth"
)

// Store wraps redis storage to expose repository interfaces.
type Store struct {
	inner *redisstore.Storage[Repos]
}

// NewStore creates a new Store wrapper around the provided redis storage.
func NewStore(s *redisstore.Storage[Repos]) *Store {
	return &Store{inner: s}
}

// DoTx executes a transactional function using repository interfaces.
func (s *Store) DoTx(ctx context.Context, f func(ctx context.Context, r authsvc.RedisRepos) error) error {
	//nolint:wrapcheck // unnecessary
	return s.inner.DoTx(ctx, func(ctx context.Context, r *Repos) error {
		return f(ctx, r)
	})
}

// Raw returns access to repositories without a transaction.
func (s *Store) Raw() authsvc.RedisRepos { return s.inner.Raw() }
