package migrations

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestCreateIndexes verifies that all indexes are created correctly
func TestCreateIndexes(t *testing.T) {
	// Skip if no MongoDB connection is available
	t.Skip("Integration test - requires MongoDB connection")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to test database
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	testDB := client.Database("test_indexes")

	// Clean up test database after test
	defer testDB.Drop(ctx)

	// Create indexes
	if err := CreateIndexes(testDB); err != nil {
		t.Fatalf("Failed to create indexes: %v", err)
	}

	// Verify users collection indexes
	t.Run("UsersIndexes", func(t *testing.T) {
		indexes := testDB.Collection("users").Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list users indexes: %v", err)
		}
		defer cursor.Close(ctx)

		var indexFound bool
		for cursor.Next(ctx) {
			var index bson.M
			if err := cursor.Decode(&index); err != nil {
				t.Fatalf("Failed to decode index: %v", err)
			}

			if name, ok := index["name"].(string); ok && name == "email_unique_idx" {
				indexFound = true

				// Verify unique constraint
				if unique, ok := index["unique"].(bool); !ok || !unique {
					t.Error("email_unique_idx should be unique")
				}
			}
		}

		if !indexFound {
			t.Error("email_unique_idx not found in users collection")
		}
	})

	// Verify companies collection indexes
	t.Run("CompaniesIndexes", func(t *testing.T) {
		indexes := testDB.Collection("companies").Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list companies indexes: %v", err)
		}
		defer cursor.Close(ctx)

		var indexFound bool
		for cursor.Next(ctx) {
			var index bson.M
			if err := cursor.Decode(&index); err != nil {
				t.Fatalf("Failed to decode index: %v", err)
			}

			if name, ok := index["name"].(string); ok && name == "userId_idx" {
				indexFound = true
			}
		}

		if !indexFound {
			t.Error("userId_idx not found in companies collection")
		}
	})

	// Verify transactions collection indexes
	t.Run("TransactionsIndexes", func(t *testing.T) {
		indexes := testDB.Collection("transactions").Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list transactions indexes: %v", err)
		}
		defer cursor.Close(ctx)

		expectedIndexes := map[string]bool{
			"companyId_idx":            false,
			"companyId_year_month_idx": false,
			"status_idx":               false,
			"companyId_status_idx":     false,
		}

		for cursor.Next(ctx) {
			var index bson.M
			if err := cursor.Decode(&index); err != nil {
				t.Fatalf("Failed to decode index: %v", err)
			}

			if name, ok := index["name"].(string); ok {
				if _, exists := expectedIndexes[name]; exists {
					expectedIndexes[name] = true
				}
			}
		}

		for name, found := range expectedIndexes {
			if !found {
				t.Errorf("Index %s not found in transactions collection", name)
			}
		}
	})

	// Verify categories collection indexes
	t.Run("CategoriesIndexes", func(t *testing.T) {
		indexes := testDB.Collection("categories").Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list categories indexes: %v", err)
		}
		defer cursor.Close(ctx)

		expectedIndexes := map[string]bool{
			"companyId_idx":      false,
			"companyId_type_idx": false,
		}

		for cursor.Next(ctx) {
			var index bson.M
			if err := cursor.Decode(&index); err != nil {
				t.Fatalf("Failed to decode index: %v", err)
			}

			if name, ok := index["name"].(string); ok {
				if _, exists := expectedIndexes[name]; exists {
					expectedIndexes[name] = true
				}
			}
		}

		for name, found := range expectedIndexes {
			if !found {
				t.Errorf("Index %s not found in categories collection", name)
			}
		}
	})

	// Verify recurring collection indexes
	t.Run("RecurringIndexes", func(t *testing.T) {
		indexes := testDB.Collection("recurring").Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			t.Fatalf("Failed to list recurring indexes: %v", err)
		}
		defer cursor.Close(ctx)

		expectedIndexes := map[string]bool{
			"companyId_idx":            false,
			"companyId_dayOfMonth_idx": false,
		}

		for cursor.Next(ctx) {
			var index bson.M
			if err := cursor.Decode(&index); err != nil {
				t.Fatalf("Failed to decode index: %v", err)
			}

			if name, ok := index["name"].(string); ok {
				if _, exists := expectedIndexes[name]; exists {
					expectedIndexes[name] = true
				}
			}
		}

		for name, found := range expectedIndexes {
			if !found {
				t.Errorf("Index %s not found in recurring collection", name)
			}
		}
	})
}

// TestIndexCreationIdempotency verifies that creating indexes multiple times doesn't cause errors
func TestIndexCreationIdempotency(t *testing.T) {
	t.Skip("Integration test - requires MongoDB connection")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	testDB := client.Database("test_indexes_idempotency")
	defer testDB.Drop(ctx)

	// Create indexes first time
	if err := CreateIndexes(testDB); err != nil {
		t.Fatalf("Failed to create indexes first time: %v", err)
	}

	// Create indexes second time (should not error)
	if err := CreateIndexes(testDB); err != nil {
		t.Fatalf("Failed to create indexes second time: %v", err)
	}

	// Create indexes third time (should not error)
	if err := CreateIndexes(testDB); err != nil {
		t.Fatalf("Failed to create indexes third time: %v", err)
	}
}

// BenchmarkIndexCreation benchmarks the index creation process
func BenchmarkIndexCreation(b *testing.B) {
	b.Skip("Benchmark test - requires MongoDB connection")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	for i := 0; i < b.N; i++ {
		testDB := client.Database("benchmark_indexes")
		defer testDB.Drop(ctx)

		b.StartTimer()
		if err := CreateIndexes(testDB); err != nil {
			b.Fatalf("Failed to create indexes: %v", err)
		}
		b.StopTimer()
	}
}
