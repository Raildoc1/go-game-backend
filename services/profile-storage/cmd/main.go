package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
	"log"
	"os"
	"time"

	pb "go-game-backend/gen/profilestorage"
	grpcserver "go-game-backend/services/profile-storage/internal/handlers/grpc"
	postgresrepo "go-game-backend/services/profile-storage/internal/repository/postgres"
	profilestorage "go-game-backend/services/profile-storage/internal/services/profilestorage"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// Config holds the configuration for the profile storage service.
type Config struct {
	Service         *service.Config           `yaml:"service"`
	GRPC            *service.GRPCServerConfig `yaml:"grpc"`
	Postgres        *postgresrepo.Config      `yaml:"postgres"`
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
		zap.String("service", "profile-storage"),
		zap.String("version", cfg.Service.Version),
	)

	if err = run(ctx, cfg, logger); err != nil {
		logger.ErrorCtx(ctx, "application stopped with error", zap.Error(err))
	}
	logger.InfoCtx(ctx, "application stopped successfully")
}

func run(ctx context.Context, cfg *Config, logger *logging.ZapLogger) error {
	repo, err := postgresrepo.New(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer service.Stop(ctx, repo, "postgres repository", logger)

	logicSvc := profilestorage.New(repo)
	srv := grpcserver.NewServer(logicSvc)

	serv := service.NewBuilder().
		WithGRPCServer(cfg.GRPC, func(s *grpc.Server) {
			pb.RegisterProfileStorageServiceServer(s, srv)
		}).
		Build()

	if err := serv.Run(ctx, cfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("service stopped with error: %w", err)
	}

	return nil
}
