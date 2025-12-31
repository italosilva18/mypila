package handlers

import (
	"context"
	"fmt"
	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateRecurring adds a new recurring rule
func CreateRecurring(c *fiber.Ctx) error {
	var req models.CreateRecurringRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
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

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user inputs after validation
	req.Description = helpers.SanitizeString(req.Description)
	req.Category = helpers.SanitizeString(req.Category)

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(req.CompanyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	rule := models.RecurringTransaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyObjID,
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		DayOfMonth:  req.DayOfMonth,
		CreatedAt:   time.Now(),
	}

	collection := database.GetCollection("recurring")
	_, err = collection.InsertOne(context.Background(), rule)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar regra recorrente"})
	}

	return c.JSON(rule)
}

// GetRecurring lists rules for a company
func GetRecurring(c *fiber.Ctx) error {
	companyID := c.Query("companyId")
	if companyID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Company ID is required"})
	}

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	collection := database.GetCollection("recurring")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"companyId": companyObjID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch rules"})
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse rules"})
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return c.JSON(rules)
}

// DeleteRecurring removes a rule
func DeleteRecurring(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Validate ownership
	_, err = helpers.ValidateRecurringOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection("recurring")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete"})
	}
	return c.SendStatus(200)
}

// ProcessRecurring checks and creates transactions for the given month/year
func ProcessRecurring(c *fiber.Ctx) error {
	companyID := c.Query("companyId")
	month := c.Query("month")
	year := c.Query("year") // e.g. "2024" or 2024 (int)

	if companyID == "" || month == "" || year == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing params"})
	}

	// Convert companyID string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rulesColl := database.GetCollection("recurring")
	transColl := database.GetCollection("transactions")

	// Get all rules
	cursor, err := rulesColl.Find(ctx, bson.M{"companyId": companyObjID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch recurring rules"})
	}
	defer cursor.Close(ctx)

	var rules []models.RecurringTransaction
	if err = cursor.All(ctx, &rules); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse recurring rules"})
	}

	createdCount := 0

	for _, rule := range rules {
		// Check if transaction already exists for this rule + month + year
		// We use Description + Amount + Month + Year as a pseudo-unique key check
		// Or ideally store "SourceRuleID" in Transaction, but for now simple check:
		// Fix Year parsing
		var yearInt int
		fmt.Sscanf(year, "%d", &yearInt)

		filter := bson.M{
			"companyId":   companyObjID,
			"description": rule.Description,
			"month":       month,
			"year":        yearInt,
		}

		count, _ := transColl.CountDocuments(ctx, filter)

		if count == 0 {
			// Create it
			newTrans := models.Transaction{
				ID:          primitive.NewObjectID(),
				CompanyID:   companyObjID,
				Description: rule.Description,
				Amount:      rule.Amount,
				Category:    rule.Category,
				Status:      models.StatusOpen,
				Month:       month,
				Year:        yearInt,
			}
			transColl.InsertOne(ctx, newTrans)
			createdCount++
		}
	}

	return c.JSON(fiber.Map{
		"message": "Processed",
		"created": createdCount,
	})
}
