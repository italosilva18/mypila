package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

// GetCompanies returns all companies owned by the authenticated user
func GetCompanies(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	rows, err := database.Query(ctx,
		`SELECT id, user_id, name, created_at, updated_at FROM companies WHERE user_id = $1 ORDER BY name`,
		userID)
	if err != nil {
		return helpers.CompanyFetchFailed(c, err)
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.UserID, &company.Name, &company.CreatedAt, &company.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "scan_company", err)
		}
		companies = append(companies, company)
	}

	if companies == nil {
		companies = []models.Company{}
	}

	return c.JSON(companies)
}

// CreateCompany creates a new company for the authenticated user
func CreateCompany(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize
	req.Name = helpers.SanitizeString(req.Name)

	now := time.Now()
	company := models.Company{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO companies (id, user_id, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		company.ID, company.UserID, company.Name, company.CreatedAt, company.UpdatedAt)
	if err != nil {
		return helpers.CompanyCreateFailed(c, err)
	}

	return c.Status(201).JSON(company)
}

// UpdateCompany updates an existing company (ownership validated)
func UpdateCompany(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Parse request body
	var req models.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize
	req.Name = helpers.SanitizeString(req.Name)

	// Check ownership and update
	result, err := database.Pool.Exec(ctx,
		`UPDATE companies SET name = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`,
		req.Name, time.Now(), companyID, userID)
	if err != nil {
		return helpers.CompanyUpdateFailed(c, err)
	}

	if result.RowsAffected() == 0 {
		// Check if company exists
		var exists bool
		database.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)", companyID).Scan(&exists)
		if !exists {
			return helpers.CompanyNotFound(c)
		}
		return helpers.Forbidden(c, helpers.ErrCodeForbidden, "Voce nao tem permissao para acessar esta empresa", helpers.ErrorDetails{
			"reason": "Esta empresa pertence a outro usuario",
		})
	}

	// Fetch updated company
	var company models.Company
	err = database.QueryRow(ctx,
		`SELECT id, user_id, name, created_at, updated_at FROM companies WHERE id = $1`,
		companyID).Scan(&company.ID, &company.UserID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_company", err)
	}

	return c.JSON(company)
}

// DeleteCompany deletes a company and all related data in cascade (ownership validated)
func DeleteCompany(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse and validate company ID
	id := c.Params("id")
	companyID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Delete with ownership check (CASCADE will handle related records)
	result, err := database.Pool.Exec(ctx,
		`DELETE FROM companies WHERE id = $1 AND user_id = $2`,
		companyID, userID)
	if err != nil {
		return helpers.CompanyDeleteFailed(c, err)
	}

	if result.RowsAffected() == 0 {
		// Check if company exists
		var exists bool
		database.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)", companyID).Scan(&exists)
		if !exists {
			return helpers.CompanyNotFound(c)
		}
		return helpers.Forbidden(c, helpers.ErrCodeForbidden, "Voce nao tem permissao para excluir esta empresa", helpers.ErrorDetails{
			"reason": "Esta empresa pertence a outro usuario",
		})
	}

	return c.JSON(fiber.Map{"message": "Empresa excluida com sucesso"})
}
