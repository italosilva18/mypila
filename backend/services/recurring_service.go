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
	// ErrRecurringNotFound is returned when a recurring transaction is not found
	ErrRecurringNotFound = errors.New("recurring transaction not found")
)

// RecurringService contains business logic for recurring transaction operations
type RecurringService struct {
	collection *mongo.Collection
}

// NewRecurringService creates a new instance of RecurringService
func NewRecurringService() *RecurringService {
	return &RecurringService{
		collection: database.GetCollection("recurring"),
	}
}

// GetCollection returns the recurring collection
func (s *RecurringService) GetCollection() *mongo.Collection {
	return s.collection
}

// GetByCompanyID retrieves all recurring transactions for a company
func (s *RecurringService) GetByCompanyID(ctx context.Context, companyID primitive.ObjectID) ([]models.RecurringTransaction, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"companyId": companyID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return rules, nil
}

// GetByID retrieves a recurring transaction by its ID
func (s *RecurringService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.RecurringTransaction, error) {
	var rule models.RecurringTransaction
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecurringNotFound
		}
		return nil, err
	}
	return &rule, nil
}

// Create creates a new recurring transaction with validation
func (s *RecurringService) Create(ctx context.Context, companyID primitive.ObjectID, req models.CreateRecurringRequest) (*models.RecurringTransaction, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Description, "description"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidatePositiveNumber(req.Amount, "amount"),
		helpers.ValidateDayOfMonth(req.DayOfMonth),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateNoScriptTags(req.Description, "description"),
		helpers.ValidateMongoInjection(req.Description, "description"),
		helpers.ValidateSQLInjection(req.Description, "description"),
		helpers.ValidateNoScriptTags(req.Category, "category"),
		helpers.ValidateMongoInjection(req.Category, "category"),
		helpers.ValidateSQLInjection(req.Category, "category"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user inputs
	sanitizedDescription := helpers.SanitizeString(req.Description)
	sanitizedCategory := helpers.SanitizeString(req.Category)

	rule := &models.RecurringTransaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyID,
		Description: sanitizedDescription,
		Amount:      req.Amount,
		Category:    sanitizedCategory,
		DayOfMonth:  req.DayOfMonth,
		CreatedAt:   time.Now(),
	}

	_, err := s.collection.InsertOne(ctx, rule)
	if err != nil {
		return nil, nil, err
	}

	return rule, nil, nil
}

// Update updates an existing recurring transaction with validation
func (s *RecurringService) Update(ctx context.Context, id primitive.ObjectID, req models.CreateRecurringRequest) (*models.RecurringTransaction, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Description, "description"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidatePositiveNumber(req.Amount, "amount"),
		helpers.ValidateDayOfMonth(req.DayOfMonth),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateNoScriptTags(req.Description, "description"),
		helpers.ValidateMongoInjection(req.Description, "description"),
		helpers.ValidateSQLInjection(req.Description, "description"),
		helpers.ValidateNoScriptTags(req.Category, "category"),
		helpers.ValidateMongoInjection(req.Category, "category"),
		helpers.ValidateSQLInjection(req.Category, "category"),
	)

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Sanitize user inputs
	sanitizedDescription := helpers.SanitizeString(req.Description)
	sanitizedCategory := helpers.SanitizeString(req.Category)

	update := bson.M{
		"$set": bson.M{
			"description": sanitizedDescription,
			"amount":      req.Amount,
			"category":    sanitizedCategory,
			"dayOfMonth":  req.DayOfMonth,
		},
	}

	result, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, nil, err
	}

	if result.MatchedCount == 0 {
		return nil, nil, ErrRecurringNotFound
	}

	// Fetch updated rule
	var rule models.RecurringTransaction
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		return nil, nil, err
	}

	return &rule, nil, nil
}

// Delete removes a recurring transaction by its ID
func (s *RecurringService) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrRecurringNotFound
	}

	return nil
}

// ValidateOwnership checks if a recurring transaction belongs to a user's company
func (s *RecurringService) ValidateOwnership(ctx context.Context, recurringID primitive.ObjectID, userCompanyIDs []primitive.ObjectID) (*models.RecurringTransaction, error) {
	rule, err := s.GetByID(ctx, recurringID)
	if err != nil {
		return nil, err
	}

	// Check if recurring's company is in user's companies
	for _, companyID := range userCompanyIDs {
		if rule.CompanyID == companyID {
			return rule, nil
		}
	}

	return nil, ErrUnauthorized
}

// DeleteByCompanyID removes all recurring transactions for a company (cascade delete)
func (s *RecurringService) DeleteByCompanyID(ctx context.Context, companyID primitive.ObjectID) error {
	_, err := s.collection.DeleteMany(ctx, bson.M{"companyId": companyID})
	return err
}

// ProcessForPeriod checks and creates transactions for the given month/year
func (s *RecurringService) ProcessForPeriod(ctx context.Context, companyID primitive.ObjectID, month string, year int) (int, error) {
	transactionService := NewTransactionService()

	// Get all recurring rules for the company
	rules, err := s.GetByCompanyID(ctx, companyID)
	if err != nil {
		return 0, err
	}

	createdCount := 0

	for _, rule := range rules {
		// Check if transaction already exists for this rule + month + year
		exists, err := transactionService.ExistsByFilter(ctx, companyID, rule.Description, month, year)
		if err != nil {
			continue
		}

		if !exists {
			// Create new transaction
			_, err := transactionService.CreateFromRecurring(ctx, rule, month, year)
			if err == nil {
				createdCount++
			}
		}
	}

	return createdCount, nil
}

// GetByDayOfMonth retrieves recurring transactions for a specific day
func (s *RecurringService) GetByDayOfMonth(ctx context.Context, companyID primitive.ObjectID, day int) ([]models.RecurringTransaction, error) {
	filter := bson.M{
		"companyId":  companyID,
		"dayOfMonth": day,
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return rules, nil
}

// GetTotalAmount calculates total amount of all recurring transactions for a company
func (s *RecurringService) GetTotalAmount(ctx context.Context, companyID primitive.ObjectID) (float64, error) {
	rules, err := s.GetByCompanyID(ctx, companyID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, rule := range rules {
		total += rule.Amount
	}

	return total, nil
}

// GetByCategory retrieves recurring transactions by category
func (s *RecurringService) GetByCategory(ctx context.Context, companyID primitive.ObjectID, category string) ([]models.RecurringTransaction, error) {
	filter := bson.M{
		"companyId": companyID,
		"category":  category,
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return rules, nil
}

// Count returns the total count of recurring transactions for a company
func (s *RecurringService) Count(ctx context.Context, companyID primitive.ObjectID) (int64, error) {
	return s.collection.CountDocuments(ctx, bson.M{"companyId": companyID})
}

// ExistsByDescription checks if a recurring transaction with the given description exists
func (s *RecurringService) ExistsByDescription(ctx context.Context, companyID primitive.ObjectID, description string) (bool, error) {
	filter := bson.M{
		"companyId":   companyID,
		"description": description,
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetDueTodayForAllCompanies retrieves all recurring transactions due today
func (s *RecurringService) GetDueTodayForAllCompanies(ctx context.Context) ([]models.RecurringTransaction, error) {
	today := time.Now().Day()

	cursor, err := s.collection.Find(ctx, bson.M{"dayOfMonth": today})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return rules, nil
}

// ProcessAllDueToday processes all recurring transactions due today for all companies
func (s *RecurringService) ProcessAllDueToday(ctx context.Context) (int, error) {
	transactionService := NewTransactionService()

	today := time.Now()
	month := helpers.GetMonthName(int(today.Month()))
	year := today.Year()

	rules, err := s.GetDueTodayForAllCompanies(ctx)
	if err != nil {
		return 0, err
	}

	createdCount := 0

	for _, rule := range rules {
		// Check if transaction already exists
		exists, err := transactionService.ExistsByFilter(ctx, rule.CompanyID, rule.Description, month, year)
		if err != nil {
			continue
		}

		if !exists {
			_, err := transactionService.CreateFromRecurring(ctx, rule, month, year)
			if err == nil {
				createdCount++
			}
		}
	}

	return createdCount, nil
}
