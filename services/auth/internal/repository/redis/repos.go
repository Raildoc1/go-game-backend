package redisrepo

import (
	authsvc "go-game-backend/services/auth/internal/services/auth"

	"github.com/redis/go-redis/v9"
)

// Repos aggregates all Redis-backed repositories used by the auth service.
type Repos struct {
	session authsvc.SessionRepository
}

// NewRepos creates Repos with initialized sub-repositories.
func NewRepos(defaultCmdable redis.Cmdable) *Repos {
	return &Repos{
		session: NewSessionRepo(defaultCmdable),
	}
}

// Session returns repository for session data.
func (r Repos) Session() authsvc.SessionRepository {
	return r.session
}
