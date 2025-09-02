// Package kafka provides helpers and config for interacting with Kafka
package kafka

import (
	"time"

	k "github.com/segmentio/kafka-go"
)

// ForwarderConfig contains settings for the outbox forwarder.
type ForwarderConfig struct {
	Brokers      []string      `yaml:"brokers"`
	PollInterval time.Duration `yaml:"poll-interval"`
	BatchSize    int32         `yaml:"batch-size"`
}

// ReaderConfig contains Kafka reader settings.
type ReaderConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group-id"`
}

// NewWriter creates a kafka writer for the given brokers.
func NewWriter(brokers []string) *k.Writer {
	return &k.Writer{
		Addr:                   k.TCP(brokers...),
		AllowAutoTopicCreation: false,
		RequiredAcks:           k.RequireAll,
	}
}

// NewReader creates a kafka reader from the provided configuration.
func NewReader(cfg *ReaderConfig) *k.Reader {
	return k.NewReader(k.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
}
