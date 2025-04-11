package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type DB struct {
	*sql.DB
	cfg *config.PostgresConfig
}

func New(cfg *config.PostgresConfig) (*DB, error) {
	log.Debug().
		Str("host", cfg.Host).
		Str("port", cfg.Port).
		Str("dbname", cfg.DB).
		Str("user", cfg.User).
		Msg("Connecting to PostgreSQL")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DB, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open database connection")
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.IdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionLifetime)

	log.Debug().
		Int("maxConnections", cfg.MaxConnections).
		Int("idleConnections", cfg.IdleConnections).
		Dur("connectionLifetime", cfg.ConnectionLifetime).
		Msg("Configured connection pool")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Info().Msg("Successfully connected to PostgreSQL")
	return &DB{
		DB:  db,
		cfg: cfg,
	}, nil
}

func (db *DB) Close() error {
	log.Debug().Msg("Closing database connection")
	return db.DB.Close()
}

func (db *DB) WithContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), db.cfg.QueryTimeout)
}

func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	log.Debug().Msg("Beginning database transaction")
	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
	}
	return tx, err
}

func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	log.Debug().Msg("Executing database transaction")
	startTime := time.Now()

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			log.Error().
				Interface("panic", p).
				Dur("duration", time.Since(startTime)).
				Msg("Transaction panicked, rolling back")
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		log.Debug().
			Err(err).
			Dur("duration", time.Since(startTime)).
			Msg("Transaction failed, rolling back")
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(startTime)).
			Msg("Failed to commit transaction")
		return err
	}

	log.Debug().
		Dur("duration", time.Since(startTime)).
		Msg("Transaction committed successfully")
	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	log.Debug().Msg("Pinging database")
	err := db.PingContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
	}
	return err
}
