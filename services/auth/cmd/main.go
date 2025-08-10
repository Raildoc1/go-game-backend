package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"go-game-backend/pkg/jwtfactory"
	"go-game-backend/pkg/logging"
	redisstore "go-game-backend/pkg/redis"
	"go-game-backend/pkg/service"
	httphand "go-game-backend/services/auth/internal/handlers/http"
	postgresrepo "go-game-backend/services/auth/internal/repository/postgres"
	redisrepo "go-game-backend/services/auth/internal/repository/redis"
	authserv "go-game-backend/services/auth/internal/services/auth"
	tknfactory "go-game-backend/services/auth/internal/services/token"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Service         *service.Config           `yaml:"service"`
	HTTP            *service.HTTPServerConfig `yaml:"http"`
	AuthService     *authserv.Config          `yaml:"auth-service"`
	Redis           *redisstore.Config        `yaml:"redis"`
	TokenFactory    *tknfactory.Config        `yaml:"token-factory"`
	JWTConfig       *JWTConfig                `yaml:"jwt"`
	ShutdownTimeout time.Duration             `yaml:"shutdown-timeout"`
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
	redisRepo := redisrepo.New(redisStore)
	postgresRepo := postgresrepo.New()
	authService := authserv.New(cfg.AuthService, postgresRepo, redisRepo, tknFactory)
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
