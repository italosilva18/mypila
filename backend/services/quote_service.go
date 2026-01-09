package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

var (
	// ErrQuoteNotFound is returned when a quote is not found
	ErrQuoteNotFound = errors.New("quote not found")

	// ErrQuoteExecuted is returned when trying to edit an executed quote
	ErrQuoteExecuted = errors.New("cannot edit an executed quote")

	// ErrInvalidQuoteStatus is returned when an invalid status is provided
	ErrInvalidQuoteStatus = errors.New("invalid quote status")

	// quoteNumberMutex prevents race condition in quote number generation
	quoteNumberMutex sync.Mutex
)

// QuoteService contains business logic for quote operations
type QuoteService struct {
	collection *mongo.Collection
}

// NewQuoteService creates a new instance of QuoteService
func NewQuoteService() *QuoteService {
	return &QuoteService{
		collection: database.GetCollection("quotes"),
	}
}

// GetCollection returns the quotes collection
func (s *QuoteService) GetCollection() *mongo.Collection {
	return s.collection
}

// generateQuoteNumber generates sequential quote number: ORC-2024-001
func (s *QuoteService) generateQuoteNumber(ctx context.Context, companyID primitive.ObjectID) (string, error) {
	quoteNumberMutex.Lock()
	defer quoteNumberMutex.Unlock()

	year := time.Now().Year()

	// Find the last quote of the year for this company
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	filter := bson.M{
		"companyId": companyID,
		"number": bson.M{
			"$regex": fmt.Sprintf("^ORC-%d-", year),
		},
	}

	var lastQuote models.Quote
	err := s.collection.FindOne(ctx, filter, opts).Decode(&lastQuote)

	nextNumber := 1
	if err == nil {
		// Extract number from last quote
		var lastNum int
		_, err := fmt.Sscanf(lastQuote.Number, "ORC-%d-%d", new(int), &lastNum)
		if err == nil {
			nextNumber = lastNum + 1
		}
	}

	return fmt.Sprintf("ORC-%d-%03d", year, nextNumber), nil
}

// GetByCompanyID retrieves all quotes for a company with optional status filter
func (s *QuoteService) GetByCompanyID(ctx context.Context, companyID primitive.ObjectID, status string) ([]models.Quote, error) {
	filter := bson.M{"companyId": companyID}
	if status != "" {
		filter["status"] = status
	}

	// Sort by creation date descending
	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err = cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	if quotes == nil {
		quotes = []models.Quote{}
	}

	return quotes, nil
}

// GetByID retrieves a quote by its ID
func (s *QuoteService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Quote, error) {
	var quote models.Quote
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&quote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrQuoteNotFound
		}
		return nil, err
	}
	return &quote, nil
}

// Create creates a new quote with validation
func (s *QuoteService) Create(ctx context.Context, companyID primitive.ObjectID, req models.CreateQuoteRequest) (*models.Quote, []helpers.ValidationError, error) {
	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateMaxLength(req.ClientName, "clientName", 100),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateMaxLength(req.Title, "title", 200),
		helpers.ValidateMaxLength(req.Description, "description", 1000),
		helpers.ValidateMaxLength(req.Notes, "notes", 2000),
		helpers.ValidateNoScriptTags(req.ClientName, "clientName"),
		helpers.ValidateNoScriptTags(req.Title, "title"),
		helpers.ValidateMongoInjection(req.ClientName, "clientName"),
		helpers.ValidateMongoInjection(req.Title, "title"),
	)

	if len(req.Items) == 0 {
		validationErrors = append(validationErrors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Generate quote number
	quoteNumber, err := s.generateQuoteNumber(ctx, companyID)
	if err != nil {
		return nil, nil, err
	}

	// Process items and calculate totals
	items, subtotal := s.processItems(req.Items)

	// Calculate discount
	discountType := models.DiscountType(req.DiscountType)
	if discountType == "" {
		discountType = models.DiscountValue
	}

	total := s.calculateTotal(subtotal, req.Discount, discountType)

	// Parse validUntil
	validUntil := time.Now().AddDate(0, 0, 30) // Default: 30 days
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	// Parse templateID
	var templateID primitive.ObjectID
	if req.TemplateID != "" {
		tid, err := primitive.ObjectIDFromHex(req.TemplateID)
		if err == nil {
			templateID = tid
		}
	}

	now := time.Now()
	quote := &models.Quote{
		ID:             primitive.NewObjectID(),
		CompanyID:      companyID,
		Number:         quoteNumber,
		ClientName:     helpers.SanitizeString(req.ClientName),
		ClientEmail:    helpers.SanitizeString(req.ClientEmail),
		ClientPhone:    helpers.SanitizeString(req.ClientPhone),
		ClientDocument: helpers.SanitizeString(req.ClientDocument),
		ClientAddress:  helpers.SanitizeString(req.ClientAddress),
		ClientCity:     helpers.SanitizeString(req.ClientCity),
		ClientState:    helpers.SanitizeString(req.ClientState),
		ClientZipCode:  helpers.SanitizeString(req.ClientZipCode),
		Title:          helpers.SanitizeString(req.Title),
		Description:    helpers.SanitizeString(req.Description),
		Items:          items,
		Subtotal:       subtotal,
		Discount:       req.Discount,
		DiscountType:   discountType,
		Total:          total,
		Status:         models.QuoteDraft,
		ValidUntil:     validUntil,
		Notes:          helpers.SanitizeString(req.Notes),
		TemplateID:     templateID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = s.collection.InsertOne(ctx, quote)
	if err != nil {
		return nil, nil, err
	}

	return quote, nil, nil
}

// Update updates an existing quote with validation
func (s *QuoteService) Update(ctx context.Context, id primitive.ObjectID, existingQuote *models.Quote, req models.UpdateQuoteRequest) (*models.Quote, []helpers.ValidationError, error) {
	// Check if quote can be edited
	if existingQuote.Status == models.QuoteExecuted {
		return nil, nil, ErrQuoteExecuted
	}

	// Validate input
	validationErrors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateMaxLength(req.ClientName, "clientName", 100),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateMaxLength(req.Title, "title", 200),
		helpers.ValidateNoScriptTags(req.ClientName, "clientName"),
		helpers.ValidateNoScriptTags(req.Title, "title"),
	)

	if len(req.Items) == 0 {
		validationErrors = append(validationErrors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

	if helpers.HasErrors(validationErrors) {
		return nil, validationErrors, ErrInvalidInput
	}

	// Process items and calculate totals
	items, subtotal := s.processItems(req.Items)

	// Calculate discount
	discountType := models.DiscountType(req.DiscountType)
	if discountType == "" {
		discountType = models.DiscountValue
	}

	total := s.calculateTotal(subtotal, req.Discount, discountType)

	// Parse validUntil
	validUntil := existingQuote.ValidUntil
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	// Parse templateID
	var templateID primitive.ObjectID
	if req.TemplateID != "" {
		tid, err := primitive.ObjectIDFromHex(req.TemplateID)
		if err == nil {
			templateID = tid
		}
	}

	update := bson.M{
		"$set": bson.M{
			"clientName":     helpers.SanitizeString(req.ClientName),
			"clientEmail":    helpers.SanitizeString(req.ClientEmail),
			"clientPhone":    helpers.SanitizeString(req.ClientPhone),
			"clientDocument": helpers.SanitizeString(req.ClientDocument),
			"clientAddress":  helpers.SanitizeString(req.ClientAddress),
			"clientCity":     helpers.SanitizeString(req.ClientCity),
			"clientState":    helpers.SanitizeString(req.ClientState),
			"clientZipCode":  helpers.SanitizeString(req.ClientZipCode),
			"title":          helpers.SanitizeString(req.Title),
			"description":    helpers.SanitizeString(req.Description),
			"items":          items,
			"subtotal":       subtotal,
			"discount":       req.Discount,
			"discountType":   discountType,
			"total":          total,
			"validUntil":     validUntil,
			"notes":          helpers.SanitizeString(req.Notes),
			"templateId":     templateID,
			"updatedAt":      time.Now(),
		},
	}

	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, nil, err
	}

	// Fetch updated quote
	var quote models.Quote
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&quote)
	if err != nil {
		return nil, nil, err
	}

	return &quote, nil, nil
}

// Delete removes a quote by its ID
func (s *QuoteService) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrQuoteNotFound
	}

	return nil
}

// UpdateStatus updates the status of a quote
func (s *QuoteService) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) (*models.Quote, error) {
	// Validate status
	validStatuses := map[string]bool{
		string(models.QuoteDraft):    true,
		string(models.QuoteSent):     true,
		string(models.QuoteApproved): true,
		string(models.QuoteRejected): true,
		string(models.QuoteExecuted): true,
	}

	if !validStatuses[status] {
		return nil, ErrInvalidQuoteStatus
	}

	update := bson.M{
		"$set": bson.M{
			"status":    models.QuoteStatus(status),
			"updatedAt": time.Now(),
		},
	}

	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}

	var quote models.Quote
	err = s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&quote)
	if err != nil {
		return nil, err
	}

	return &quote, nil
}

// Duplicate creates a copy of an existing quote
func (s *QuoteService) Duplicate(ctx context.Context, originalQuote *models.Quote) (*models.Quote, error) {
	// Generate new quote number
	quoteNumber, err := s.generateQuoteNumber(ctx, originalQuote.CompanyID)
	if err != nil {
		return nil, err
	}

	// Create new IDs for items
	var newItems []models.QuoteItem
	for _, item := range originalQuote.Items {
		newItem := item
		newItem.ID = primitive.NewObjectID()
		newItems = append(newItems, newItem)
	}

	now := time.Now()
	newQuote := &models.Quote{
		ID:             primitive.NewObjectID(),
		CompanyID:      originalQuote.CompanyID,
		Number:         quoteNumber,
		ClientName:     originalQuote.ClientName,
		ClientEmail:    originalQuote.ClientEmail,
		ClientPhone:    originalQuote.ClientPhone,
		ClientDocument: originalQuote.ClientDocument,
		ClientAddress:  originalQuote.ClientAddress,
		ClientCity:     originalQuote.ClientCity,
		ClientState:    originalQuote.ClientState,
		ClientZipCode:  originalQuote.ClientZipCode,
		Title:          originalQuote.Title + " (Copia)",
		Description:    originalQuote.Description,
		Items:          newItems,
		Subtotal:       originalQuote.Subtotal,
		Discount:       originalQuote.Discount,
		DiscountType:   originalQuote.DiscountType,
		Total:          originalQuote.Total,
		Status:         models.QuoteDraft,
		ValidUntil:     now.AddDate(0, 0, 30),
		Notes:          originalQuote.Notes,
		TemplateID:     originalQuote.TemplateID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = s.collection.InsertOne(ctx, newQuote)
	if err != nil {
		return nil, err
	}

	return newQuote, nil
}

// GetComparison returns quoted vs executed comparison
func (s *QuoteService) GetComparison(ctx context.Context, quote *models.Quote) (*models.QuoteComparison, error) {
	transCollection := database.GetCollection("transactions")

	var comparisonItems []models.QuoteComparisonItem
	var executedTotal float64

	for _, item := range quote.Items {
		compItem := models.QuoteComparisonItem{
			Description: item.Description,
			Quoted:      item.Total,
			Executed:    0,
			Variance:    0,
		}

		// If item has category, search transactions for that category
		if !item.CategoryID.IsZero() {
			compItem.CategoryID = item.CategoryID.Hex()

			filter := bson.M{
				"companyId": quote.CompanyID,
				"category":  item.CategoryID.Hex(),
			}

			cursor, err := transCollection.Find(ctx, filter)
			if err == nil {
				var transactions []models.Transaction
				if err = cursor.All(ctx, &transactions); err == nil {
					for _, t := range transactions {
						compItem.Executed += t.Amount
					}
				}
				cursor.Close(ctx)
			}
		}

		compItem.Variance = compItem.Quoted - compItem.Executed
		executedTotal += compItem.Executed
		comparisonItems = append(comparisonItems, compItem)
	}

	variance := quote.Total - executedTotal
	variancePercent := 0.0
	if quote.Total > 0 {
		variancePercent = (variance / quote.Total) * 100
	}

	comparison := &models.QuoteComparison{
		QuoteID:         quote.ID.Hex(),
		QuotedTotal:     quote.Total,
		ExecutedTotal:   executedTotal,
		Variance:        variance,
		VariancePercent: variancePercent,
		Items:           comparisonItems,
	}

	return comparison, nil
}

// ValidateOwnership checks if a quote belongs to a user's company
func (s *QuoteService) ValidateOwnership(ctx context.Context, quoteID primitive.ObjectID, userCompanyIDs []primitive.ObjectID) (*models.Quote, error) {
	quote, err := s.GetByID(ctx, quoteID)
	if err != nil {
		return nil, err
	}

	// Check if quote's company is in user's companies
	for _, companyID := range userCompanyIDs {
		if quote.CompanyID == companyID {
			return quote, nil
		}
	}

	return nil, ErrUnauthorized
}

// DeleteByCompanyID removes all quotes for a company (cascade delete)
func (s *QuoteService) DeleteByCompanyID(ctx context.Context, companyID primitive.ObjectID) error {
	_, err := s.collection.DeleteMany(ctx, bson.M{"companyId": companyID})
	return err
}

// GetByStatus retrieves quotes by status for a company
func (s *QuoteService) GetByStatus(ctx context.Context, companyID primitive.ObjectID, status models.QuoteStatus) ([]models.Quote, error) {
	filter := bson.M{
		"companyId": companyID,
		"status":    status,
	}

	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err = cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	if quotes == nil {
		quotes = []models.Quote{}
	}

	return quotes, nil
}

// GetExpired retrieves quotes that have passed their valid until date
func (s *QuoteService) GetExpired(ctx context.Context, companyID primitive.ObjectID) ([]models.Quote, error) {
	filter := bson.M{
		"companyId":  companyID,
		"validUntil": bson.M{"$lt": time.Now()},
		"status":     bson.M{"$nin": []models.QuoteStatus{models.QuoteApproved, models.QuoteExecuted, models.QuoteRejected}},
	}

	opts := options.Find().SetSort(bson.M{"validUntil": -1})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err = cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	if quotes == nil {
		quotes = []models.Quote{}
	}

	return quotes, nil
}

// GetTotalByStatus calculates total value of quotes by status
func (s *QuoteService) GetTotalByStatus(ctx context.Context, companyID primitive.ObjectID, status models.QuoteStatus) (float64, error) {
	quotes, err := s.GetByStatus(ctx, companyID, status)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, q := range quotes {
		total += q.Total
	}

	return total, nil
}

// processItems processes quote items and calculates subtotal
func (s *QuoteService) processItems(items []models.CreateQuoteItemRequest) ([]models.QuoteItem, float64) {
	var processedItems []models.QuoteItem
	var subtotal float64

	for _, item := range items {
		itemTotal := item.Quantity * item.UnitPrice
		subtotal += itemTotal

		quoteItem := models.QuoteItem{
			ID:          primitive.NewObjectID(),
			Description: helpers.SanitizeString(item.Description),
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       itemTotal,
		}

		if item.CategoryID != "" {
			catID, err := primitive.ObjectIDFromHex(item.CategoryID)
			if err == nil {
				quoteItem.CategoryID = catID
			}
		}

		processedItems = append(processedItems, quoteItem)
	}

	return processedItems, subtotal
}

// calculateTotal calculates the total with discount applied
func (s *QuoteService) calculateTotal(subtotal, discount float64, discountType models.DiscountType) float64 {
	if discountType == models.DiscountPercent {
		return subtotal - (subtotal * discount / 100)
	}
	return subtotal - discount
}
