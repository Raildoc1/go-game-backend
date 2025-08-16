// Package postgresstore provides PostgreSQL storage with migrations support.
package postgresstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
)

// Config holds PostgreSQL connection settings.
type Config struct {
	DSN string `yaml:"dsn"`
}

// Storage manages PostgreSQL connection pool and migrations.
type Storage struct {
	cfg  *Config
	pool *pgxpool.Pool
}

// New creates a new Storage with the given configuration.
func New(ctx context.Context, cfg *Config) (*Storage, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgx pool connect: %w", err)
	}
	return &Storage{cfg: cfg, pool: pool}, nil
}

// Pool exposes underlying pgx pool.
func (s *Storage) Pool() *pgxpool.Pool {
	return s.pool
}

// Stop closes the database connection pool.
func (s *Storage) Stop() error {
	s.pool.Close()
	return nil
}

// Migrate applies migrations from the provided filesystem and directory.
func (s *Storage) Migrate(migrations fs.FS, dir string) error {
	if migrations == nil {
		return nil
	}

	source, err := iofs.New(migrations, dir)
	if err != nil {
		return fmt.Errorf("iofs: %w", err)
	}

	db, err := sql.Open("pgx", s.cfg.DSN)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer func() { _ = db.Close() }()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
