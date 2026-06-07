package database

import (
	"time"

	"github.com/bimal009/Zovly/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewDB(cfg *config.Config) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.DB.URL)
	if err != nil {
		return nil, err
	}
	// Simple protocol avoids named/unnamed prepared statements, which fail on
	// PgBouncer in transaction mode (Neon pooler).
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	db := sqlx.NewDb(stdlib.OpenDB(*connConfig), "pgx")

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
