package redisrepo

import (
	"context"

	redisstore "go-game-backend/pkg/redis"
	playersvc "go-game-backend/services/players/internal/services/player"
)

// Store wraps redis storage to expose repository interfaces.
type Store struct {
	inner *redisstore.Storage[Repos]
}

// NewStore creates new Store.
func NewStore(s *redisstore.Storage[Repos]) *Store { return &Store{inner: s} }

// DoTx executes function within transaction.
func (s *Store) DoTx(ctx context.Context, f func(context.Context, playersvc.RedisRepos) error) error {
	return s.inner.DoTx(ctx, func(ctx context.Context, r *Repos) error { return f(ctx, r) })
}

// Raw returns access to repositories without transaction.
func (s *Store) Raw() playersvc.RedisRepos { return s.inner.Raw() }
