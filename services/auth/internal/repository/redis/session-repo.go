// Package redisrepo implements the session repository using Redis as storage.
package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	redisstore "go-game-backend/pkg/redis"
	"go-game-backend/services/auth/internal/dto"
)

// SessionRepo implements a session repository backed by Redis.
type SessionRepo struct {
	redisstore.BaseRepo
}

// NewSessionRepo creates a new Session Repository instance.
func NewSessionRepo(defaultCmdable redis.Cmdable) *SessionRepo {
	return &SessionRepo{
		redisstore.NewBaseRepo(defaultCmdable),
	}
}

// SetSessionToken stores a session token for the given user with an expiration time.
func (r *SessionRepo) SetSessionToken(ctx context.Context, userID int64, token uuid.UUID, expiresAt time.Time) error {
	key := fmt.Sprintf("session_token:%v", userID)

	setCmd := r.Cmd(ctx).Set(ctx, key, token.String(), 0)
	if err := setCmd.Err(); err != nil {
		return fmt.Errorf("redis: set '%s': %w", key, err)
	}

	expCmd := r.Cmd(ctx).ExpireAt(ctx, key, expiresAt)
	if err := expCmd.Err(); err != nil {
		return fmt.Errorf("redis: set '%s' expiration time: %w", key, err)
	}

	return nil
}

// SetRefreshToken stores a refresh token and its associated session info.
func (r *SessionRepo) SetRefreshToken(
	ctx context.Context,
	token uuid.UUID,
	sessionInfo dto.SessionInfo,
	expiresAt time.Time,
) error {
	key := fmt.Sprintf("refresh_token:%s", token.String())

	sessionToken := sessionInfo.SessionToken.String()
	userID := sessionInfo.UserID
	setRes := r.Cmd(ctx).HSet(ctx, key, "session_token", sessionToken, "user_id", userID)
	if err := setRes.Err(); err != nil {
		return fmt.Errorf("redis: set '%s': %w", key, err)
	}

	expRes := r.Cmd(ctx).ExpireAt(ctx, key, expiresAt)
	if err := expRes.Err(); err != nil {
		return fmt.Errorf("redis: set '%s' expiration time: %w", key, err)
	}

	return nil
}

// RemoveRefreshToken deletes a refresh token from the store.
func (r *SessionRepo) RemoveRefreshToken(ctx context.Context, token uuid.UUID) error {
	key := fmt.Sprintf("refresh_token:%s", token)

	res := r.Cmd(ctx).Del(ctx, key)
	if err := res.Err(); err != nil {
		return fmt.Errorf("redis: remove '%s': %w", key, err)
	}

	return nil
}

// GetSessionInfo retrieves session information stored under the given refresh token.
func (r *SessionRepo) GetSessionInfo(ctx context.Context, refreshToken uuid.UUID) (dto.SessionInfo, error) {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)

	resCmd := r.Cmd(ctx).HGetAll(ctx, key)
	if err := resCmd.Err(); err != nil {
		return dto.SessionInfo{}, fmt.Errorf("redis: get '%s': %w", key, err)
	}

	var sessionInfo dto.SessionInfo
	err := resCmd.Scan(&sessionInfo)
	if err != nil {
		return dto.SessionInfo{}, fmt.Errorf("parse session info: %w", err)
	}

	return sessionInfo, nil
}
