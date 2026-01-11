package helpers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/models"
)

// ValidateCompanyOwnership validates that the authenticated user owns the specified company
// Returns the company if validation succeeds, or sends an error response and returns nil
func ValidateCompanyOwnership(c *fiber.Ctx, companyID uuid.UUID) (*models.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get authenticated user ID from context
	userIDStr, ok := c.Locals("userId").(string)
	if !ok {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: user not authenticated"})
		return nil, fiber.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: invalid user ID"})
		return nil, fiber.ErrUnauthorized
	}

	// Fetch company from database
	var company models.Company
	err = database.QueryRow(ctx,
		`SELECT id, user_id, name, created_at, updated_at FROM companies WHERE id = $1`,
		companyID).Scan(&company.ID, &company.UserID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
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
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
		return nil, fiber.ErrBadRequest
	}

	return ValidateCompanyOwnership(c, companyID)
}

// GetUserIDFromContext extracts and validates the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDStr, ok := c.Locals("userId").(string)
	if !ok {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: user not authenticated"})
		return uuid.Nil, fiber.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Status(401).JSON(fiber.Map{"error": "Unauthorized: invalid user ID"})
		return uuid.Nil, fiber.ErrUnauthorized
	}

	return userID, nil
}

// ValidateTransactionOwnership validates that a transaction belongs to a company owned by the user
func ValidateTransactionOwnership(c *fiber.Ctx, transactionID uuid.UUID) (*models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch transaction
	var transaction models.Transaction
	var description *string
	err := database.QueryRow(ctx,
		`SELECT id, company_id, category, description, amount, month, year, status, created_at, updated_at
		 FROM transactions WHERE id = $1`,
		transactionID).Scan(&transaction.ID, &transaction.CompanyID, &transaction.Category, &description,
		&transaction.Amount, &transaction.Month, &transaction.Year, &transaction.Status,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Transaction not found"})
		return nil, fiber.ErrNotFound
	}
	if description != nil {
		transaction.Description = *description
	}

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, transaction.CompanyID)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// ValidateCategoryOwnership validates that a category belongs to a company owned by the user
func ValidateCategoryOwnership(c *fiber.Ctx, categoryID uuid.UUID) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch category
	var category models.Category
	err := database.QueryRow(ctx,
		`SELECT id, company_id, name, type, color, budget, created_at, updated_at FROM categories WHERE id = $1`,
		categoryID).Scan(&category.ID, &category.CompanyID, &category.Name, &category.Type,
		&category.Color, &category.Budget, &category.CreatedAt, &category.UpdatedAt)
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
func ValidateRecurringOwnership(c *fiber.Ctx, recurringID uuid.UUID) (*models.RecurringTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch recurring transaction
	var recurring models.RecurringTransaction
	err := database.QueryRow(ctx,
		`SELECT id, company_id, description, amount, category, day_of_month, created_at, updated_at
		 FROM recurring_transactions WHERE id = $1`,
		recurringID).Scan(&recurring.ID, &recurring.CompanyID, &recurring.Description,
		&recurring.Amount, &recurring.Category, &recurring.DayOfMonth, &recurring.CreatedAt, &recurring.UpdatedAt)
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

// ValidateQuoteOwnership validates that a quote belongs to a company owned by the user
func ValidateQuoteOwnership(c *fiber.Ctx, quoteID uuid.UUID) (*models.Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch quote
	var quote models.Quote
	var templateID *uuid.UUID
	err := database.QueryRow(ctx,
		`SELECT id, company_id, number, client_name, client_email, client_phone, client_document,
		 client_address, client_city, client_state, client_zip_code, title, description,
		 subtotal, discount, discount_type, total, status, valid_until, notes, template_id,
		 created_at, updated_at FROM quotes WHERE id = $1`,
		quoteID).Scan(&quote.ID, &quote.CompanyID, &quote.Number, &quote.ClientName, &quote.ClientEmail,
		&quote.ClientPhone, &quote.ClientDocument, &quote.ClientAddress, &quote.ClientCity,
		&quote.ClientState, &quote.ClientZipCode, &quote.Title, &quote.Description,
		&quote.Subtotal, &quote.Discount, &quote.DiscountType, &quote.Total, &quote.Status,
		&quote.ValidUntil, &quote.Notes, &templateID, &quote.CreatedAt, &quote.UpdatedAt)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Quote not found"})
		return nil, fiber.ErrNotFound
	}
	quote.TemplateID = templateID

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, quote.CompanyID)
	if err != nil {
		return nil, err
	}

	return &quote, nil
}

// ValidateQuoteTemplateOwnership validates that a quote template belongs to a company owned by the user
func ValidateQuoteTemplateOwnership(c *fiber.Ctx, templateID uuid.UUID) (*models.QuoteTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch template
	var template models.QuoteTemplate
	err := database.QueryRow(ctx,
		`SELECT id, company_id, name, header_text, footer_text, terms_text, primary_color, logo_url, is_default, created_at, updated_at
		 FROM quote_templates WHERE id = $1`,
		templateID).Scan(&template.ID, &template.CompanyID, &template.Name, &template.HeaderText,
		&template.FooterText, &template.TermsText, &template.PrimaryColor, &template.LogoURL,
		&template.IsDefault, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		c.Status(404).JSON(fiber.Map{"error": "Quote template not found"})
		return nil, fiber.ErrNotFound
	}

	// Validate company ownership
	_, err = ValidateCompanyOwnership(c, template.CompanyID)
	if err != nil {
		return nil, err
	}

	return &template, nil
}
