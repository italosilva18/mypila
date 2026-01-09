package handlers

import (
	"context"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

const collectionName = "transactions"

// GetAllTransactions returns paginated transactions (filtered by company with ownership validation)
func GetAllTransactions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection(collectionName)

	// Parse pagination parameters
	page := parsePage(c.Query("page", "1"))
	limit := parseLimit(c.Query("limit", "50"))

	// Build query filter
	query := bson.M{}
	if companyID := c.Query("companyId"); companyID != "" {
		objID, err := primitive.ObjectIDFromHex(companyID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
		}

		// Validate ownership before allowing query
		_, err = helpers.ValidateCompanyOwnership(c, objID)
		if err != nil {
			return err
		}

		query["companyId"] = objID
	} else {
		// If no companyId is provided, return transactions for all user's companies
		userID, err := helpers.GetUserIDFromContext(c)
		if err != nil {
			return err
		}

		// Get all company IDs owned by user
		companyCollection := database.GetCollection("companies")
		companyCursor, err := companyCollection.Find(ctx, bson.M{"userId": userID})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user companies"})
		}
		defer companyCursor.Close(ctx)

		var companies []models.Company
		if err := companyCursor.All(ctx, &companies); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode companies"})
		}

		// Extract company IDs
		companyIDs := make([]primitive.ObjectID, len(companies))
		for i, company := range companies {
			companyIDs[i] = company.ID
		}

		// Filter transactions by user's company IDs
		query["companyId"] = bson.M{"$in": companyIDs}
	}

	// Count total documents matching the query
	total, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count transactions"})
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	skip := (page - 1) * limit

	// Query options with pagination and sorting
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.D{{Key: "year", Value: -1}, {Key: "month", Value: -1}}) // Most recent first

	// Execute query
	cursor, err := collection.Find(ctx, query, findOptions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode transactions"})
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	// Build paginated response
	response := models.PaginatedTransactions{
		Data: transactions,
		Pagination: models.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return c.JSON(response)
}

// parsePage parses the page query parameter with validation
func parsePage(pageStr string) int {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// parseLimit parses the limit query parameter with validation (max 100)
func parseLimit(limitStr string) int {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 50 // Default
	}
	if limit > 100 {
		return 100 // Maximum
	}
	return limit
}

// GetTransaction returns a single transaction by ID (ownership validated)
func GetTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Validate ownership
	transaction, err := helpers.ValidateTransactionOwnership(c, objID)
	if err != nil {
		return err
	}

	return c.JSON(transaction)
}

// CreateTransaction creates a new transaction
func CreateTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
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

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user inputs after validation
	req.Description = helpers.SanitizeString(req.Description)
	req.Category = helpers.SanitizeString(req.Category)

	companyID, err := primitive.ObjectIDFromHex(req.CompanyID)
	if err != nil {
		return helpers.SendValidationError(c, "companyId", "Formato de ID da empresa inválido")
	}

	// Validate company ownership before creating transaction
	_, err = helpers.ValidateCompanyOwnership(c, companyID)
	if err != nil {
		return err
	}

	transaction := models.Transaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyID,
		Month:       req.Month,
		Year:        req.Year,
		Amount:      req.Amount,
		Category:    req.Category,
		Status:      req.Status,
		Description: req.Description,
	}

	collection := database.GetCollection(collectionName)

	_, err = collection.InsertOne(ctx, transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to create transaction: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar transação", "details": err.Error()})
	}

	return c.Status(201).JSON(transaction)
}

// UpdateTransaction updates an existing transaction (ownership validated)
func UpdateTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	_, err = helpers.ValidateTransactionOwnership(c, objID)
	if err != nil {
		return err
	}

	var req models.UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
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

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user inputs after validation
	req.Description = helpers.SanitizeString(req.Description)
	req.Category = helpers.SanitizeString(req.Category)

	collection := database.GetCollection(collectionName)

	update := bson.M{
		"$set": bson.M{
			"month":       req.Month,
			"year":        req.Year,
			"amount":      req.Amount,
			"category":    req.Category,
			"status":      req.Status,
			"description": req.Description,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar transação"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Transação não encontrada"})
	}

	// Fetch updated transaction
	var transaction models.Transaction
	collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)

	return c.JSON(transaction)
}

// DeleteTransaction deletes a transaction (ownership validated)
func DeleteTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Validate ownership
	_, err = helpers.ValidateTransactionOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(collectionName)

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete transaction"})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
	}

	return c.JSON(fiber.Map{"message": "Transaction deleted successfully"})
}

// GetStats returns financial statistics (ownership validated)
func GetStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection(collectionName)

	query := bson.M{}
	if companyID := c.Query("companyId"); companyID != "" {
		objID, err := primitive.ObjectIDFromHex(companyID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
		}

		// Validate ownership
		_, err = helpers.ValidateCompanyOwnership(c, objID)
		if err != nil {
			return err
		}

		query["companyId"] = objID
	} else {
		// If no companyId, return stats for all user's companies
		userID, err := helpers.GetUserIDFromContext(c)
		if err != nil {
			return err
		}

		// Get all company IDs owned by user
		companyCollection := database.GetCollection("companies")
		companyCursor, err := companyCollection.Find(ctx, bson.M{"userId": userID})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user companies"})
		}
		defer companyCursor.Close(ctx)

		var companies []models.Company
		if err := companyCursor.All(ctx, &companies); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode companies"})
		}

		// Extract company IDs
		companyIDs := make([]primitive.ObjectID, len(companies))
		for i, company := range companies {
			companyIDs[i] = company.ID
		}

		query["companyId"] = bson.M{"$in": companyIDs}
	}

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode transactions"})
	}

	stats := models.Stats{Paid: 0, Open: 0, Total: 0}
	for _, t := range transactions {
		if t.Status == models.StatusPaid {
			stats.Paid += t.Amount
		} else {
			stats.Open += t.Amount
		}
		stats.Total += t.Amount
	}

	return c.JSON(stats)
}

// ToggleStatus toggles the status of a transaction (ownership validated)
func ToggleStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Validate ownership
	transaction, err := helpers.ValidateTransactionOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(collectionName)

	// Toggle status
	newStatus := models.StatusOpen
	if transaction.Status == models.StatusOpen {
		newStatus = models.StatusPaid
	}

	// Update
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": newStatus}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	transaction.Status = newStatus
	return c.JSON(transaction)
}

// SeedTransactions seeds initial data if collection is empty
func SeedTransactions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection(collectionName)

	// Check if already has data
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count documents"})
	}

	if count > 0 {
		return c.JSON(fiber.Map{"message": "Data already seeded", "count": count})
	}

	// Check for M2M Company
	companyCollection := database.GetCollection("companies")
	var m2mCompany models.Company
	err = companyCollection.FindOne(ctx, bson.M{"name": "M2M Financeiro"}).Decode(&m2mCompany)

	// Create M2M company if it helps to seed
	if err != nil {
		// Get current user ID for seeding
		userID, err := helpers.GetUserIDFromContext(c)
		if err != nil {
			return err
		}

		m2mCompany = models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      "M2M Financeiro",
			CreatedAt: time.Now(),
		}
		if _, err := companyCollection.InsertOne(ctx, m2mCompany); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create default company"})
		}
	}

	// Initial transactions
	initialData := []interface{}{
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Janeiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Fevereiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Março", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Abril", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Maio", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Junho", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Julho", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Agosto", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Setembro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Outubro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Novembro", Year: 2024, Amount: 5000, Category: models.CategorySalary, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Dezembro", Year: 2024, Amount: 5000, Category: models.CategorySalary, Status: models.StatusOpen},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Acumulado", Year: 2024, Amount: 3765.34, Category: models.CategoryVacation, Status: models.StatusOpen, Description: "Baseado no salário da carteira R$ 1.412"},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Agosto", Year: 2024, Amount: 1000, Category: models.CategoryAICost, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Setembro", Year: 2024, Amount: 1000, Category: models.CategoryAICost, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Outubro", Year: 2024, Amount: 1000, Category: models.CategoryAICost, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Novembro", Year: 2024, Amount: 1100, Category: models.CategoryAICost, Status: models.StatusPaid},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Dezembro", Year: 2024, Amount: 1000, Category: models.CategoryAICost, Status: models.StatusOpen},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Novembro", Year: 2024, Amount: 1000, Category: models.CategoryDockerCost, Status: models.StatusOpen},
		models.Transaction{ID: primitive.NewObjectID(), CompanyID: m2mCompany.ID, Month: "Dezembro", Year: 2024, Amount: 1000, Category: models.CategoryDockerCost, Status: models.StatusOpen},
	}

	_, err = collection.InsertMany(ctx, initialData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to seed data"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Data seeded successfully", "count": len(initialData)})
}
