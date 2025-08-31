// Package main starts the players service.
package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/kafka"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
	playerkafka "go-game-backend/services/players/internal/ingester/kafka"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the players service.
type Config struct {
	Service         *service.Config           `yaml:"service"`
	HTTP            *service.HTTPServerConfig `yaml:"http"`
	Kafka           *kafka.ReaderConfig       `yaml:"kafka"`
	ShutdownTimeout time.Duration             `yaml:"shutdown-timeout"`
}

func main() {
	cfg, err := service.LoadConfig[Config](
		"./configs/default.yaml",
		os.Stdout,
		func(err error) {
			log.Fatal(err)
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logging.NewZapLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}

	ctx := logging.WithContextFields(
		context.Background(),
		zap.String("service", "players"),
		zap.String("version", cfg.Service.Version),
	)

	if err = run(ctx, cfg, logger); err != nil {
		logger.ErrorCtx(ctx, "application stopped with error", zap.Error(err))
	}
	logger.InfoCtx(ctx, "application stopped successfully")
}

func run(ctx context.Context, cfg *Config, logger *logging.ZapLogger) error {
	reader := kafka.NewReader(cfg.Kafka)
	ing := playerkafka.NewUserCreated(reader, logger)

	serv := service.NewBuilder().
		WithGo(func(ctx context.Context) error { return ing.Run(ctx) }).
		WithHTTPServer(cfg.HTTP, func() http.Handler {
			router := gin.Default()

			api := router.Group("/api/v1")
			{
				_ = api
			}

			return router
		}).
		Build()

	if err := serv.Run(ctx, cfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("service stopped with error: %w", err)
	}

	return nil
}
