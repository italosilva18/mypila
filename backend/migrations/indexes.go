package migrations

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes creates all necessary MongoDB indexes for optimal query performance
func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Users collection indexes
	if err := createUsersIndexes(ctx, db); err != nil {
		return err
	}

	// Companies collection indexes
	if err := createCompaniesIndexes(ctx, db); err != nil {
		return err
	}

	// Transactions collection indexes
	if err := createTransactionsIndexes(ctx, db); err != nil {
		return err
	}

	// Categories collection indexes
	if err := createCategoriesIndexes(ctx, db); err != nil {
		return err
	}

	// Recurring transactions collection indexes
	if err := createRecurringIndexes(ctx, db); err != nil {
		return err
	}

	// Refresh tokens collection indexes
	if err := createRefreshTokensIndexes(ctx, db); err != nil {
		return err
	}

	log.Println("All database indexes created successfully")
	return nil
}

// createUsersIndexes creates indexes for the users collection
func createUsersIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Unique index on email for fast user lookup and authentication
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("email_unique_idx"),
	}

	_, err := collection.Indexes().CreateOne(ctx, emailIndex)
	if err != nil {
		log.Printf("Warning: Failed to create email index on users: %v", err)
		return err
	}

	log.Println("Created index: users.email_unique_idx")
	return nil
}

// createCompaniesIndexes creates indexes for the companies collection
func createCompaniesIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("companies")

	// Index on userId for fast company lookup by user
	userIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetName("userId_idx"),
	}

	_, err := collection.Indexes().CreateOne(ctx, userIdIndex)
	if err != nil {
		log.Printf("Warning: Failed to create userId index on companies: %v", err)
		return err
	}

	log.Println("Created index: companies.userId_idx")
	return nil
}

// createTransactionsIndexes creates indexes for the transactions collection
func createTransactionsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("transactions")

	// Index on companyId for fast transaction lookup by company
	companyIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "companyId", Value: 1}},
		Options: options.Index().SetName("companyId_idx"),
	}

	// Compound index on companyId + year + month for efficient period queries
	// This index supports queries filtering by company and date range
	periodIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "year", Value: -1},      // Descending for recent first
			{Key: "month", Value: 1},
		},
		Options: options.Index().SetName("companyId_year_month_idx"),
	}

	// Index on status for filtering by transaction status
	statusIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}},
		Options: options.Index().SetName("status_idx"),
	}

	// Compound index for company + status queries (common use case)
	companyStatusIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "status", Value: 1},
		},
		Options: options.Index().SetName("companyId_status_idx"),
	}

	indexes := []mongo.IndexModel{
		companyIdIndex,
		periodIndex,
		statusIndex,
		companyStatusIndex,
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes on transactions: %v", err)
		return err
	}

	log.Println("Created indexes: transactions.companyId_idx, companyId_year_month_idx, status_idx, companyId_status_idx")
	return nil
}

// createCategoriesIndexes creates indexes for the categories collection
func createCategoriesIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("categories")

	// Index on companyId for fast category lookup by company
	companyIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "companyId", Value: 1}},
		Options: options.Index().SetName("companyId_idx"),
	}

	// Compound index on companyId + type for filtering categories by type
	companyTypeIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetName("companyId_type_idx"),
	}

	indexes := []mongo.IndexModel{
		companyIdIndex,
		companyTypeIndex,
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes on categories: %v", err)
		return err
	}

	log.Println("Created indexes: categories.companyId_idx, companyId_type_idx")
	return nil
}

// createRecurringIndexes creates indexes for the recurring transactions collection
func createRecurringIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("recurring")

	// Index on companyId for fast recurring transaction lookup by company
	companyIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "companyId", Value: 1}},
		Options: options.Index().SetName("companyId_idx"),
	}

	// Compound index on companyId + dayOfMonth for scheduled processing
	companyDayIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "dayOfMonth", Value: 1},
		},
		Options: options.Index().SetName("companyId_dayOfMonth_idx"),
	}

	indexes := []mongo.IndexModel{
		companyIdIndex,
		companyDayIndex,
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes on recurring: %v", err)
		return err
	}

	log.Println("Created indexes: recurring.companyId_idx, companyId_dayOfMonth_idx")
	return nil
}

// createRefreshTokensIndexes creates indexes for the refresh_tokens collection
func createRefreshTokensIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("refresh_tokens")

	// Unique index on tokenHash for fast token lookup
	// This is the primary lookup method when validating refresh tokens
	tokenHashIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "tokenHash", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("tokenHash_unique_idx"),
	}

	// Index on userId for finding all tokens for a user (logout-all functionality)
	userIdIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetName("userId_idx"),
	}

	// Compound index on userId + isRevoked for efficient queries
	// Used when revoking all active tokens for a user
	userIdRevokedIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "isRevoked", Value: 1},
		},
		Options: options.Index().SetName("userId_isRevoked_idx"),
	}

	// TTL index on expiresAt for automatic cleanup of expired tokens
	// MongoDB will automatically delete documents when expiresAt is past
	expiresAtIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0).SetName("expiresAt_ttl_idx"),
	}

	indexes := []mongo.IndexModel{
		tokenHashIndex,
		userIdIndex,
		userIdRevokedIndex,
		expiresAtIndex,
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes on refresh_tokens: %v", err)
		return err
	}

	log.Println("Created indexes: refresh_tokens.tokenHash_unique_idx, userId_idx, userId_isRevoked_idx, expiresAt_ttl_idx")
	return nil
}
