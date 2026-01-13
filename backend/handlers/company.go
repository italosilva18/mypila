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

// nullString returns nil for empty strings, otherwise the pointer to the string
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

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
		`SELECT id, user_id, name, cnpj, legal_name, trade_name, email, phone, address, city, state, zip_code, logo_url, created_at, updated_at
		 FROM companies WHERE user_id = $1 ORDER BY name`,
		userID)
	if err != nil {
		return helpers.CompanyFetchFailed(c, err)
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var company models.Company
		var cnpj, legalName, tradeName, email, phone, address, city, state, zipCode, logoURL *string
		if err := rows.Scan(&company.ID, &company.UserID, &company.Name, &cnpj, &legalName, &tradeName, &email, &phone, &address, &city, &state, &zipCode, &logoURL, &company.CreatedAt, &company.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "scan_company", err)
		}
		if cnpj != nil { company.CNPJ = *cnpj }
		if legalName != nil { company.LegalName = *legalName }
		if tradeName != nil { company.TradeName = *tradeName }
		if email != nil { company.Email = *email }
		if phone != nil { company.Phone = *phone }
		if address != nil { company.Address = *address }
		if city != nil { company.City = *city }
		if state != nil { company.State = *state }
		if zipCode != nil { company.ZipCode = *zipCode }
		if logoURL != nil { company.LogoURL = *logoURL }
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
	req.LegalName = helpers.SanitizeString(req.LegalName)
	req.TradeName = helpers.SanitizeString(req.TradeName)
	req.Address = helpers.SanitizeString(req.Address)
	req.City = helpers.SanitizeString(req.City)

	now := time.Now()
	company := models.Company{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		CNPJ:      req.CNPJ,
		LegalName: req.LegalName,
		TradeName: req.TradeName,
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		City:      req.City,
		State:     req.State,
		ZipCode:   req.ZipCode,
		LogoURL:   req.LogoURL,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO companies (id, user_id, name, cnpj, legal_name, trade_name, email, phone, address, city, state, zip_code, logo_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		company.ID, company.UserID, company.Name, nullString(company.CNPJ), nullString(company.LegalName), nullString(company.TradeName),
		nullString(company.Email), nullString(company.Phone), nullString(company.Address), nullString(company.City),
		nullString(company.State), nullString(company.ZipCode), nullString(company.LogoURL), company.CreatedAt, company.UpdatedAt)
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
	var req models.UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations (name is optional in update)
	var errors []helpers.ValidationError
	if req.Name != "" {
		errors = helpers.CollectErrors(
			helpers.ValidateMaxLength(req.Name, "name", 100),
			helpers.ValidateNoScriptTags(req.Name, "name"),
			helpers.ValidateSQLInjection(req.Name, "name"),
		)
	}

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize
	req.Name = helpers.SanitizeString(req.Name)
	req.LegalName = helpers.SanitizeString(req.LegalName)
	req.TradeName = helpers.SanitizeString(req.TradeName)
	req.Address = helpers.SanitizeString(req.Address)
	req.City = helpers.SanitizeString(req.City)

	// Check ownership and update
	result, err := database.Pool.Exec(ctx,
		`UPDATE companies SET
			name = COALESCE(NULLIF($1, ''), name),
			cnpj = COALESCE($2, cnpj),
			legal_name = COALESCE($3, legal_name),
			trade_name = COALESCE($4, trade_name),
			email = COALESCE($5, email),
			phone = COALESCE($6, phone),
			address = COALESCE($7, address),
			city = COALESCE($8, city),
			state = COALESCE($9, state),
			zip_code = COALESCE($10, zip_code),
			logo_url = COALESCE($11, logo_url),
			updated_at = $12
		 WHERE id = $13 AND user_id = $14`,
		req.Name, nullString(req.CNPJ), nullString(req.LegalName), nullString(req.TradeName),
		nullString(req.Email), nullString(req.Phone), nullString(req.Address), nullString(req.City),
		nullString(req.State), nullString(req.ZipCode), nullString(req.LogoURL),
		time.Now(), companyID, userID)
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
	var cnpj, legalName, tradeName, email, phone, address, city, state, zipCode, logoURL *string
	err = database.QueryRow(ctx,
		`SELECT id, user_id, name, cnpj, legal_name, trade_name, email, phone, address, city, state, zip_code, logo_url, created_at, updated_at
		 FROM companies WHERE id = $1`,
		companyID).Scan(&company.ID, &company.UserID, &company.Name, &cnpj, &legalName, &tradeName, &email, &phone, &address, &city, &state, &zipCode, &logoURL, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_company", err)
	}
	if cnpj != nil { company.CNPJ = *cnpj }
	if legalName != nil { company.LegalName = *legalName }
	if tradeName != nil { company.TradeName = *tradeName }
	if email != nil { company.Email = *email }
	if phone != nil { company.Phone = *phone }
	if address != nil { company.Address = *address }
	if city != nil { company.City = *city }
	if state != nil { company.State = *state }
	if zipCode != nil { company.ZipCode = *zipCode }
	if logoURL != nil { company.LogoURL = *logoURL }

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
