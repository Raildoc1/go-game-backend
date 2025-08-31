// Package main starts the authentication service.
package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/jwtfactory"
	"go-game-backend/pkg/kafka"
	"go-game-backend/pkg/logging"
	outboxpkg "go-game-backend/pkg/outbox"
	"go-game-backend/pkg/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"

	postgresstore "go-game-backend/pkg/postgres"
	redisstore "go-game-backend/pkg/redis"
	httphand "go-game-backend/services/auth/internal/handlers/http"
	postgresrepo "go-game-backend/services/auth/internal/repository/postgres"
	redisrepo "go-game-backend/services/auth/internal/repository/redis"
	authserv "go-game-backend/services/auth/internal/services/auth"
	tknfactory "go-game-backend/services/auth/internal/services/token"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the auth service.
type Config struct {
	Service         *service.Config           `yaml:"service"`
	HTTP            *service.HTTPServerConfig `yaml:"http"`
	AuthService     *authserv.Config          `yaml:"auth-service"`
	Redis           *redisstore.Config        `yaml:"redis"`
	TokenFactory    *tknfactory.Config        `yaml:"token-factory"`
	Postgres        *postgresstore.Config     `yaml:"postgres"`
	Kafka           *kafka.ForwarderConfig    `yaml:"kafka"`
	JWTConfig       *JWTConfig                `yaml:"jwt"`
	ShutdownTimeout time.Duration             `yaml:"shutdown-timeout"`
}

// JWTConfig holds the configuration for the JWT generation.
type JWTConfig struct {
	Algorithm string `yaml:"algorithm"`
	Secret    string `yaml:"secret"`
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
		zap.String("service", "auth"),
		zap.String("version", cfg.Service.Version),
	)

	if err = run(ctx, cfg, logger); err != nil {
		logger.ErrorCtx(ctx, "application stopped with error", zap.Error(err))
	}
	logger.InfoCtx(ctx, "application stopped successfully")
}

func run(ctx context.Context, cfg *Config, logger *logging.ZapLogger) error {
	tokenAuth := jwtauth.New(cfg.JWTConfig.Algorithm, []byte(cfg.JWTConfig.Secret), nil)
	jwtFactory := jwtfactory.New(tokenAuth)
	tknFactory := tknfactory.New(jwtFactory, cfg.TokenFactory)

	redisStore := redisstore.New(cfg.Redis, logger)
	defer service.Stop(ctx, redisStore, "redis storage", logger)

	redisRepo := redisrepo.New(redisStore)

	storage, err := postgresstore.NewStorage(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer service.Stop(ctx, storage, "postgres storage", logger)

	repo := postgresrepo.New(storage.Pool(), logger)
	defer service.Stop(ctx, repo, "postgres repository", logger)

	outboxStore := outboxpkg.NewRepository(storage.Pool())
	writer := kafka.NewWriter(cfg.Kafka.Brokers)
	defer writer.Close()
	forwarder := outboxpkg.NewForwarder(outboxStore, writer, cfg.Kafka.PollInterval, cfg.Kafka.BatchSize)

	authService := authserv.New(cfg.AuthService, repo, redisRepo, tknFactory, outboxStore)
	httpHandler := httphand.New(authService, logger)

	serv := service.NewBuilder().
		WithGo(func(ctx context.Context) error {
			forwarder.Run(ctx)
			return nil
		}).
		WithHTTPServer(cfg.HTTP, func() http.Handler {
			router := gin.Default()

			api := router.Group("/api/v1")
			{
				api.POST("/login", httpHandler.Login)
				api.POST("/register", httpHandler.Register)
				api.POST("/refresh", httpHandler.RefreshToken)
			}

			return router
		}).
		Build()

	if err := serv.Run(ctx, cfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("service stopped with error: %w", err)
	}

	return nil
}
