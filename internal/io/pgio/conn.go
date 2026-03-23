package pgio

import (
	"context"
	"fmt"
	"sync"

	"github.com/gnames/gnames/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbOnce sync.Once
var dbPool *pgxpool.Pool

// conn creates a pool of connections to PostgreSQL database.
func conn(cfg config.Config) (*pgio, error) {
	var err error

	pgxCfg, err := pgxpool.ParseConfig(opts(cfg))
	if err != nil {
		return nil, fmt.Errorf("pgio.conn: %w", err)
	}
	pgxCfg.MaxConns = 15

	dbOnce.Do(func() {
		dbPool, err = pgxpool.NewWithConfig(
			context.Background(),
			pgxCfg,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("pgio.conn: %w", err)
	}
	return &pgio{db: dbPool}, nil
}

// opts creates a string with options for connecting to PostgreSQL database.
func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PgHost, cfg.PgPort, cfg.PgUser, cfg.PgPass, cfg.PgDB)
}
