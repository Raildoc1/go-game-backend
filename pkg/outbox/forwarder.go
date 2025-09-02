package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Forwarder periodically sends events from the outbox to Kafka.
type Forwarder struct {
	store        *Repository
	writer       *kafka.Writer
	pollInterval time.Duration
	batchSize    int32
}

// NewForwarder creates a new Forwarder instance.
func NewForwarder(store *Repository, writer *kafka.Writer, pollInterval time.Duration, batchSize int32) *Forwarder {
	return &Forwarder{store: store, writer: writer, pollInterval: pollInterval, batchSize: batchSize}
}

// Run starts the forwarder loop and blocks until the context is done.
func (f *Forwarder) Run(ctx context.Context) {
	ticker := time.NewTicker(f.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = f.processBatch(ctx)
		}
	}
}

func (f *Forwarder) processBatch(ctx context.Context) error {
	events, err := f.store.Fetch(ctx, f.batchSize)
	if err != nil {
		return err
	}
	for _, e := range events {
		msg := kafka.Message{Topic: e.Topic, Value: e.Payload}
		if err := f.writer.WriteMessages(ctx, msg); err != nil {
			// if publishing fails, stop processing to retry later
			return fmt.Errorf("write messages: %w", err)
		}
		if err := f.store.MarkProcessed(ctx, e.ID); err != nil {
			return fmt.Errorf("mark processed: %w", err)
		}
	}
	return nil
}
