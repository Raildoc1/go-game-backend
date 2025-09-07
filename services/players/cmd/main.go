// Package main starts the players service.
package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/kafka"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
	"go-game-backend/services/players/internal/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	postgresstore "go-game-backend/pkg/postgres"
	redisstore "go-game-backend/pkg/redis"

	statehand "go-game-backend/services/players/internal/handlers/state"
	playerkafka "go-game-backend/services/players/internal/ingester/kafka"

	postgresrepo "go-game-backend/services/players/internal/repository/postgres"
	redisrepo "go-game-backend/services/players/internal/repository/redis"
	playersvc "go-game-backend/services/players/internal/services/player"

	playerslocker "go-game-backend/services/players/pkg/locker"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the players service.
type Config struct {
	Service         *service.Config           `yaml:"service"`
	HTTP            *service.HTTPServerConfig `yaml:"http"`
	PlayerService   *playersvc.Config         `yaml:"player-service"`
	Redis           *redisstore.Config        `yaml:"redis"`
	Postgres        *postgresstore.Config     `yaml:"postgres"`
	Kafka           *kafka.ReaderConfig       `yaml:"kafka"`
	JWT             *middleware.JWTConfig     `yaml:"jwt"`
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
		os.Exit(1)
	}
	logger.InfoCtx(ctx, "application stopped successfully")
}

func run(ctx context.Context, cfg *Config, logger *logging.ZapLogger) error {
	rxStorage := redisstore.New(cfg.Redis, logger, redisrepo.NewRepos)
	defer service.Stop(ctx, rxStorage, "redis storage", logger)
	rxStore := redisrepo.NewStore(rxStorage)

	pgStorage, err := postgresstore.New(ctx, cfg.Postgres, postgresrepo.NewRepos)
	if err != nil {
		return fmt.Errorf("create postgres storage: %w", err)
	}
	defer service.Stop(ctx, pgStorage, "postgres storage", logger)
	pgStore := postgresrepo.NewStore(pgStorage)

	locker := playerslocker.NewFromStorage(rxStorage, cfg.PlayerService.PlayerLockTTL)
	playerService := playersvc.New(cfg.PlayerService, pgStore)

	validator := middleware.NewSessionValidator(cfg.JWT, rxStore.Raw().Session(), logger)
	lockMw := middleware.NewPlayerLock(locker)

	stateHTTPHandler := statehand.New(playerService, logger)

	reader := kafka.NewReader(cfg.Kafka)
	defer service.Close(ctx, reader, "kafka reader", logger)
	ing := playerkafka.NewUserCreated(reader, logger, playerService, hub, locker)

	serv := service.NewBuilder().
		WithGo(func(ctx context.Context) error {
			if err := ing.Run(ctx); err != nil {
				return fmt.Errorf("kafka user created reader: %w", err)
			}
			return nil
		}).
		WithHTTPServer(cfg.HTTP, func() http.Handler {
			router := gin.Default()

			api := router.Group("/api/v1", validator.Middleware())
			{
				playerAPI := api.Group("/player", lockMw.Middleware())
				{
					stateHTTPHandler.Register(playerAPI.Group("/state"))
				}
			}

			return router
		}).
		Build()

	if err := serv.Run(ctx, cfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("service stopped with error: %w", err)
	}

	return nil
}
