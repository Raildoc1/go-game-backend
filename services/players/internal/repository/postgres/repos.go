package postgresrepo

import (
	playersvc "go-game-backend/services/players/internal/services/player"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repos aggregates all PostgreSQL repositories used by players service.
type Repos struct {
	player playersvc.PlayerRepository
}

// NewRepos creates repository bundle.
func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{player: NewPlayerRepo(pool)}
}

// Player returns player repository.
func (r *Repos) Player() playersvc.PlayerRepository { return r.player }
