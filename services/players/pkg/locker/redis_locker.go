// Package locker provides utilities for locking player-related resources.
package locker

import (
	"context"
	"time"

	"go-game-backend/pkg/futils"
	redisstore "go-game-backend/pkg/redis"
	playerredis "go-game-backend/services/players/pkg/redis"
)

// LockDoer abstracts locking capabilities of a storage.
type LockDoer interface {
	DoWithLock(ctx context.Context, key string, ttl time.Duration, f futils.CtxF) error
}

// RedisPlayerLocker uses Redis to lock operations on a player.
type RedisPlayerLocker struct {
	store LockDoer
	ttl   time.Duration
}

// NewRedisPlayerLocker creates a new player locker backed by Redis.
func NewRedisPlayerLocker(store LockDoer, ttl time.Duration) *RedisPlayerLocker {
	return &RedisPlayerLocker{store: store, ttl: ttl}
}

// NewFromStorage builds RedisPlayerLocker from a redisstore.Storage.
func NewFromStorage[T any](store *redisstore.Storage[T], ttl time.Duration) *RedisPlayerLocker {
	return NewRedisPlayerLocker(store, ttl)
}

// DoWithPlayerLock obtains a lock for the given user ID and executes f.
func (l *RedisPlayerLocker) DoWithPlayerLock(ctx context.Context, userID int64, f futils.CtxF) error {
	key := playerredis.PlayerLockKey(userID)
	return l.store.DoWithLock(ctx, key, l.ttl, f)
}
