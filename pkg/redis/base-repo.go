package redisstore

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// BaseRepo provides helper methods to access redis.Cmdable bound to context transactions.
type BaseRepo struct {
	defaultCmdable redis.Cmdable
}

// NewBaseRepo create new BaseRepo instance
func NewBaseRepo(defaultCmdable redis.Cmdable) BaseRepo {
	return BaseRepo{
		defaultCmdable: defaultCmdable,
	}
}

// Cmd return either pipeline redis.Cmdable retrieved from context or the default one
func (r BaseRepo) Cmd(ctx context.Context) redis.Cmdable {
	if cmdable := cmdableFromCtx(ctx); cmdable != nil {
		return cmdable
	}
	return r.defaultCmdable
}
