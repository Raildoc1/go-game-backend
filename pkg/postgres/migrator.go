package postgresstore

// Package: defines migrator type to run PostgreSQL migrations separate from storage generics.

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
)

// Migrator applies database migrations using provided configuration.
type Migrator struct {
	cfg *Config
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(cfg *Config) *Migrator {
	return &Migrator{cfg: cfg}
}

// Migrate applies migrations from the provided filesystem and directory.
func (m *Migrator) Migrate(migrations fs.FS, dir string) error {
	if migrations == nil {
		return nil
	}

	source, err := iofs.New(migrations, dir)
	if err != nil {
		return fmt.Errorf("iofs: %w", err)
	}

	db, err := sql.Open("pgx", m.cfg.DSN)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer func() { _ = db.Close() }()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
