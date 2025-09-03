package redisrepo

import (
	playersvc "go-game-backend/services/players/internal/services/player"

	"github.com/redis/go-redis/v9"
)

// Repos aggregates Redis repositories used by players service.
type Repos struct {
	session playersvc.SessionRepository
}

// NewRepos builds repository bundle.
func NewRepos(cmd redis.Cmdable) *Repos {
	return &Repos{session: NewSessionRepo(cmd)}
}

// Session returns session repository.
func (r *Repos) Session() playersvc.SessionRepository { return r.session }
