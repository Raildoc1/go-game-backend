package postgresrepo

import (
	"context"

	postgresstore "go-game-backend/pkg/postgres"
	authsvc "go-game-backend/services/auth/internal/services/auth"
)

// Store wraps postgres storage to expose repository interfaces.
type Store struct {
	inner *postgresstore.Storage[Repos]
}

// NewStore creates a new Store wrapper around the provided postgres storage.
func NewStore(s *postgresstore.Storage[Repos]) *Store {
	return &Store{inner: s}
}

// DoTx executes a transactional function using repository interfaces.
func (s *Store) DoTx(ctx context.Context, f func(ctx context.Context, r authsvc.PostgresRepos) error) error {
	//nolint:wrapcheck // unnecessary
	return s.inner.DoTx(ctx, func(ctx context.Context, r *Repos) error {
		return f(ctx, r)
	})
}

// Raw returns access to repositories without a transaction.
func (s *Store) Raw() authsvc.PostgresRepos { return s.inner.Raw() }
