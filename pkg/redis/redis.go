// Package redisstore provides helpers for interacting with Redis, including
// transaction management and distributed locking.
package redisstore

import (
	"context"
	"fmt"
	"go-game-backend/pkg/futils"
	"go-game-backend/pkg/logging"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"

	"go.uber.org/zap"
)

type ctxValueKey string

const redisPipelineKey ctxValueKey = "redisPipeline"

// Config defines options for connecting to a Redis server.
type Config struct {
	ServerAddr string `yaml:"server-address"`
}

// Storage wraps a Redis client and provides helpers for executing commands
// and managing distributed locks.
type Storage[TRepos any] struct {
	cfg    *Config
	rdb    *redis.Client
	locker *redislock.Client
	logger *logging.ZapLogger
	repos  *TRepos
}

// New creates a Storage instance configured with the provided settings and
// logger.
func New[TRepos any](
	cfg *Config,
	logger *logging.ZapLogger,
	factory func(redis.Cmdable) *TRepos,
) *Storage[TRepos] {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.ServerAddr})
	locker := redislock.New(rdb)
	return &Storage[TRepos]{
		cfg:    cfg,
		rdb:    rdb,
		locker: locker,
		logger: logger,
		repos:  factory(rdb),
	}
}

// Stop closes the underlying Redis client connection.
func (s *Storage[TRepos]) Stop() error {
	err := s.rdb.Close()
	if err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}
	return nil
}

// DoTx executes the provided function within a Redis transaction
// context.
func (s *Storage[TRepos]) DoTx(ctx context.Context, f futils.CtxFT[*TRepos]) error {
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		ctxWithPipe := context.WithValue(ctx, redisPipelineKey, pipe)

		return f(ctxWithPipe, s.repos)
	})
	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}
	return nil
}

func (s *Storage[TRepos]) Raw() *TRepos {
	return s.repos
}

// DoWithLock obtains a distributed lock for the specified key and executes the
// supplied function while holding the lock.
func (s *Storage[TRepos]) DoWithLock(ctx context.Context, key string, ttl time.Duration, f futils.CtxF) error {
	lock, err := s.locker.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return fmt.Errorf("failed to obtain lock: %w", err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		err := lock.Release(ctx)
		if err != nil {
			s.logger.ErrorCtx(
				ctx,
				"failed to release player lock",
				zap.Error(err),
				zap.String("key", key),
			)
		}
	}(lock, ctx)
	err = f(ctx)
	if err != nil {
		return err
	}
	return nil
}

func cmdableFromCtx(ctx context.Context) redis.Cmdable {
	if p := ctx.Value(redisPipelineKey); p != nil {
		return p.(redis.Pipeliner)
	}
	return nil
}
