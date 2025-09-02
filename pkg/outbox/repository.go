package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"go-game-backend/pkg/outbox/sqlc"
	postgresstore "go-game-backend/pkg/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides access to outbox events stored in PostgreSQL.
type Repository struct {
	postgresstore.BaseRepo[*sqlc.Queries]
}

// NewRepository creates a new Repository with the given pgx pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{BaseRepo: postgresstore.NewBaseRepo(sqlc.New(pool))}
}

// Add inserts a new event into the outbox table.
func (r *Repository) Add(ctx context.Context, topic string, payload []byte) error {
	if err := r.Q(ctx).AddEvent(ctx, sqlc.AddEventParams{Topic: topic, Payload: payload}); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

// AddJSON inserts a new event into the outbox table, converting payload to JSON.
func (r *Repository) AddJSON(ctx context.Context, topic string, payload any) error {
	js, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	if err := r.Q(ctx).AddEvent(ctx, sqlc.AddEventParams{Topic: topic, Payload: js}); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

// Fetch returns a batch of unprocessed events.
func (r *Repository) Fetch(ctx context.Context, limit int32) ([]Event, error) {
	rows, err := r.Q(ctx).FetchEvents(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox events: %w", err)
	}
	events := make([]Event, len(rows))
	for i, row := range rows {
		events[i] = Event{ID: row.ID, Topic: row.Topic, Payload: row.Payload}
	}
	return events, nil
}

// MarkProcessed marks the event as processed.
func (r *Repository) MarkProcessed(ctx context.Context, id int64) error {
	if err := r.Q(ctx).MarkProcessed(ctx, id); err != nil {
		return fmt.Errorf("mark outbox event processed: %w", err)
	}
	return nil
}
