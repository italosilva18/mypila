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

func GetCategories(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	// Convert companyId string to UUID
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
		`SELECT id, company_id, name, type, color, budget, created_at, updated_at
		 FROM categories WHERE company_id = $1 ORDER BY name`,
		companyUUID)
	if err != nil {
		return helpers.CategoryFetchFailed(c, err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.CompanyID, &cat.Name, &cat.Type, &cat.Color, &cat.Budget, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "decode_categories", err)
		}
		categories = append(categories, cat)
	}

	// If no categories found, return empty array
	if categories == nil {
		categories = []models.Category{}
	}

	return c.JSON(categories)
}

func CreateCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	// Convert companyId string to UUID
	companyUUID, err := uuid.Parse(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyUUID)
	if err != nil {
		return err
	}

	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateCategoryType(string(req.Type)),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
		helpers.ValidateBudget(req.Budget, "budget"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation
	req.Name = helpers.SanitizeString(req.Name)

	// Default to EXPENSE if not provided
	categoryType := req.Type
	if categoryType == "" {
		categoryType = models.Expense
	}

	now := time.Now()
	category := models.Category{
		ID:        uuid.New(),
		CompanyID: companyUUID,
		Name:      req.Name,
		Type:      categoryType,
		Color:     req.Color,
		Budget:    req.Budget,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO categories (id, company_id, name, type, color, budget, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		category.ID, category.CompanyID, category.Name, category.Type, category.Color, category.Budget, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return helpers.CategoryCreateFailed(c, err)
	}

	return c.Status(201).JSON(category)
}

func UpdateCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, categoryID)
	if err != nil {
		return err
	}

	var req models.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateCategoryType(string(req.Type)),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
		helpers.ValidateBudget(req.Budget, "budget"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation
	req.Name = helpers.SanitizeString(req.Name)

	now := time.Now()
	_, err = database.Pool.Exec(ctx,
		`UPDATE categories SET name = $1, type = $2, color = $3, budget = $4, updated_at = $5 WHERE id = $6`,
		req.Name, req.Type, req.Color, req.Budget, now, categoryID)
	if err != nil {
		return helpers.CategoryUpdateFailed(c, err)
	}

	var updatedCategory models.Category
	err = database.QueryRow(ctx,
		`SELECT id, company_id, name, type, color, budget, created_at, updated_at FROM categories WHERE id = $1`,
		categoryID).Scan(&updatedCategory.ID, &updatedCategory.CompanyID, &updatedCategory.Name,
		&updatedCategory.Type, &updatedCategory.Color, &updatedCategory.Budget,
		&updatedCategory.CreatedAt, &updatedCategory.UpdatedAt)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_updated_category", err)
	}

	return c.JSON(updatedCategory)
}

func DeleteCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, categoryID)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
	if err != nil {
		return helpers.CategoryDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}
