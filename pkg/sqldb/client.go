package sqldb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	nrpgx "github.com/newrelic/go-agent/v3/integrations/nrpgx5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewClient initializes and returns a new database client based on the provided configuration.
func NewClient(envPrefix string) (*gorm.DB, error) {
	cfg, err := LoadConfig(envPrefix)
	if err != nil {
		return nil, err
	}
	dsn := getDSN(cfg)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewInstrumentedClient initializes and returns a new database client with New Relic instrumentation based on the provided configuration.
func NewInstrumentedClient(envPrefix string) (*gorm.DB, error) {
	cfg, err := LoadConfig(envPrefix)
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(getDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse pgx pool config: %w", err)
	}

	poolCfg.BeforeConnect = func(_ context.Context, connCfg *pgx.ConnConfig) error {
		connCfg.Tracer = nrpgx.NewTracer()
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("open pgx pool: %w", err)
	}

	return gorm.Open(postgres.New(postgres.Config{Conn: stdlib.OpenDBFromPool(pool)}), &gorm.Config{})
}

func getDSN(cfg *Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		cfg.TimeZone,
	)
}
