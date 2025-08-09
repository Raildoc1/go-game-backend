package service

import (
	"context"
	"fmt"
	"go-game-backend/pkg/logging"
	"go.uber.org/zap"
)

type StopperWithError interface {
	Stop() error
}

func Stop(ctx context.Context, target StopperWithError, name string, logger *logging.ZapLogger) {
	err := target.Stop()
	if err != nil {
		logger.ErrorCtx(ctx, fmt.Sprintf("failed to stop %s", name), zap.Error(err))
	}
}
