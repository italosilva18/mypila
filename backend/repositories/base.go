package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseRepository defines common database operations
// This interface can be extended by specific repositories
type BaseRepository interface {
	// FindByID retrieves a document by its ID
	FindByID(ctx context.Context, id primitive.ObjectID, result interface{}) error

	// FindAll retrieves all documents matching a filter
	FindAll(ctx context.Context, filter interface{}, results interface{}) error

	// Create inserts a new document
	Create(ctx context.Context, document interface{}) error

	// Update modifies an existing document
	Update(ctx context.Context, id primitive.ObjectID, update interface{}) error

	// Delete removes a document by its ID
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Count returns the number of documents matching a filter
	Count(ctx context.Context, filter interface{}) (int64, error)
}

// RepositoryError represents a repository-level error
type RepositoryError struct {
	Operation string
	Err       error
}

func (e *RepositoryError) Error() string {
	return e.Operation + ": " + e.Err.Error()
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Err:       err,
	}
}
