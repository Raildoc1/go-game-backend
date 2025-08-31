package redisstore

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type BaseRepo struct {
	defaultCmdable redis.Cmdable
}

func NewBaseRepo(defaultCmdable redis.Cmdable) BaseRepo {
	return BaseRepo{
		defaultCmdable: defaultCmdable,
	}
}

func (r BaseRepo) Cmd(ctx context.Context) redis.Cmdable {
	if cmdable := cmdableFromCtx(ctx); cmdable != nil {
		return cmdable
	}
	return r.defaultCmdable
}
