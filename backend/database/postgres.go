package database

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// Transaction error types for proper error handling
var (
	ErrTransactionAborted  = errors.New("transaction was aborted")
	ErrTransactionFailed   = errors.New("transaction failed to commit")
	ErrNoActiveTransaction = errors.New("no active transaction in context")
	ErrNotFound            = errors.New("record not found")
)

// Connect establishes a connection pool to PostgreSQL
func Connect() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mypila_user:mypila_secret_2024@localhost:5434/mypila?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return err
	}

	// Configure pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	// Ping the database
	if err := Pool.Ping(ctx); err != nil {
		return err
	}

	log.Println("Connected to PostgreSQL!")
	return nil
}

// ConnectWithIndexes connects to PostgreSQL (indexes are created via init.sql)
func ConnectWithIndexes() error {
	return Connect()
}

// Disconnect closes the PostgreSQL connection pool gracefully
func Disconnect() {
	if Pool != nil {
		Pool.Close()
		log.Println("Disconnected from PostgreSQL")
	}
}

// GetPool returns the connection pool
func GetPool() *pgxpool.Pool {
	return Pool
}

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	MaxRetries int
	Timeout    time.Duration
	IsoLevel   pgx.TxIsoLevel
}

// DefaultTransactionOptions returns sensible defaults for transactions
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		MaxRetries: 3,
		Timeout:    30 * time.Second,
		IsoLevel:   pgx.ReadCommitted,
	}
}

// WithTransaction executes a function within a PostgreSQL transaction.
// It handles transaction management, retries on transient errors, and proper cleanup.
func WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return WithTransactionOptions(ctx, DefaultTransactionOptions(), fn)
}

// WithTransactionOptions executes a function within a PostgreSQL transaction with custom options.
func WithTransactionOptions(ctx context.Context, opts TransactionOptions, fn func(tx pgx.Tx) error) error {
	if Pool == nil {
		return errors.New("database pool not initialized")
	}

	txCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[TRANSACTION] Retrying transaction, attempt %d/%d", attempt, opts.MaxRetries)
		}

		lastErr = executeTransaction(txCtx, opts.IsoLevel, fn)
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr
		}

		select {
		case <-txCtx.Done():
			return txCtx.Err()
		case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
		}
	}

	return lastErr
}

func executeTransaction(ctx context.Context, isoLevel pgx.TxIsoLevel, fn func(tx pgx.Tx) error) error {
	tx, err := Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: isoLevel})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL serialization failures can be retried
	errStr := err.Error()
	return contains(errStr, "serialization_failure") ||
		contains(errStr, "deadlock_detected") ||
		contains(errStr, "40001") || // serialization_failure
		contains(errStr, "40P01") // deadlock_detected
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RunInTransaction is a convenience wrapper that executes multiple operations atomically.
func RunInTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return WithTransaction(ctx, fn)
}

// QueryRow executes a query that returns a single row
func QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return Pool.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns multiple rows
func Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return Pool.Query(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows
func Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := Pool.Exec(ctx, sql, args...)
	return err
}
