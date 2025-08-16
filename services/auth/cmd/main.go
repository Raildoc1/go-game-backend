package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/jwtfactory"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"

	redisstore "go-game-backend/pkg/redis"

	profilestoragegrpc "go-game-backend/services/auth/internal/gateway/profilestorage/grpc"
	httphand "go-game-backend/services/auth/internal/handlers/http"
	redisrepo "go-game-backend/services/auth/internal/repository/redis"
	authserv "go-game-backend/services/auth/internal/services/auth"
	tknfactory "go-game-backend/services/auth/internal/services/token"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the auth service.
type Config struct {
	// Service contains generic service configuration such as version.
	Service *service.Config `yaml:"service"`
	// HTTP defines settings for the HTTP server.
	HTTP *service.HTTPServerConfig `yaml:"http"`
	// ShutdownTimeout specifies how long to wait for graceful shutdown.
	AuthService        *authserv.Config           `yaml:"auth-service"`
	Redis              *redisstore.Config         `yaml:"redis"`
	TokenFactory       *tknfactory.Config         `yaml:"token-factory"`
	ProfileStorageGRPC *profilestoragegrpc.Config `yaml:"profile-storage-grpc"`
	JWTConfig          *JWTConfig                 `yaml:"jwt"`
	ShutdownTimeout    time.Duration              `yaml:"shutdown-timeout"`
}

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

	profileStorageGateway, err := profilestoragegrpc.New(cfg.ProfileStorageGRPC)
	if err != nil {
		return fmt.Errorf("failed to create profile storage gateway: %w", err)
	}
	defer service.Stop(ctx, profileStorageGateway, "profile storage gateway", logger)

	authService := authserv.New(cfg.AuthService, profileStorageGateway, redisRepo, tknFactory)
	httpHandler := httphand.New(authService, logger)

	serv := service.NewBuilder().
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
