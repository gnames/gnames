package pgio

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gnames/gnames/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbOnce sync.Once

// conn creates a pool of connections to PostgreSQL database.
func conn(cfg config.Config) (*pgio, error) {
	var db *pgxpool.Pool
	var err error

	pgxCfg, err := pgxpool.ParseConfig(opts(cfg))
	if err != nil {
		slog.Error("Cannot parse pgx config", "error", err)
		return nil, err
	}
	pgxCfg.MaxConns = 15

	dbOnce.Do(func() {
		db, err = pgxpool.NewWithConfig(
			context.Background(),
			pgxCfg,
		)
	})
	if err != nil {
		slog.Error("Cannot connect to database", "error", err)
		return nil, err
	}
	res := &pgio{db: db}
	return res, nil
}

// opts creates a string with options for connecting to PostgreSQL database.
func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PgHost, cfg.PgUser, cfg.PgPass, cfg.PgDB)
}
