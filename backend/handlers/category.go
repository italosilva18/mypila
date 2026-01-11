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
		return helpers.MissingRequiredParam(c, "companyId")
	}

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(categoryCollection)

	cursor, err := collection.Find(ctx, bson.M{"companyId": companyObjID})
	if err != nil {
		return helpers.CategoryFetchFailed(c, err)
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return helpers.DatabaseError(c, "decode_categories", err)
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

	// Convert companyId string to ObjectID
	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
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
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
		helpers.ValidateBudget(req.Budget, "budget"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation
	req.Name = helpers.SanitizeString(req.Name)

	collection := database.GetCollection(categoryCollection)

	// Default to EXPENSE if not provided
	categoryType := req.Type
	if categoryType == "" {
		categoryType = models.Expense
	}

	category := models.Category{
		ID:        primitive.NewObjectID(),
		CompanyID: companyObjID,
		Name:      req.Name,
		Type:      categoryType,
		Color:     req.Color,
		Budget:    req.Budget,
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, category)
	if err != nil {
		return helpers.CategoryCreateFailed(c, err)
	}

	return c.Status(201).JSON(category)
}

func UpdateCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, objID)
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
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
		helpers.ValidateBudget(req.Budget, "budget"),
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
		return helpers.CategoryUpdateFailed(c, err)
	}

	var updatedCategory models.Category
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedCategory)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_updated_category", err)
	}

	return c.JSON(updatedCategory)
}

func DeleteCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateCategoryOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(categoryCollection)

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return helpers.CategoryDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}
