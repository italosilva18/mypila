package helpers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m2m-backend/database"
	"m2m-backend/models"
)

// ValidateCompanyOwnership validates that the authenticated user owns the specified company
// Returns the company if validation succeeds, or sends an error response and returns nil
func ValidateCompanyOwnership(c *fiber.Ctx, companyID primitive.ObjectID) (*models.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get authenticated user ID from context
	userIDStr, ok := c.Locals("userId").(string)
	if !ok {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: user not authenticated"})
		return nil, fiber.ErrUnauthorized
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: invalid user ID"})
		return nil, fiber.ErrUnauthorized
	}

	// Fetch company from database
	collection := database.GetCollection("companies")
	var company models.Company
	err = collection.FindOne(ctx, bson.M{"_id": companyID}).Decode(&company)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Company not found"})
		return nil, fiber.ErrNotFound
	}

	// Validate ownership
	if company.UserID != userID {
		c.Status(403).JSON(fiber.Map{
			"error":   "Forbidden: you do not have permission to access this company",
			"details": "This company belongs to another user",
		})
		return nil, fiber.ErrForbidden
	}

	return &company, nil
}

// ValidateCompanyOwnershipByString is a convenience wrapper that accepts company ID as string
func ValidateCompanyOwnershipByString(c *fiber.Ctx, companyIDStr string) (*models.Company, error) {
	companyID, err := primitive.ObjectIDFromHex(companyIDStr)
	if err != nil {
		c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
		return nil, fiber.ErrBadRequest
	}

	return ValidateCompanyOwnership(c, companyID)
}

// GetUserIDFromContext extracts and validates the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (primitive.ObjectID, error) {
	userIDStr, ok := c.Locals("userId").(string)
	if !ok {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: user not authenticated"})
		return primitive.NilObjectID, fiber.ErrUnauthorized
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: invalid user ID"})
		return primitive.NilObjectID, fiber.ErrUnauthorized
	}

	return userID, nil
}

// ValidateTransactionOwnership validates that a transaction belongs to a company owned by the user
func ValidateTransactionOwnership(c *fiber.Ctx, transactionID primitive.ObjectID) (*models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch transaction
	collection := database.GetCollection("transactions")
	var transaction models.Transaction
	err := collection.FindOne(ctx, bson.M{"_id": transactionID}).Decode(&transaction)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
		return nil, fiber.ErrNotFound
	}

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, transaction.CompanyID)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// ValidateCategoryOwnership validates that a category belongs to a company owned by the user
func ValidateCategoryOwnership(c *fiber.Ctx, categoryID primitive.ObjectID) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch category
	collection := database.GetCollection("categories")
	var category models.Category
	err := collection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Category not found"})
		return nil, fiber.ErrNotFound
	}

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, category.CompanyID)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

// ValidateRecurringOwnership validates that a recurring transaction belongs to a company owned by the user
func ValidateRecurringOwnership(c *fiber.Ctx, recurringID primitive.ObjectID) (*models.RecurringTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch recurring transaction
	collection := database.GetCollection("recurring")
	var recurring models.RecurringTransaction
	err := collection.FindOne(ctx, bson.M{"_id": recurringID}).Decode(&recurring)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Recurring transaction not found"})
		return nil, fiber.ErrNotFound
	}

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, recurring.CompanyID)
	if err != nil {
		return nil, err
	}

	return &recurring, nil
}
