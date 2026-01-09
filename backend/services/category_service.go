package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

var (
	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")
)

// CategoryService contains business logic for category operations
type CategoryService struct {
	collection *mongo.Collection
}

// NewCategoryService creates a new instance of CategoryService
func NewCategoryService() *CategoryService {
	return &CategoryService{
		collection: database.GetCollection("categories"),
	}
}

// GetCollection returns the categories collection
func (s *CategoryService) GetCollection() *mongo.Collection {
	return s.collection
}

// GetByCompanyID retrieves all categories for a company
func (s *CategoryService) GetByCompanyID(ctx context.Context, companyID primitive.ObjectID) ([]models.Category, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"companyId": companyID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []models.Category{}
	}

	return categories, nil
}

// GetByID retrieves a category by its ID
func (s *CategoryService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var category models.Category
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

// Create creates a new category with validation
func (s *CategoryService) Create(ctx context.Context, companyID primitive.ObjectID, req models.CreateCategoryRequest) (*models.Category, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user input
	sanitizedName := helpers.SanitizeString(req.Name)

	// Default to EXPENSE if not provided
	categoryType := req.Type
	if categoryType == "" {
		categoryType = models.Expense
	}

	category := &models.Category{
		ID:        primitive.NewObjectID(),
		CompanyID: companyID,
		Name:      sanitizedName,
		Type:      categoryType,
		Color:     req.Color,
		Budget:    req.Budget,
		CreatedAt: time.Now(),
	}

	_, err := s.collection.InsertOne(ctx, category)
	if err != nil {
		return nil, nil, err
	}

	return category, nil, nil
}

// Update updates an existing category with validation
func (s *CategoryService) Update(ctx context.Context, id primitive.ObjectID, req models.UpdateCategoryRequest) (*models.Category, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user input
	sanitizedName := helpers.SanitizeString(req.Name)

	update := bson.M{
		"$set": bson.M{
			"name":   sanitizedName,
			"type":   req.Type,
			"color":  req.Color,
			"budget": req.Budget,
		},
	}

	result, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, nil, err
	}

	if result.MatchedCount == 0 {
		return nil, nil, ErrCategoryNotFound
	}

	// Fetch updated category
	var category models.Category
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		return nil, nil, err
	}

	return &category, nil, nil
}

// Delete removes a category by its ID
func (s *CategoryService) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// ValidateOwnership checks if a category belongs to a user's company
func (s *CategoryService) ValidateOwnership(ctx context.Context, categoryID primitive.ObjectID, userCompanyIDs []primitive.ObjectID) (*models.Category, error) {
	category, err := s.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// Check if category's company is in user's companies
	for _, companyID := range userCompanyIDs {
		if category.CompanyID == companyID {
			return category, nil
		}
	}

	return nil, ErrUnauthorized
}

// DeleteByCompanyID removes all categories for a company (cascade delete)
func (s *CategoryService) DeleteByCompanyID(ctx context.Context, companyID primitive.ObjectID) error {
	_, err := s.collection.DeleteMany(ctx, bson.M{"companyId": companyID})
	return err
}

// GetByType retrieves categories filtered by type (EXPENSE or INCOME)
func (s *CategoryService) GetByType(ctx context.Context, companyID primitive.ObjectID, categoryType models.CategoryType) ([]models.Category, error) {
	filter := bson.M{
		"companyId": companyID,
		"type":      categoryType,
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []models.Category{}
	}

	return categories, nil
}

// ExistsByName checks if a category with the given name exists for a company
func (s *CategoryService) ExistsByName(ctx context.Context, companyID primitive.ObjectID, name string) (bool, error) {
	filter := bson.M{
		"companyId": companyID,
		"name":      name,
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTotalBudget calculates total budget for all categories of a company
func (s *CategoryService) GetTotalBudget(ctx context.Context, companyID primitive.ObjectID) (float64, error) {
	categories, err := s.GetByCompanyID(ctx, companyID)
	if err != nil {
		return 0, err
	}

	var totalBudget float64
	for _, cat := range categories {
		totalBudget += cat.Budget
	}

	return totalBudget, nil
}

// GetBudgetByType calculates total budget by category type
func (s *CategoryService) GetBudgetByType(ctx context.Context, companyID primitive.ObjectID, categoryType models.CategoryType) (float64, error) {
	categories, err := s.GetByType(ctx, companyID, categoryType)
	if err != nil {
		return 0, err
	}

	var totalBudget float64
	for _, cat := range categories {
		totalBudget += cat.Budget
	}

	return totalBudget, nil
}

// SeedDefaultCategories creates default categories for a new company
func (s *CategoryService) SeedDefaultCategories(ctx context.Context, companyID primitive.ObjectID) error {
	defaultCategories := []models.Category{
		{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Salario",
			Type:      models.Income,
			Color:     "#22c55e",
			Budget:    0,
			CreatedAt: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Alimentacao",
			Type:      models.Expense,
			Color:     "#f59e0b",
			Budget:    1000,
			CreatedAt: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Transporte",
			Type:      models.Expense,
			Color:     "#3b82f6",
			Budget:    500,
			CreatedAt: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Moradia",
			Type:      models.Expense,
			Color:     "#8b5cf6",
			Budget:    2000,
			CreatedAt: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Lazer",
			Type:      models.Expense,
			Color:     "#ec4899",
			Budget:    300,
			CreatedAt: time.Now(),
		},
	}

	docs := make([]interface{}, len(defaultCategories))
	for i, cat := range defaultCategories {
		docs[i] = cat
	}

	_, err := s.collection.InsertMany(ctx, docs)
	return err
}
