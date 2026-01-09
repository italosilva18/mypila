package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

const quoteCollection = "quotes"

// Mutex to prevent race condition in quote number generation
var quoteNumberMutex sync.Mutex

// generateQuoteNumber gera número sequencial para orçamentos: ORC-2024-001
// Uses mutex to prevent duplicate numbers from concurrent requests
func generateQuoteNumber(ctx context.Context, companyID primitive.ObjectID) (string, error) {
	quoteNumberMutex.Lock()
	defer quoteNumberMutex.Unlock()
	collection := database.GetCollection(quoteCollection)

	year := time.Now().Year()

	// Buscar o último orçamento do ano para esta empresa
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	filter := bson.M{
		"companyId": companyID,
		"number": bson.M{
			"$regex": fmt.Sprintf("^ORC-%d-", year),
		},
	}

	var lastQuote models.Quote
	err := collection.FindOne(ctx, filter, opts).Decode(&lastQuote)

	nextNumber := 1
	if err == nil {
		// Extrair número do último orçamento
		var lastNum int
		_, err := fmt.Sscanf(lastQuote.Number, "ORC-%d-%d", new(int), &lastNum)
		if err == nil {
			nextNumber = lastNum + 1
		}
	}

	return fmt.Sprintf("ORC-%d-%03d", year, nextNumber), nil
}

// GetQuotes lista todos os orçamentos de uma empresa
func GetQuotes(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Company ID is required"})
	}

	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(quoteCollection)

	// Filtro opcional por status
	filter := bson.M{"companyId": companyObjID}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	// Ordenar por data de criação decrescente
	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch quotes"})
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err = cursor.All(ctx, &quotes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse quotes"})
	}

	if quotes == nil {
		quotes = []models.Quote{}
	}

	return c.JSON(quotes)
}

// GetQuote busca um orçamento por ID
func GetQuote(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	quote, err := helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	return c.JSON(quote)
}

// CreateQuote cria um novo orçamento
func CreateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	companyId := c.Query("companyId")
	if companyId == "" {
		return helpers.SendValidationError(c, "companyId", "ID da empresa é obrigatório")
	}

	companyObjID, err := primitive.ObjectIDFromHex(companyId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid company ID format"})
	}

	// Validate company ownership
	_, err = helpers.ValidateCompanyOwnership(c, companyObjID)
	if err != nil {
		return err
	}

	var req models.CreateQuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateMaxLength(req.ClientName, "clientName", 100),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateMaxLength(req.Title, "title", 200),
		helpers.ValidateMaxLength(req.Description, "description", 1000),
		helpers.ValidateMaxLength(req.Notes, "notes", 2000),
		helpers.ValidateNoScriptTags(req.ClientName, "clientName"),
		helpers.ValidateNoScriptTags(req.Title, "title"),
		helpers.ValidateMongoInjection(req.ClientName, "clientName"),
		helpers.ValidateMongoInjection(req.Title, "title"),
		helpers.ValidateNonNegative(req.Discount, "discount"),
	)

	if len(req.Items) == 0 {
		errors = append(errors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

	// Validar cada item do orcamento
	for i, item := range req.Items {
		if item.Description == "" {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].description", i),
				Message: "Descricao do item e obrigatoria",
			})
		}
		if item.Quantity <= 0 {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].quantity", i),
				Message: "Quantidade deve ser maior que zero",
			})
		}
		if item.UnitPrice <= 0 {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].unitPrice", i),
				Message: "Preco unitario deve ser maior que zero",
			})
		}
	}

	// Validar desconto percentual
	if req.DiscountType == "PERCENT" && req.Discount > 100 {
		errors = append(errors, helpers.ValidationError{
			Field:   "discount",
			Message: "Desconto percentual deve ser no maximo 100%",
		})
	}

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize inputs
	req.ClientName = helpers.SanitizeString(req.ClientName)
	req.ClientEmail = helpers.SanitizeString(req.ClientEmail)
	req.ClientPhone = helpers.SanitizeString(req.ClientPhone)
	req.ClientDocument = helpers.SanitizeString(req.ClientDocument)
	req.ClientAddress = helpers.SanitizeString(req.ClientAddress)
	req.ClientCity = helpers.SanitizeString(req.ClientCity)
	req.ClientState = helpers.SanitizeString(req.ClientState)
	req.ClientZipCode = helpers.SanitizeString(req.ClientZipCode)
	req.Title = helpers.SanitizeString(req.Title)
	req.Description = helpers.SanitizeString(req.Description)
	req.Notes = helpers.SanitizeString(req.Notes)

	// Gerar número do orçamento
	quoteNumber, err := generateQuoteNumber(ctx, companyObjID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar número do orçamento"})
	}

	// Processar itens e calcular totais
	var items []models.QuoteItem
	var subtotal float64

	for _, item := range req.Items {
		itemTotal := item.Quantity * item.UnitPrice
		subtotal += itemTotal

		quoteItem := models.QuoteItem{
			ID:          primitive.NewObjectID(),
			Description: helpers.SanitizeString(item.Description),
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       itemTotal,
		}

		if item.CategoryID != "" {
			catID, err := primitive.ObjectIDFromHex(item.CategoryID)
			if err == nil {
				quoteItem.CategoryID = catID
			}
		}

		items = append(items, quoteItem)
	}

	// Calcular desconto
	var total float64
	discountType := models.DiscountType(req.DiscountType)
	if discountType == "" {
		discountType = models.DiscountValue
	}

	if discountType == models.DiscountPercent {
		total = subtotal - (subtotal * req.Discount / 100)
	} else {
		total = subtotal - req.Discount
	}

	// Parse validUntil
	validUntil := time.Now().AddDate(0, 0, 30) // Default: 30 dias
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	// Parse templateID
	var templateID primitive.ObjectID
	if req.TemplateID != "" {
		tid, err := primitive.ObjectIDFromHex(req.TemplateID)
		if err == nil {
			templateID = tid
		}
	}

	collection := database.GetCollection(quoteCollection)

	now := time.Now()
	quote := models.Quote{
		ID:             primitive.NewObjectID(),
		CompanyID:      companyObjID,
		Number:         quoteNumber,
		ClientName:     req.ClientName,
		ClientEmail:    req.ClientEmail,
		ClientPhone:    req.ClientPhone,
		ClientDocument: req.ClientDocument,
		ClientAddress:  req.ClientAddress,
		ClientCity:     req.ClientCity,
		ClientState:    req.ClientState,
		ClientZipCode:  req.ClientZipCode,
		Title:          req.Title,
		Description:    req.Description,
		Items:          items,
		Subtotal:       subtotal,
		Discount:       req.Discount,
		DiscountType:   discountType,
		Total:          total,
		Status:         models.QuoteDraft,
		ValidUntil:     validUntil,
		Notes:          req.Notes,
		TemplateID:     templateID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = collection.InsertOne(ctx, quote)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar orçamento"})
	}

	return c.Status(201).JSON(quote)
}

// UpdateQuote atualiza um orçamento existente
func UpdateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	existingQuote, err := helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	// Não permitir edição de orçamentos executados
	if existingQuote.Status == models.QuoteExecuted {
		return c.Status(400).JSON(fiber.Map{"error": "Não é possível editar um orçamento executado"})
	}

	var req models.UpdateQuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateMaxLength(req.ClientName, "clientName", 100),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateMaxLength(req.Title, "title", 200),
		helpers.ValidateNoScriptTags(req.ClientName, "clientName"),
		helpers.ValidateNoScriptTags(req.Title, "title"),
		helpers.ValidateNonNegative(req.Discount, "discount"),
	)

	if len(req.Items) == 0 {
		errors = append(errors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

	// Validar cada item do orcamento
	for i, item := range req.Items {
		if item.Description == "" {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].description", i),
				Message: "Descricao do item e obrigatoria",
			})
		}
		if item.Quantity <= 0 {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].quantity", i),
				Message: "Quantidade deve ser maior que zero",
			})
		}
		if item.UnitPrice <= 0 {
			errors = append(errors, helpers.ValidationError{
				Field:   fmt.Sprintf("items[%d].unitPrice", i),
				Message: "Preco unitario deve ser maior que zero",
			})
		}
	}

	// Validar desconto percentual
	if req.DiscountType == "PERCENT" && req.Discount > 100 {
		errors = append(errors, helpers.ValidationError{
			Field:   "discount",
			Message: "Desconto percentual deve ser no maximo 100%",
		})
	}

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Processar itens e calcular totais
	var items []models.QuoteItem
	var subtotal float64

	for _, item := range req.Items {
		itemTotal := item.Quantity * item.UnitPrice
		subtotal += itemTotal

		quoteItem := models.QuoteItem{
			ID:          primitive.NewObjectID(),
			Description: helpers.SanitizeString(item.Description),
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       itemTotal,
		}

		if item.CategoryID != "" {
			catID, err := primitive.ObjectIDFromHex(item.CategoryID)
			if err == nil {
				quoteItem.CategoryID = catID
			}
		}

		items = append(items, quoteItem)
	}

	// Calcular desconto
	var total float64
	discountType := models.DiscountType(req.DiscountType)
	if discountType == "" {
		discountType = models.DiscountValue
	}

	if discountType == models.DiscountPercent {
		total = subtotal - (subtotal * req.Discount / 100)
	} else {
		total = subtotal - req.Discount
	}

	// Parse validUntil
	validUntil := existingQuote.ValidUntil
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	// Parse templateID
	var templateID primitive.ObjectID
	if req.TemplateID != "" {
		tid, err := primitive.ObjectIDFromHex(req.TemplateID)
		if err == nil {
			templateID = tid
		}
	}

	collection := database.GetCollection(quoteCollection)

	update := bson.M{
		"$set": bson.M{
			"clientName":     helpers.SanitizeString(req.ClientName),
			"clientEmail":    helpers.SanitizeString(req.ClientEmail),
			"clientPhone":    helpers.SanitizeString(req.ClientPhone),
			"clientDocument": helpers.SanitizeString(req.ClientDocument),
			"clientAddress":  helpers.SanitizeString(req.ClientAddress),
			"clientCity":     helpers.SanitizeString(req.ClientCity),
			"clientState":    helpers.SanitizeString(req.ClientState),
			"clientZipCode":  helpers.SanitizeString(req.ClientZipCode),
			"title":          helpers.SanitizeString(req.Title),
			"description":    helpers.SanitizeString(req.Description),
			"items":          items,
			"subtotal":       subtotal,
			"discount":       req.Discount,
			"discountType":   discountType,
			"total":          total,
			"validUntil":     validUntil,
			"notes":          helpers.SanitizeString(req.Notes),
			"templateId":     templateID,
			"updatedAt":      time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar orçamento"})
	}

	var updatedQuote models.Quote
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedQuote)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar orçamento atualizado"})
	}

	return c.JSON(updatedQuote)
}

// DeleteQuote exclui um orçamento
func DeleteQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	collection := database.GetCollection(quoteCollection)

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete quote"})
	}

	return c.SendStatus(204)
}

// DuplicateQuote duplica um orçamento existente
func DuplicateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	originalQuote, err := helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	// Gerar novo número
	quoteNumber, err := generateQuoteNumber(ctx, originalQuote.CompanyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar número do orçamento"})
	}

	// Criar novos IDs para os itens
	var newItems []models.QuoteItem
	for _, item := range originalQuote.Items {
		newItem := item
		newItem.ID = primitive.NewObjectID()
		newItems = append(newItems, newItem)
	}

	now := time.Now()
	newQuote := models.Quote{
		ID:             primitive.NewObjectID(),
		CompanyID:      originalQuote.CompanyID,
		Number:         quoteNumber,
		ClientName:     originalQuote.ClientName,
		ClientEmail:    originalQuote.ClientEmail,
		ClientPhone:    originalQuote.ClientPhone,
		ClientDocument: originalQuote.ClientDocument,
		ClientAddress:  originalQuote.ClientAddress,
		ClientCity:     originalQuote.ClientCity,
		ClientState:    originalQuote.ClientState,
		ClientZipCode:  originalQuote.ClientZipCode,
		Title:          originalQuote.Title + " (Cópia)",
		Description:    originalQuote.Description,
		Items:          newItems,
		Subtotal:       originalQuote.Subtotal,
		Discount:       originalQuote.Discount,
		DiscountType:   originalQuote.DiscountType,
		Total:          originalQuote.Total,
		Status:         models.QuoteDraft,
		ValidUntil:     now.AddDate(0, 0, 30),
		Notes:          originalQuote.Notes,
		TemplateID:     originalQuote.TemplateID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	collection := database.GetCollection(quoteCollection)
	_, err = collection.InsertOne(ctx, newQuote)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao duplicar orçamento"})
	}

	return c.Status(201).JSON(newQuote)
}

// UpdateQuoteStatus atualiza o status de um orçamento
func UpdateQuoteStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	var req models.UpdateQuoteStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Validar status
	validStatuses := map[string]bool{
		string(models.QuoteDraft):    true,
		string(models.QuoteSent):     true,
		string(models.QuoteApproved): true,
		string(models.QuoteRejected): true,
		string(models.QuoteExecuted): true,
	}

	if !validStatuses[req.Status] {
		return helpers.SendValidationError(c, "status", "Status inválido")
	}

	collection := database.GetCollection(quoteCollection)

	update := bson.M{
		"$set": bson.M{
			"status":    models.QuoteStatus(req.Status),
			"updatedAt": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar status"})
	}

	var updatedQuote models.Quote
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedQuote)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar orçamento atualizado"})
	}

	return c.JSON(updatedQuote)
}

// GetQuoteComparison retorna comparativo orçado vs realizado
func GetQuoteComparison(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.SendValidationError(c, "id", "Formato de ID inválido")
	}

	// Validate ownership e buscar orçamento
	quote, err := helpers.ValidateQuoteOwnership(c, objID)
	if err != nil {
		return err
	}

	// Buscar transações por categoria
	transCollection := database.GetCollection("transactions")

	var comparisonItems []models.QuoteComparisonItem
	var executedTotal float64

	for _, item := range quote.Items {
		compItem := models.QuoteComparisonItem{
			Description: item.Description,
			Quoted:      item.Total,
			Executed:    0,
			Variance:    0,
		}

		// Se o item tem categoria, buscar transações dessa categoria
		if !item.CategoryID.IsZero() {
			compItem.CategoryID = item.CategoryID.Hex()

			// Buscar transações da categoria no período do orçamento
			filter := bson.M{
				"companyId": quote.CompanyID,
				"category":  item.CategoryID.Hex(),
			}

			cursor, err := transCollection.Find(ctx, filter)
			if err == nil {
				var transactions []models.Transaction
				if err = cursor.All(ctx, &transactions); err == nil {
					for _, t := range transactions {
						compItem.Executed += t.Amount
					}
				}
				cursor.Close(ctx)
			}
		}

		compItem.Variance = compItem.Quoted - compItem.Executed
		executedTotal += compItem.Executed
		comparisonItems = append(comparisonItems, compItem)
	}

	variance := quote.Total - executedTotal
	variancePercent := 0.0
	if quote.Total > 0 {
		variancePercent = (variance / quote.Total) * 100
	}

	comparison := models.QuoteComparison{
		QuoteID:         quote.ID.Hex(),
		QuotedTotal:     quote.Total,
		ExecutedTotal:   executedTotal,
		Variance:        variance,
		VariancePercent: variancePercent,
		Items:           comparisonItems,
	}

	return c.JSON(comparison)
}
