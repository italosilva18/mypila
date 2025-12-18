package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

const categoryCollection = "categories"

func GetCategories(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Company ID is required"})
	}

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(categoryCollection)

	cursor, err := collection.Find(ctx, bson.M{"companyId": companyObjID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse categories"})
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
		return helpers.SendValidationError(c, "companyId", "ID da empresa é obrigatório")
	}

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation
	req.Name = helpers.SanitizeString(req.Name)

	collection := database.GetCollection(categoryCollection)

	category := models.Category{
		ID:        primitive.NewObjectID(),
		CompanyID: companyObjID,
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		Budget:    req.Budget,
		CreatedAt: time.Now(),
	}

	// Default to EXPENSE if not provided
	if category.Type == "" {
		category.Type = models.Expense
	}

	_, err = collection.InsertOne(ctx, category)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar categoria"})
	}

	return c.Status(201).JSON(category)
}

func UpdateCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, objID)
	if err != nil {
		return err
	}

	var req models.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 50),
		helpers.ValidateHexColor(req.Color),
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation
	req.Name = helpers.SanitizeString(req.Name)

	collection := database.GetCollection(categoryCollection)

	update := bson.M{
		"$set": bson.M{
			"name":   req.Name,
			"type":   req.Type,
			"color":  req.Color,
			"budget": req.Budget,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar categoria"})
	}

	var updatedCategory models.Category
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedCategory)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar categoria atualizada"})
	}

	return c.JSON(updatedCategory)
}

func DeleteCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(categoryCollection)

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete category"})
	}

	return c.SendStatus(204)
}
