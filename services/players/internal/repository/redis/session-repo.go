package redisrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	redisstore "go-game-backend/pkg/redis"
)

// SessionRepo implements session repository backed by Redis.
type SessionRepo struct {
	redisstore.BaseRepo
}

// NewSessionRepo creates new SessionRepo.
func NewSessionRepo(cmd redis.Cmdable) *SessionRepo {
	return &SessionRepo{BaseRepo: redisstore.NewBaseRepo(cmd)}
}

// GetSessionToken retrieves session token for user.
func (r *SessionRepo) GetSessionToken(ctx context.Context, userID int64) (uuid.UUID, error) {
	key := fmt.Sprintf("session_token:%v", userID)
	res := r.Cmd(ctx).Get(ctx, key)
	if err := res.Err(); err != nil {
		return uuid.Nil, fmt.Errorf("redis get '%s': %w", key, err)
	}
	str, err := res.Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("redis result '%s': %w", key, err)
	}
	token, err := uuid.Parse(str)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid: %w", err)
	}
	return token, nil
}
