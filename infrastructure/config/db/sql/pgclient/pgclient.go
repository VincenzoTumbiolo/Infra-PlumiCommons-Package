package pgclient

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PgClientConfig struct {
	MaxConnections  int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig return the default values found in the pgx library
func DefaultConfig() PgClientConfig {
	return PgClientConfig{
		MaxConnections:  4,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}
}

// New creates a connection pool from a data source name
func New(dsn string, config PgClientConfig) (*pgxpool.Pool, error) {
	configx, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing config from dsn: %w", err)
	}
	configx.MaxConns = int32(config.MaxConnections)
	configx.MaxConnLifetime = config.ConnMaxLifetime
	configx.MaxConnIdleTime = config.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), configx)
	if err != nil {
		return nil, fmt.Errorf("error creating pool: %w", err)
	}

	return pool, nil
}

// Migrate runs migrations with the provided source driver against the provided connection pool
func Migrate(migrationDriver source.Driver, pool *pgxpool.Pool) error {
	driver, err := pgxmigrate.WithInstance(stdlib.OpenDBFromPool(pool), &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("error creating pgx migration driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", migrationDriver, "pgx", driver)
	if err != nil {
		return fmt.Errorf("error istantiating migrator: %w", err)
	}

	if err = migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}
