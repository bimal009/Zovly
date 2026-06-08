package database

import (
	"context"
	"fmt"
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewDB(cfg *config.Config) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.DB.URL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	connConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	db := sqlx.NewDb(stdlib.OpenDB(*connConfig), "pgx")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)    // default: 25
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)    // default: 10
	db.SetConnMaxLifetime(cfg.DB.ConnLifetime) // default: 30min
	db.SetConnMaxIdleTime(cfg.DB.ConnIdleTime) // default: 5min

	return db, nil
}
