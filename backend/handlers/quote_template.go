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

// GetQuoteTemplates lista todos os templates de orcamento de uma empresa
func GetQuoteTemplates(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	companyUUID, err := uuid.Parse(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	rows, err := database.Query(ctx,
		`SELECT id, company_id, name, header_text, footer_text, terms_text, primary_color, logo_url, is_default, created_at, updated_at
		 FROM quote_templates WHERE company_id = $1 ORDER BY name`,
		companyUUID)
	if err != nil {
		return helpers.QuoteTemplateFetchFailed(c, err)
	}
	defer rows.Close()

	var templates []models.QuoteTemplate
	for rows.Next() {
		var t models.QuoteTemplate
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.Name, &t.HeaderText, &t.FooterText,
			&t.TermsText, &t.PrimaryColor, &t.LogoURL, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "decode_templates", err)
		}
		templates = append(templates, t)
	}

	if templates == nil {
		templates = []models.QuoteTemplate{}
	}

	return c.JSON(templates)
}

// GetQuoteTemplate busca um template por ID
func GetQuoteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	templateID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	template, err := helpers.ValidateQuoteTemplateOwnership(c, templateID)
	if err != nil {
		return err
	}

	return c.JSON(template)
}

// CreateQuoteTemplate cria um novo template de orcamento
func CreateQuoteTemplate(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	companyUUID, err := uuid.Parse(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	var req models.CreateQuoteTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateMaxLength(req.HeaderText, "headerText", 500),
		helpers.ValidateMaxLength(req.FooterText, "footerText", 500),
		helpers.ValidateMaxLength(req.TermsText, "termsText", 2000),
		helpers.ValidateHexColor(req.PrimaryColor),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateNoScriptTags(req.HeaderText, "headerText"),
		helpers.ValidateNoScriptTags(req.FooterText, "footerText"),
		helpers.ValidateNoScriptTags(req.TermsText, "termsText"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize inputs
	req.Name = helpers.SanitizeString(req.Name)
	req.HeaderText = helpers.SanitizeString(req.HeaderText)
	req.FooterText = helpers.SanitizeString(req.FooterText)
	req.TermsText = helpers.SanitizeString(req.TermsText)

	// Se este template deve ser o padrao, remover flag dos outros
	if req.IsDefault {
		_, err = database.Pool.Exec(ctx,
			`UPDATE quote_templates SET is_default = false WHERE company_id = $1 AND is_default = true`,
			companyUUID)
		if err != nil {
			return helpers.DatabaseError(c, "update_existing_templates", err)
		}
	}

	// Default color
	if req.PrimaryColor == "" {
		req.PrimaryColor = "#78716c"
	}

	now := time.Now()
	template := models.QuoteTemplate{
		ID:           uuid.New(),
		CompanyID:    companyUUID,
		Name:         req.Name,
		HeaderText:   req.HeaderText,
		FooterText:   req.FooterText,
		TermsText:    req.TermsText,
		PrimaryColor: req.PrimaryColor,
		LogoURL:      req.LogoURL,
		IsDefault:    req.IsDefault,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO quote_templates (id, company_id, name, header_text, footer_text, terms_text, primary_color, logo_url, is_default, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		template.ID, template.CompanyID, template.Name, template.HeaderText, template.FooterText,
		template.TermsText, template.PrimaryColor, template.LogoURL, template.IsDefault, template.CreatedAt, template.UpdatedAt)
	if err != nil {
		return helpers.QuoteTemplateCreateFailed(c, err)
	}

	return c.Status(201).JSON(template)
}

// UpdateQuoteTemplate atualiza um template existente
func UpdateQuoteTemplate(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	templateID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	existingTemplate, err := helpers.ValidateQuoteTemplateOwnership(c, templateID)
	if err != nil {
		return err
	}

	var req models.UpdateQuoteTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateMaxLength(req.HeaderText, "headerText", 500),
		helpers.ValidateMaxLength(req.FooterText, "footerText", 500),
		helpers.ValidateMaxLength(req.TermsText, "termsText", 2000),
		helpers.ValidateHexColor(req.PrimaryColor),
		helpers.ValidateNoScriptTags(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Se este template deve ser o padrao, remover flag dos outros
	if req.IsDefault && !existingTemplate.IsDefault {
		_, err = database.Pool.Exec(ctx,
			`UPDATE quote_templates SET is_default = false WHERE company_id = $1 AND is_default = true AND id != $2`,
			existingTemplate.CompanyID, templateID)
		if err != nil {
			return helpers.DatabaseError(c, "update_existing_templates", err)
		}
	}

	now := time.Now()
	_, err = database.Pool.Exec(ctx,
		`UPDATE quote_templates SET name = $1, header_text = $2, footer_text = $3, terms_text = $4,
		 primary_color = $5, logo_url = $6, is_default = $7, updated_at = $8 WHERE id = $9`,
		helpers.SanitizeString(req.Name), helpers.SanitizeString(req.HeaderText),
		helpers.SanitizeString(req.FooterText), helpers.SanitizeString(req.TermsText),
		req.PrimaryColor, req.LogoURL, req.IsDefault, now, templateID)
	if err != nil {
		return helpers.QuoteTemplateUpdateFailed(c, err)
	}

	var updatedTemplate models.QuoteTemplate
	err = database.QueryRow(ctx,
		`SELECT id, company_id, name, header_text, footer_text, terms_text, primary_color, logo_url, is_default, created_at, updated_at
		 FROM quote_templates WHERE id = $1`,
		templateID).Scan(&updatedTemplate.ID, &updatedTemplate.CompanyID, &updatedTemplate.Name, &updatedTemplate.HeaderText,
		&updatedTemplate.FooterText, &updatedTemplate.TermsText, &updatedTemplate.PrimaryColor,
		&updatedTemplate.LogoURL, &updatedTemplate.IsDefault, &updatedTemplate.CreatedAt, &updatedTemplate.UpdatedAt)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_updated_template", err)
	}

	return c.JSON(updatedTemplate)
}

// DeleteQuoteTemplate exclui um template
func DeleteQuoteTemplate(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	templateID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteTemplateOwnership(c, templateID)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `DELETE FROM quote_templates WHERE id = $1`, templateID)
	if err != nil {
		return helpers.QuoteTemplateDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}
