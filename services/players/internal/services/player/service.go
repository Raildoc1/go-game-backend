package playersvc

import (
	"context"
	"errors"
	"fmt"
	"go-game-backend/services/players/pkg/models"

	authmodels "go-game-backend/services/auth/pkg/models"
)

// ErrNotFound is returned when player does not exist.
var ErrNotFound = errors.New("player not found")

// New creates player service instance.
func New(cfg *Config, pg PostgresStore) *Service {
	return &Service{cfg: cfg, pgStore: pg}
}

// GetInitialState creates player if necessary and returns aggregated state.
func (s *Service) GetInitialState(ctx context.Context, userID int64) (*models.InitialState, error) {
	p, err := s.pgStore.Raw().Player().Get(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if err := s.pgStore.Raw().Player().CreateOrUpdate(ctx, userID, ""); err != nil {
				return nil, fmt.Errorf("create player: %w", err)
			}
			p = Player{UserID: userID}
		} else {
			return nil, err
		}
	}
	state := &models.InitialState{
		Nickname: p.Nickname,
		Wallet:   models.WalletState{Soft: 100, Hard: 10},
	}
	return state, nil
}

// HandleUserCreated creates or updates player profile using event data.
func (s *Service) HandleUserCreated(ctx context.Context, evt authmodels.UserCreatedEvent) error {
	if err := s.pgStore.Raw().Player().CreateOrUpdate(ctx, evt.UserID, evt.Nickname);  err != nil {
		return fmt.Errorf("create or update player: %w", err)
	}
	return nil
}
