package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
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
	serv := service.NewBuilder().
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
