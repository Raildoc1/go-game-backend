package redisrepo

import (
	"github.com/redis/go-redis/v9"
	auth "go-game-backend/services/auth/internal/services/auth"
)

type Repos struct {
	session *SessionRepo
}

var _ auth.RxRepos = (*Repos)(nil)

func NewRepos(defaultCmdable redis.Cmdable) *Repos {
	return &Repos{
		session: NewSessionRepo(defaultCmdable),
	}
}

func (r Repos) Session() auth.SessionRepo {
	return r.session
}
