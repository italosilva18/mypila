package services

import (
	"context"
	"errors"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

var (
	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")
)

// TransactionService contains business logic for transaction operations
type TransactionService struct {
	collection *mongo.Collection
}

// NewTransactionService creates a new instance of TransactionService
func NewTransactionService() *TransactionService {
	return &TransactionService{
		collection: database.GetCollection("transactions"),
	}
}

// GetCollection returns the transactions collection
func (s *TransactionService) GetCollection() *mongo.Collection {
	return s.collection
}

// GetPaginated retrieves paginated transactions for a company or all user companies
func (s *TransactionService) GetPaginated(ctx context.Context, companyID *primitive.ObjectID, userCompanyIDs []primitive.ObjectID, page, limit int) (*models.PaginatedTransactions, error) {
	query := bson.M{}

	if companyID != nil {
		query["companyId"] = *companyID
	} else if len(userCompanyIDs) > 0 {
		query["companyId"] = bson.M{"$in": userCompanyIDs}
	}

	// Count total documents matching the query
	total, err := s.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	skip := (page - 1) * limit

	// Query options with pagination and sorting
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.D{{Key: "year", Value: -1}, {Key: "month", Value: -1}})

	cursor, err := s.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return &models.PaginatedTransactions{
		Data: transactions,
		Pagination: models.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetByID retrieves a transaction by its ID
func (s *TransactionService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error) {
	var transaction models.Transaction
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

// Create creates a new transaction with validation
func (s *TransactionService) Create(ctx context.Context, companyID primitive.ObjectID, req models.CreateTransactionRequest) (*models.Transaction, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidatePositiveNumber(req.Amount, "amount"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateMonth(req.Month),
		helpers.ValidateRange(req.Year, 2000, 2100, "year"),
		helpers.ValidateStatus(string(req.Status)),
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

	transaction := &models.Transaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyID,
		Month:       req.Month,
		Year:        req.Year,
		Amount:      req.Amount,
		Category:    sanitizedCategory,
		Status:      req.Status,
		Description: sanitizedDescription,
	}

	_, err := s.collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, nil, err
	}

	return transaction, nil, nil
}

// Update updates an existing transaction with validation
func (s *TransactionService) Update(ctx context.Context, id primitive.ObjectID, req models.UpdateTransactionRequest) (*models.Transaction, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidatePositiveNumber(req.Amount, "amount"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateMonth(req.Month),
		helpers.ValidateRange(req.Year, 2000, 2100, "year"),
		helpers.ValidateStatus(string(req.Status)),
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
			"month":       req.Month,
			"year":        req.Year,
			"amount":      req.Amount,
			"category":    sanitizedCategory,
			"status":      req.Status,
			"description": sanitizedDescription,
		},
	}

	result, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, nil, err
	}

	if result.MatchedCount == 0 {
		return nil, nil, ErrTransactionNotFound
	}

	// Fetch updated transaction
	var transaction models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		return nil, nil, err
	}

	return &transaction, nil, nil
}

// Delete removes a transaction by its ID
func (s *TransactionService) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

// GetStats calculates financial statistics for a company or all user companies
func (s *TransactionService) GetStats(ctx context.Context, companyID *primitive.ObjectID, userCompanyIDs []primitive.ObjectID) (*models.Stats, error) {
	query := bson.M{}

	if companyID != nil {
		query["companyId"] = *companyID
	} else if len(userCompanyIDs) > 0 {
		query["companyId"] = bson.M{"$in": userCompanyIDs}
	}

	cursor, err := s.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	stats := &models.Stats{Paid: 0, Open: 0, Total: 0}
	for _, t := range transactions {
		if t.Status == models.StatusPaid {
			stats.Paid += t.Amount
		} else {
			stats.Open += t.Amount
		}
		stats.Total += t.Amount
	}

	return stats, nil
}

// ToggleStatus toggles the status of a transaction between PAGO and ABERTO
func (s *TransactionService) ToggleStatus(ctx context.Context, id primitive.ObjectID, currentStatus models.Status) (*models.Transaction, error) {
	newStatus := models.StatusOpen
	if currentStatus == models.StatusOpen {
		newStatus = models.StatusPaid
	}

	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": newStatus}})
	if err != nil {
		return nil, err
	}

	var transaction models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetByCompanyID retrieves all transactions for a company
func (s *TransactionService) GetByCompanyID(ctx context.Context, companyID primitive.ObjectID) ([]models.Transaction, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"companyId": companyID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return transactions, nil
}

// Seed creates initial transaction data if collection is empty
func (s *TransactionService) Seed(ctx context.Context, companyID primitive.ObjectID) (int, error) {
	// Check if already has data
	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	if count > 0 {
		return int(count), nil
	}

	// Initial transactions
	initialData := []interface{}{
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Janeiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Fevereiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Marco", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Abril", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Maio", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Junho", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Julho", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Agosto", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Setembro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Outubro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Novembro", Year: 2024, Amount: 5000, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: companyID, Month: "Dezembro", Year: 2024, Amount: 5000, Category: models.CategorySalary, Status: models.StatusOpen},
	}

	_, err = s.collection.InsertMany(ctx, initialData)
	if err != nil {
		return 0, err
	}

	return len(initialData), nil
}

// ValidateOwnership checks if a transaction belongs to a user's company
func (s *TransactionService) ValidateOwnership(ctx context.Context, transactionID primitive.ObjectID, userCompanyIDs []primitive.ObjectID) (*models.Transaction, error) {
	transaction, err := s.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Check if transaction's company is in user's companies
	for _, companyID := range userCompanyIDs {
		if transaction.CompanyID == companyID {
			return transaction, nil
		}
	}

	return nil, ErrUnauthorized
}

// DeleteByCompanyID removes all transactions for a company (cascade delete)
func (s *TransactionService) DeleteByCompanyID(ctx context.Context, companyID primitive.ObjectID) error {
	_, err := s.collection.DeleteMany(ctx, bson.M{"companyId": companyID})
	return err
}

// GetStatsByPeriod calculates stats for a specific month/year
func (s *TransactionService) GetStatsByPeriod(ctx context.Context, companyID primitive.ObjectID, month string, year int) (*models.Stats, error) {
	query := bson.M{
		"companyId": companyID,
		"month":     month,
		"year":      year,
	}

	cursor, err := s.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	stats := &models.Stats{Paid: 0, Open: 0, Total: 0}
	for _, t := range transactions {
		if t.Status == models.StatusPaid {
			stats.Paid += t.Amount
		} else {
			stats.Open += t.Amount
		}
		stats.Total += t.Amount
	}

	return stats, nil
}

// ExistsByFilter checks if transactions exist matching the given filter
func (s *TransactionService) ExistsByFilter(ctx context.Context, companyID primitive.ObjectID, description, month string, year int) (bool, error) {
	filter := bson.M{
		"companyId":   companyID,
		"description": description,
		"month":       month,
		"year":        year,
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateFromRecurring creates a transaction from a recurring rule
func (s *TransactionService) CreateFromRecurring(ctx context.Context, rule models.RecurringTransaction, month string, year int) (*models.Transaction, error) {
	transaction := &models.Transaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   rule.CompanyID,
		Description: rule.Description,
		Amount:      rule.Amount,
		Category:    rule.Category,
		Status:      models.StatusOpen,
		Month:       month,
		Year:        year,
	}

	_, err := s.collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
