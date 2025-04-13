package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type TxManager interface {
	RunTransaction(ctx context.Context, fn func(*sql.Tx) error) error
}

type DBTxManager struct {
	db *DB
}

func NewTxManager(db *DB) *DBTxManager {
	return &DBTxManager{
		db: db,
	}
}

func (tm *DBTxManager) RunTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	log.Debug().Msg("Starting database transaction")
	startTime := time.Now()

	tx, err := tm.db.BeginTx(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
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
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Debug().
		Dur("duration", time.Since(startTime)).
		Msg("Transaction committed successfully")
	return nil
}
