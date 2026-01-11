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

const quoteTemplateCollection = "quote_templates"

// GetQuoteTemplates lista todos os templates de orcamento de uma empresa
func GetQuoteTemplates(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.MissingRequiredParam(c, "companyId")
	}

	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(quoteTemplateCollection)

	cursor, err := collection.Find(ctx, bson.M{"companyId": companyObjID})
	if err != nil {
		return helpers.QuoteTemplateFetchFailed(c, err)
	}
	defer cursor.Close(ctx)

	var templates []models.QuoteTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return helpers.DatabaseError(c, "decode_templates", err)
	}

	if templates == nil {
		templates = []models.QuoteTemplate{}
	}

	return c.JSON(templates)
}

// GetQuoteTemplate busca um template por ID
func GetQuoteTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	template, err := helpers.ValidateQuoteTemplateOwnership(c, objID)
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

	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
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

	collection := database.GetCollection(quoteTemplateCollection)

	// Se este template deve ser o padrao, remover flag dos outros
	if req.IsDefault {
		_, err = collection.UpdateMany(ctx, bson.M{
			"companyId": companyObjID,
			"isDefault": true,
		}, bson.M{
			"$set": bson.M{"isDefault": false},
		})
		if err != nil {
			return helpers.DatabaseError(c, "update_existing_templates", err)
		}
	}

	// Default color
	if req.PrimaryColor == "" {
		req.PrimaryColor = "#78716c"
	}

	template := models.QuoteTemplate{
		ID:           primitive.NewObjectID(),
		CompanyID:    companyObjID,
		Name:         req.Name,
		HeaderText:   req.HeaderText,
		FooterText:   req.FooterText,
		TermsText:    req.TermsText,
		PrimaryColor: req.PrimaryColor,
		LogoURL:      req.LogoURL,
		IsDefault:    req.IsDefault,
		CreatedAt:    time.Now(),
	}

	_, err = collection.InsertOne(ctx, template)
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
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	existingTemplate, err := helpers.ValidateQuoteTemplateOwnership(c, objID)
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

	collection := database.GetCollection(quoteTemplateCollection)

	// Se este template deve ser o padrao, remover flag dos outros
	if req.IsDefault && !existingTemplate.IsDefault {
		_, err = collection.UpdateMany(ctx, bson.M{
			"companyId": existingTemplate.CompanyID,
			"isDefault": true,
			"_id":       bson.M{"$ne": objID},
		}, bson.M{
			"$set": bson.M{"isDefault": false},
		})
		if err != nil {
			return helpers.DatabaseError(c, "update_existing_templates", err)
		}
	}

	update := bson.M{
		"$set": bson.M{
			"name":         helpers.SanitizeString(req.Name),
			"headerText":   helpers.SanitizeString(req.HeaderText),
			"footerText":   helpers.SanitizeString(req.FooterText),
			"termsText":    helpers.SanitizeString(req.TermsText),
			"primaryColor": req.PrimaryColor,
			"logoUrl":      req.LogoURL,
			"isDefault":    req.IsDefault,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return helpers.QuoteTemplateUpdateFailed(c, err)
	}

	var updatedTemplate models.QuoteTemplate
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedTemplate)
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
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteTemplateOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(quoteTemplateCollection)

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return helpers.QuoteTemplateDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}
