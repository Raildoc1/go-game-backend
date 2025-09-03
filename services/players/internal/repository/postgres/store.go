package postgresrepo

import (
	"context"

	postgresstore "go-game-backend/pkg/postgres"
	playersvc "go-game-backend/services/players/internal/services/player"
)

// Store wraps postgres storage to expose repository interfaces.
type Store struct {
	inner *postgresstore.Storage[Repos]
}

// NewStore creates new Store.
func NewStore(s *postgresstore.Storage[Repos]) *Store { return &Store{inner: s} }

// DoTx executes a transactional function.
func (s *Store) DoTx(ctx context.Context, f func(context.Context, playersvc.PostgresRepos) error) error {
	return s.inner.DoTx(ctx, func(ctx context.Context, r *Repos) error { return f(ctx, r) })
}

// Raw returns access to repositories without transaction.
func (s *Store) Raw() playersvc.PostgresRepos { return s.inner.Raw() }
