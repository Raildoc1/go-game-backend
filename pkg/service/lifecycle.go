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

// CloserWithError represents a component that can be closed and may return
// an error during the closing process.
type CloserWithError interface {
	Close() error
}

// Stop attempts to stop the provided target and logs any error that occurs
// using the supplied logger.
func Stop(ctx context.Context, target StopperWithError, name string, logger *logging.ZapLogger) {
	err := target.Stop()
	if err != nil {
		logger.ErrorCtx(ctx, fmt.Sprintf("failed to stop %s", name), zap.Error(err))
	}
}

// Close attempts to close the provided target and logs any error that occurs
// using the supplied logger.
func Close(ctx context.Context, target CloserWithError, name string, logger *logging.ZapLogger) {
	err := target.Close()
	if err != nil {
		logger.ErrorCtx(ctx, fmt.Sprintf("failed to close %s", name), zap.Error(err))
	}
}
