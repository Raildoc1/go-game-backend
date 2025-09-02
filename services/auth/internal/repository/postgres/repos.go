package postgresrepo

import (
	outboxpkg "go-game-backend/pkg/outbox"
	authsvc "go-game-backend/services/auth/internal/services/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repos aggregates all PostgreSQL repositories used by the auth service.
type Repos struct {
	user   authsvc.UserRepository
	outbox authsvc.OutboxRepository
}

// NewRepos creates Repos with initialized sub-repositories.
func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{
		user:   NewUserRepo(pool),
		outbox: outboxpkg.NewRepository(pool),
	}
}

// User returns repository for user credentials.
func (r *Repos) User() authsvc.UserRepository { return r.user }

// Outbox returns repository for the outbox table.
func (r *Repos) Outbox() authsvc.OutboxRepository { return r.outbox }
