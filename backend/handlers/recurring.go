package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

// CreateRecurring adds a new recurring rule
func CreateRecurring(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.CreateRecurringRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Description, "description"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidateAmount(req.Amount, "amount"),
		helpers.ValidateDayOfMonth(req.DayOfMonth),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateNoScriptTags(req.Description, "description"),
		helpers.ValidateSQLInjection(req.Description, "description"),
		helpers.ValidateNoScriptTags(req.Category, "category"),
		helpers.ValidateSQLInjection(req.Category, "category"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user inputs after validation
	req.Description = helpers.SanitizeString(req.Description)
	req.Category = helpers.SanitizeString(req.Category)

	// Convert companyId string to UUID
	companyUUID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	now := time.Now()
	rule := models.RecurringTransaction{
		ID:          uuid.New(),
		CompanyID:   companyUUID,
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		DayOfMonth:  req.DayOfMonth,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO recurring_transactions (id, company_id, description, amount, category, day_of_month, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rule.ID, rule.CompanyID, rule.Description, rule.Amount, rule.Category, rule.DayOfMonth, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return helpers.RecurringCreateFailed(c, err)
	}

	return c.JSON(rule)
}

// GetRecurring lists rules for a company
func GetRecurring(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyID := c.Query("companyId")
	if companyID == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	// Convert companyId string to UUID
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	rows, err := database.Query(ctx,
		`SELECT id, company_id, description, amount, category, day_of_month, created_at, updated_at
		 FROM recurring_transactions WHERE company_id = $1 ORDER BY day_of_month`,
		companyUUID)
	if err != nil {
		return helpers.RecurringFetchFailed(c, err)
	}
	defer rows.Close()

	var rules []models.RecurringTransaction
	for rows.Next() {
		var rule models.RecurringTransaction
		if err := rows.Scan(&rule.ID, &rule.CompanyID, &rule.Description, &rule.Amount,
			&rule.Category, &rule.DayOfMonth, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "decode_recurring_rules", err)
		}
		rules = append(rules, rule)
	}

	if rules == nil {
		rules = []models.RecurringTransaction{}
	}

	return c.JSON(rules)
}

// DeleteRecurring removes a rule
func DeleteRecurring(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	recurringID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateRecurringOwnership(c, recurringID)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `DELETE FROM recurring_transactions WHERE id = $1`, recurringID)
	if err != nil {
		return helpers.RecurringDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}

// ProcessRecurring checks and creates transactions for the given month/year
func ProcessRecurring(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	companyID := c.Query("companyId")
	month := c.Query("month")
	year := c.Query("year")

	if companyID == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}
	if month == "" {
		return helpers.MissingRequiredParam(c, "month")
	}
	if year == "" {
		return helpers.MissingRequiredParam(c, "year")
	}

	// Validate month
	if err := helpers.ValidateMonth(month); err != nil {
		return helpers.SendValidationError(c, err.Field, err.Message)
	}

	// Validate year
	var yearInt int
	fmt.Sscanf(year, "%d", &yearInt)
	if err := helpers.ValidateYear(yearInt); err != nil {
		return helpers.SendValidationError(c, err.Field, err.Message)
	}

	// Convert companyID string to UUID
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	// Get all rules for this company
	rows, err := database.Query(ctx,
		`SELECT id, company_id, description, amount, category, day_of_month, created_at, updated_at
		 FROM recurring_transactions WHERE company_id = $1`,
		companyUUID)
	if err != nil {
		return helpers.RecurringFetchFailed(c, err)
	}
	defer rows.Close()

	var rules []models.RecurringTransaction
	for rows.Next() {
		var rule models.RecurringTransaction
		if err := rows.Scan(&rule.ID, &rule.CompanyID, &rule.Description, &rule.Amount,
			&rule.Category, &rule.DayOfMonth, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "decode_recurring_rules", err)
		}
		rules = append(rules, rule)
	}

	createdCount := 0

	for _, rule := range rules {
		// Check if transaction already exists for this rule + month + year
		var count int
		err := database.QueryRow(ctx,
			`SELECT COUNT(*) FROM transactions
			 WHERE company_id = $1 AND description = $2 AND month = $3 AND year = $4`,
			companyUUID, rule.Description, month, yearInt).Scan(&count)
		if err != nil {
			continue
		}

		if count == 0 {
			// Create new transaction
			now := time.Now()
			newTrans := models.Transaction{
				ID:          uuid.New(),
				CompanyID:   companyUUID,
				Description: rule.Description,
				Amount:      rule.Amount,
				Category:    rule.Category,
				Status:      models.StatusOpen,
				Month:       month,
				Year:        yearInt,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			_, err := database.Pool.Exec(ctx,
				`INSERT INTO transactions (id, company_id, category, description, amount, month, year, status, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				newTrans.ID, newTrans.CompanyID, newTrans.Category, newTrans.Description,
				newTrans.Amount, newTrans.Month, newTrans.Year, newTrans.Status,
				newTrans.CreatedAt, newTrans.UpdatedAt)
			if err != nil {
				// Log error but continue processing other rules
				continue
			}
			createdCount++
		}
	}

	return c.JSON(fiber.Map{
		"message": "Processamento concluido",
		"created": createdCount,
	})
}
