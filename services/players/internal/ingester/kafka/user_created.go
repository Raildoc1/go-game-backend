// Package kafkaingester consumes Kafka events
package kafkaingester

import (
	"context"
	"encoding/json"
	"fmt"
	"go-game-backend/pkg/logging"

	k "github.com/segmentio/kafka-go"

	authmodels "go-game-backend/services/auth/pkg/models"

	"go.uber.org/zap"
)

// UserCreated processes user-created events from Kafka.
type UserCreated struct {
	reader *k.Reader
	logger *logging.ZapLogger
}

// NewUserCreated creates a new UserCreated ingester.
func NewUserCreated(reader *k.Reader, logger *logging.ZapLogger) *UserCreated {
	return &UserCreated{reader: reader, logger: logger}
}

// Run starts consuming user-created events until the context is done.
func (i *UserCreated) Run(ctx context.Context) error {
	for {
		m, err := i.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("read kafka message: %w", err)
		}
		var evt authmodels.UserCreatedEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			i.logger.ErrorCtx(ctx, "unmarshal user-created event", zap.Error(err))
			continue
		}
		i.logger.InfoCtx(ctx, "received user-created event", zap.Int64("user_id", evt.UserID))
	}
}
