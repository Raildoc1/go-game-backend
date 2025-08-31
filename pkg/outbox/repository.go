package outbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-game-backend/pkg/outbox/sqlc"
)

// Repository provides access to outbox events stored in PostgreSQL.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewRepository creates a new Repository with the given pgx pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

// Add inserts a new event into the outbox table within the provided transaction.
func (r *Repository) Add(ctx context.Context, tx pgx.Tx, topic string, payload []byte) error {
	q := r.queries.WithTx(tx)
	if err := q.AddEvent(ctx, sqlc.AddEventParams{Topic: topic, Payload: payload}); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

// Fetch returns a batch of unprocessed events.
func (r *Repository) Fetch(ctx context.Context, limit int) ([]Event, error) {
	rows, err := r.queries.FetchEvents(ctx, int32(limit))
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
	if err := r.queries.MarkProcessed(ctx, id); err != nil {
		return fmt.Errorf("mark outbox event processed: %w", err)
	}
	return nil
}
