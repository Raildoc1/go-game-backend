package redisstore

import (
	"context"
	"fmt"
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
type Storage struct {
	cfg    *Config
	rdb    *redis.Client
	locker *redislock.Client
	logger *logging.ZapLogger
}

// New creates a Storage instance configured with the provided settings and
// logger.
func New(cfg *Config, logger *logging.ZapLogger) *Storage {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.ServerAddr})
	locker := redislock.New(rdb)
	return &Storage{
		cfg:    cfg,
		rdb:    rdb,
		locker: locker,
		logger: logger,
	}
}

// Stop closes the underlying Redis client connection.
func (s *Storage) Stop() error {
	err := s.rdb.Close()
	if err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}
	return nil
}

// DoWithTransaction executes the provided function within a Redis transaction
// context.
func (s *Storage) DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		ctxWithPipe := context.WithValue(ctx, redisPipelineKey, pipe)

		return f(ctxWithPipe)
	})
	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}
	return nil
}

// Do executes the provided function with a Redis Cmdable derived from the
// context.
func (s *Storage) Do(ctx context.Context, f func(ctx context.Context, cmdable redis.Cmdable) error) error {
	cmdable, err := s.getCmdFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cmdable: %w", err)
	}

	err = f(ctx, cmdable)
	if err != nil {
		return err
	}

	return nil
}

// DoWithLock obtains a distributed lock for the specified key and executes the
// supplied function while holding the lock.
func (s *Storage) DoWithLock(
	ctx context.Context,
	key string,
	ttl time.Duration,
	f func(ctx context.Context) error,
) error {
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

func (s *Storage) getCmdFromCtx(ctx context.Context) (redis.Cmdable, error) {
	p := ctx.Value(redisPipelineKey)
	if p == nil {
		return s.rdb, nil
	}
	pipe, ok := p.(redis.Pipeliner)
	if !ok {
		return nil, fmt.Errorf("invalid redis pipeliner in context")
	}
	return pipe, nil
}
