package postgresrepo

import (
	outboxpkg "go-game-backend/pkg/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repos aggregates all PostgreSQL repositories used by the auth service.
type Repos struct {
	user   *UserRepo
	outbox *outboxpkg.Repository
}

// NewRepos creates Repos with initialized sub-repositories.
func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{
		user:   NewUserRepo(pool),
		outbox: outboxpkg.NewRepository(pool),
	}
}

// User returns repository for user credentials.
func (r *Repos) User() *UserRepo { return r.user }

// Outbox returns repository for the outbox table.
func (r *Repos) Outbox() *outboxpkg.Repository { return r.outbox }
