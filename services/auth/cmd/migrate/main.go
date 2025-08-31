// Command auth-migrate runs database migrations for the auth service.
package main

import (
	"context"
	"fmt"
	"go-game-backend/pkg/logging"
	"go-game-backend/pkg/service"
	"go-game-backend/services/auth/migrations"
	"log"
	"os"

	postgresstore "go-game-backend/pkg/postgres"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds service and PostgreSQL configuration for the migration.
type Config struct {
	Service  *service.Config       `yaml:"service"`
	Postgres *postgresstore.Config `yaml:"postgres"`
}

func main() {
	cfg, err := service.LoadConfig[Config](
		"./configs/default.yaml",
		os.Stdout,
		func(err error) { log.Fatal(err) },
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
		zap.String("service", "auth-migrate"),
		zap.String("version", cfg.Service.Version),
	)

	if err = run(ctx, cfg, logger); err != nil {
		logger.ErrorCtx(ctx, "migration failed", zap.Error(err))
		os.Exit(1)
	}
	logger.InfoCtx(ctx, "migration completed")
}

// run executes the database migrations.
func run(ctx context.Context, cfg *Config, logger *logging.ZapLogger) error {
	migrator := postgresstore.NewMigrator(cfg.Postgres)
	if err := migrator.Migrate(migrations.Files, "."); err != nil {
		return fmt.Errorf("storage migration: %w", err)
	}
	return nil
}
