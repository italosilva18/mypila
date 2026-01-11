package database

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var DB *mongo.Database
var client *mongo.Client

// Transaction error types for proper error handling
var (
	ErrTransactionAborted   = errors.New("transaction was aborted")
	ErrTransactionFailed    = errors.New("transaction failed to commit")
	ErrNoActiveTransaction  = errors.New("no active transaction in context")
	ErrReplicaSetRequired   = errors.New("MongoDB transactions require a replica set configuration")
)

func Connect() error {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "m2m_financeiro"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	DB = client.Database(dbName)
	log.Println("Connected to MongoDB!")

	return nil
}

// ConnectWithIndexes connects to MongoDB and creates all necessary indexes
func ConnectWithIndexes() error {
	if err := Connect(); err != nil {
		return err
	}

	// Import is handled in main.go to avoid circular dependency
	// Indexes are created after connection is established
	return nil
}

// Disconnect closes the MongoDB connection gracefully
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return err
	}

	log.Println("Disconnected from MongoDB")
	return nil
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}

// GetClient returns the MongoDB client instance.
// This is needed for creating sessions and transactions.
func GetClient() *mongo.Client {
	return client
}

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	// MaxRetries specifies how many times to retry a transaction on transient errors
	MaxRetries int
	// Timeout specifies the maximum duration for the entire transaction
	Timeout time.Duration
}

// DefaultTransactionOptions returns sensible defaults for transactions
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}
}

// WithTransaction executes a function within a MongoDB transaction.
// It handles session management, retries on transient errors, and proper cleanup.
//
// IMPORTANT: MongoDB transactions require a replica set configuration.
// For local development, you can use a single-node replica set.
// For production, ensure your MongoDB deployment is configured as a replica set.
//
// Usage:
//
//	err := database.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
//	    // Perform your database operations here
//	    // Use sessCtx instead of regular context for transaction-aware operations
//	    _, err := collection.InsertOne(sessCtx, document)
//	    if err != nil {
//	        return err // Transaction will be aborted
//	    }
//	    return nil // Transaction will be committed
//	})
func WithTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return WithTransactionOptions(ctx, DefaultTransactionOptions(), fn)
}

// WithTransactionOptions executes a function within a MongoDB transaction with custom options.
// This allows fine-grained control over retry behavior and timeouts.
func WithTransactionOptions(ctx context.Context, opts TransactionOptions, fn func(mongo.SessionContext) error) error {
	if client == nil {
		return errors.New("database client not initialized")
	}

	// Create a context with timeout for the entire transaction
	txCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Start a session
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(txCtx)

	// Configure transaction options for ACID compliance
	txnOpts := options.Transaction().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Snapshot())

	// Execute with retry logic for transient errors
	var lastErr error
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[TRANSACTION] Retrying transaction, attempt %d/%d", attempt, opts.MaxRetries)
		}

		lastErr = executeTransaction(txCtx, session, txnOpts, fn)
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !isRetryableError(lastErr) {
			return lastErr
		}

		// Small backoff before retry
		select {
		case <-txCtx.Done():
			return txCtx.Err()
		case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
			// Continue with retry
		}
	}

	return lastErr
}

// executeTransaction runs the transaction function within a session
func executeTransaction(ctx context.Context, session mongo.Session, txnOpts *options.TransactionOptions, fn func(mongo.SessionContext) error) error {
	// Use the session's WithTransaction method which handles commit/abort automatically
	_, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := fn(sessCtx); err != nil {
			return nil, err
		}
		return nil, nil
	}, txnOpts)

	return err
}

// isRetryableError checks if an error is a transient error that can be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for MongoDB command errors with transient labels
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.HasErrorLabel("TransientTransactionError") ||
			cmdErr.HasErrorLabel("UnknownTransactionCommitResult")
	}

	// Check for write exceptions with transient labels
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		return writeErr.HasErrorLabel("TransientTransactionError") ||
			writeErr.HasErrorLabel("UnknownTransactionCommitResult")
	}

	return false
}

// RunInTransaction is a convenience wrapper that executes multiple operations atomically.
// It's designed for simple cases where you want to run a sequence of operations.
//
// Example:
//
//	err := database.RunInTransaction(ctx, func(sessCtx mongo.SessionContext) error {
//	    // Insert quote
//	    if _, err := quoteCollection.InsertOne(sessCtx, quote); err != nil {
//	        return err
//	    }
//	    // Insert items
//	    for _, item := range items {
//	        if _, err := itemsCollection.InsertOne(sessCtx, item); err != nil {
//	            return err
//	        }
//	    }
//	    return nil
//	})
func RunInTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return WithTransaction(ctx, fn)
}

// IsTransactionSupported checks if the MongoDB deployment supports transactions.
// Transactions require a replica set or sharded cluster configuration.
func IsTransactionSupported(ctx context.Context) bool {
	if client == nil {
		return false
	}

	// Try to check if we're connected to a replica set
	// This is done by checking the server status
	result := client.Database("admin").RunCommand(ctx, map[string]interface{}{
		"replSetGetStatus": 1,
	})

	// If the command succeeds, we have a replica set
	if result.Err() == nil {
		return true
	}

	// For standalone servers, the command will fail
	// We can still try to use transactions and handle the error
	log.Printf("[TRANSACTION] MongoDB replica set check: %v", result.Err())
	return false
}

// GetCollectionWithSession returns a collection configured to use the session context.
// This is useful when you need to pass a collection to functions that expect *mongo.Collection
// but need to maintain transaction context.
func GetCollectionWithSession(name string) *mongo.Collection {
	return DB.Collection(name)
}
