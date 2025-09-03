// Package kafkaingester consumes Kafka events
package kafkaingester

import (
	"context"
	"encoding/json"
	"fmt"
	"go-game-backend/pkg/futils"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/players/internal/ws"

	k "github.com/segmentio/kafka-go"

	authmodels "go-game-backend/services/auth/pkg/models"

	"go.uber.org/zap"
)

// UserCreated processes user-created events from Kafka.
type UserCreated struct {
	reader  *k.Reader
	logger  *logging.ZapLogger
	handler userCreatedHandler
	hub     *ws.Hub
	locker  playerLocker
}

type userCreatedHandler interface {
	HandleUserCreated(ctx context.Context, evt authmodels.UserCreatedEvent) error
}

type playerLocker interface {
	DoWithPlayerLock(ctx context.Context, userID int64, f futils.CtxF) error
}

// NewUserCreated creates a new UserCreated ingester.
func NewUserCreated(
	reader *k.Reader,
	logger *logging.ZapLogger,
	h userCreatedHandler,
	hub *ws.Hub,
	locker playerLocker,
) *UserCreated {
	return &UserCreated{
		reader:  reader,
		logger:  logger,
		handler: h,
		hub:     hub,
		locker:  locker,
	}
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

		err = i.locker.DoWithPlayerLock(ctx, evt.UserID, func(ctx context.Context) error {
			return i.handler.HandleUserCreated(ctx, evt)
		})
		if err != nil {
			i.logger.ErrorCtx(ctx, "handle user-created", zap.Error(err))
			continue
		}
		i.hub.Notify(evt.UserID, "profile-updated")
		i.logger.InfoCtx(ctx, "received user-created event", zap.Int64("user_id", evt.UserID))
	}
}
