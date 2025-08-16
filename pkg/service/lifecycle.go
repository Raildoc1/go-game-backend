package service

import (
	"context"
	"fmt"
	"go-game-backend/pkg/logging"

	"go.uber.org/zap"
)

// StopperWithError represents a component that can be stopped and may return
// an error during the stopping process.
type StopperWithError interface {
	Stop() error
}

// Stop attempts to stop the provided target and logs any error that occurs
// using the supplied logger.
func Stop(ctx context.Context, target StopperWithError, name string, logger *logging.ZapLogger) {
	err := target.Stop()
	if err != nil {
		logger.ErrorCtx(ctx, fmt.Sprintf("failed to stop %s", name), zap.Error(err))
	}
}
