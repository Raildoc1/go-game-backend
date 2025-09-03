package postgresrepo

import (
	"context"
	"errors"
	"fmt"
	"go-game-backend/services/players/internal/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	postgresstore "go-game-backend/pkg/postgres"

	playersvc "go-game-backend/services/players/internal/services/player"
)

// PlayerRepo provides access to players table.
type PlayerRepo struct {
	postgresstore.BaseRepo[*sqlc.Queries]
}

// NewPlayerRepo creates a repository instance.
func NewPlayerRepo(pool *pgxpool.Pool) *PlayerRepo {
	return &PlayerRepo{BaseRepo: postgresstore.NewBaseRepo(sqlc.New(pool))}
}

// CreateOrUpdate inserts or updates player profile.
func (r *PlayerRepo) CreateOrUpdate(ctx context.Context, userID int64, nickname string) error {
	err := r.Q(ctx).UpsertPlayer(ctx, sqlc.UpsertPlayerParams{
		UserID:   userID,
		Nickname: pgtype.Text{String: nickname, Valid: nickname != ""},
	})
	if err != nil {
		return fmt.Errorf("upsert player: %w", err)
	}
	return nil
}

// Get retrieves player by ID.
func (r *PlayerRepo) Get(ctx context.Context, userID int64) (playersvc.Player, error) {
	p, err := r.Q(ctx).GetPlayer(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return playersvc.Player{}, playersvc.ErrNotFound
		}
		return playersvc.Player{}, fmt.Errorf("get player: %w", err)
	}
	nickname := ""
	if p.Nickname.Valid {
		nickname = p.Nickname.String
	}
	return playersvc.Player{UserID: p.UserID, Nickname: nickname}, nil
}
