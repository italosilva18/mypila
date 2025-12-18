package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"m2m-backend/database"
	"m2m-backend/models"
)

// CompanyRepository handles all database operations for companies
type CompanyRepository struct {
	collection *mongo.Collection
}

// NewCompanyRepository creates a new instance of CompanyRepository
func NewCompanyRepository() *CompanyRepository {
	return &CompanyRepository{
		collection: database.GetCollection("companies"),
	}
}

// FindByID retrieves a company by its ID
func (r *CompanyRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Company, error) {
	var company models.Company
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&company)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Company not found
		}
		return nil, NewRepositoryError("FindByID", err)
	}
	return &company, nil
}

// FindByUserID retrieves all companies owned by a specific user
func (r *CompanyRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Company, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, NewRepositoryError("FindByUserID", err)
	}
	defer cursor.Close(ctx)

	var companies []models.Company
	if err := cursor.All(ctx, &companies); err != nil {
		return nil, NewRepositoryError("FindByUserID:Decode", err)
	}

	// Return empty slice if no companies found
	if companies == nil {
		companies = []models.Company{}
	}

	return companies, nil
}

// Create inserts a new company into the database
func (r *CompanyRepository) Create(ctx context.Context, company *models.Company) error {
	// Ensure ID and timestamp are set
	if company.ID.IsZero() {
		company.ID = primitive.NewObjectID()
	}
	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, company)
	if err != nil {
		return NewRepositoryError("Create", err)
	}
	return nil
}

// Update modifies an existing company
func (r *CompanyRepository) Update(ctx context.Context, id primitive.ObjectID, name string) error {
	update := bson.M{
		"$set": bson.M{
			"name": name,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return NewRepositoryError("Update", err)
	}

	if result.MatchedCount == 0 {
		return nil // Company not found, but not an error
	}

	return nil
}

// Delete removes a company by its ID
func (r *CompanyRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return NewRepositoryError("Delete", err)
	}

	if result.DeletedCount == 0 {
		return nil // Company not found, but not an error
	}

	return nil
}

// DeleteRelatedTransactions removes all transactions associated with a company
func (r *CompanyRepository) DeleteRelatedTransactions(ctx context.Context, companyID primitive.ObjectID) error {
	transactionsCollection := database.GetCollection("transactions")
	_, err := transactionsCollection.DeleteMany(ctx, bson.M{"companyId": companyID})
	if err != nil {
		return NewRepositoryError("DeleteRelatedTransactions", err)
	}
	return nil
}

// DeleteRelatedCategories removes all categories associated with a company
func (r *CompanyRepository) DeleteRelatedCategories(ctx context.Context, companyIDStr string) error {
	categoriesCollection := database.GetCollection("categories")
	_, err := categoriesCollection.DeleteMany(ctx, bson.M{"companyId": companyIDStr})
	if err != nil {
		return NewRepositoryError("DeleteRelatedCategories", err)
	}
	return nil
}

// DeleteRelatedRecurring removes all recurring transactions associated with a company
func (r *CompanyRepository) DeleteRelatedRecurring(ctx context.Context, companyIDStr string) error {
	recurringCollection := database.GetCollection("recurring")
	_, err := recurringCollection.DeleteMany(ctx, bson.M{"companyId": companyIDStr})
	if err != nil {
		return NewRepositoryError("DeleteRelatedRecurring", err)
	}
	return nil
}

// ExistsByID checks if a company exists by its ID
func (r *CompanyRepository) ExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, NewRepositoryError("ExistsByID", err)
	}
	return count > 0, nil
}

// ValidateOwnership checks if a company belongs to a specific user
func (r *CompanyRepository) ValidateOwnership(ctx context.Context, companyID, userID primitive.ObjectID) (*models.Company, error) {
	var company models.Company
	err := r.collection.FindOne(ctx, bson.M{
		"_id":    companyID,
		"userId": userID,
	}).Decode(&company)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found or not owned by user
		}
		return nil, NewRepositoryError("ValidateOwnership", err)
	}

	return &company, nil
}
