package redisrepo

import (
	"context"
	"fmt"
	"go-game-backend/pkg/futils"
	"go-game-backend/services/auth/internal/dto"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	redisstore "go-game-backend/pkg/redis"

	auth "go-game-backend/services/auth/internal/services/auth"
)

// Repository implements a session repository backed by Redis.
type Repository struct {
	store *redisstore.Storage
}

var _ auth.SessionRepository = (*Repository)(nil)

// New creates a new Repository instance.
func New(store *redisstore.Storage) *Repository {
	return &Repository{
		store: store,
	}
}

// DoWithTransaction executes the given function within a Redis transaction.
func (r *Repository) DoWithTransaction(ctx context.Context, f futils.CtxF) error {
	return r.store.DoWithTransaction(ctx, f)
}

// DoWithPlayerLock obtains a distributed lock for a specific user and runs the
// provided function while holding that lock.
func (r *Repository) DoWithPlayerLock(ctx context.Context, userID int64, ttl time.Duration, f futils.CtxF) error {
	key := fmt.Sprintf("lock:player:%v", userID)
	err := r.store.DoWithLock(ctx, key, ttl, f)
	if err != nil {
		return err
	}
	return nil
}

// SetSessionToken stores a session token for the given user with an expiration
// time.
func (r *Repository) SetSessionToken(ctx context.Context, userID int64, token uuid.UUID, expiresAt time.Time) error {
	return r.store.Do(ctx, func(ctx context.Context, cmdable redis.Cmdable) error {
		key := fmt.Sprintf("session_token:%v", userID)

		setCmd := cmdable.Set(ctx, key, token.String(), 0)
		if err := setCmd.Err(); err != nil {
			return fmt.Errorf("failed to set session token: %w", err)
		}

		expCmd := cmdable.ExpireAt(ctx, key, expiresAt)
		if err := expCmd.Err(); err != nil {
			return fmt.Errorf("failed to set session token expiration time: %w", err)
		}

		return nil
	})
}

// SetRefreshToken stores a refresh token and its associated session info.
func (r *Repository) SetRefreshToken(
	ctx context.Context,
	token uuid.UUID,
	sessionInfo dto.SessionInfo,
	expiresAt time.Time,
) error {
	return r.store.Do(ctx, func(ctx context.Context, cmdable redis.Cmdable) error {
		key := fmt.Sprintf("refresh_token:%s", token.String())

		setRes := cmdable.HSet(ctx, key,
			"session_token", sessionInfo.SessionToken.String(),
			"user_id", sessionInfo.UserID)
		if err := setRes.Err(); err != nil {
			return fmt.Errorf("redis failed to set session token: %w", err)
		}

		expRes := cmdable.ExpireAt(ctx, key, expiresAt)
		if err := expRes.Err(); err != nil {
			return fmt.Errorf("redis failed to set session token: %w", err)
		}

		return nil
	})
}

// RemoveRefreshToken deletes a refresh token from the store.
func (r *Repository) RemoveRefreshToken(ctx context.Context, token uuid.UUID) error {
	return r.store.Do(ctx, func(ctx context.Context, cmdable redis.Cmdable) error {
		key := fmt.Sprintf("refresh_token:%s", token)

		res := cmdable.Del(ctx, key)
		if err := res.Err(); err != nil {
			return fmt.Errorf("redis failed to remove refresh token: %w", err)
		}

		return nil
	})
}

// GetSessionInfo retrieves session information stored under the given refresh
// token.
func (r *Repository) GetSessionInfo(ctx context.Context, refreshToken uuid.UUID) (dto.SessionInfo, error) {
	var sessionInfo dto.SessionInfo
	err := r.store.Do(ctx, func(ctx context.Context, cmdable redis.Cmdable) error {
		key := fmt.Sprintf("refresh_token:%s", refreshToken)

		resCmd := cmdable.HGetAll(ctx, key)
		if err := resCmd.Err(); err != nil {
			return fmt.Errorf("redis failed to get session info: %w", err)
		}

		err := resCmd.Scan(&sessionInfo)
		if err != nil {
			return fmt.Errorf("failed to parse session info: %w", err)
		}

		return nil
	})
	if err != nil {
		return dto.SessionInfo{}, err
	}
	return sessionInfo, nil
}
