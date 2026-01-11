package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

// Mutex to prevent race condition in quote number generation
var quoteNumberMutex sync.Mutex

// generateQuoteNumber generates a sequential number for quotes: ORC-2024-001
func generateQuoteNumber(ctx context.Context, companyID uuid.UUID) (string, error) {
	quoteNumberMutex.Lock()
	defer quoteNumberMutex.Unlock()

	year := time.Now().Year()
	prefix := fmt.Sprintf("ORC-%d-", year)

	// Find the last quote number for this company this year
	var lastNumber string
	err := database.QueryRow(ctx,
		`SELECT number FROM quotes WHERE company_id = $1 AND number LIKE $2 ORDER BY created_at DESC LIMIT 1`,
		companyID, prefix+"%").Scan(&lastNumber)

	nextNumber := 1
	if err == nil && lastNumber != "" {
		var lastNum int
		_, err := fmt.Sscanf(lastNumber, "ORC-%d-%d", new(int), &lastNum)
		if err == nil {
			nextNumber = lastNum + 1
		}
	}

	return fmt.Sprintf("ORC-%d-%03d", year, nextNumber), nil
}

// GetQuotes lists all quotes for a company
func GetQuotes(c *fiber.Ctx) error {
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

	// Build query with optional status filter
	query := `SELECT id, company_id, number, client_name, client_email, client_phone, client_document,
		client_address, client_city, client_state, client_zip_code, title, description,
		subtotal, discount, discount_type, total, status, valid_until, notes, template_id,
		created_at, updated_at FROM quotes WHERE company_id = $1`
	args := []interface{}{companyUUID}

	if status := c.Query("status"); status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return helpers.QuoteFetchFailed(c, err)
	}
	defer rows.Close()

	var quotes []models.Quote
	for rows.Next() {
		var q models.Quote
		var templateID *uuid.UUID
		if err := rows.Scan(&q.ID, &q.CompanyID, &q.Number, &q.ClientName, &q.ClientEmail,
			&q.ClientPhone, &q.ClientDocument, &q.ClientAddress, &q.ClientCity,
			&q.ClientState, &q.ClientZipCode, &q.Title, &q.Description,
			&q.Subtotal, &q.Discount, &q.DiscountType, &q.Total, &q.Status,
			&q.ValidUntil, &q.Notes, &templateID, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "decode_quotes", err)
		}
		q.TemplateID = templateID

		// Fetch items for this quote
		q.Items, _ = getQuoteItems(ctx, q.ID)

		quotes = append(quotes, q)
	}

	if quotes == nil {
		quotes = []models.Quote{}
	}

	return c.JSON(quotes)
}

// getQuoteItems fetches items for a quote
func getQuoteItems(ctx context.Context, quoteID uuid.UUID) ([]models.QuoteItem, error) {
	rows, err := database.Query(ctx,
		`SELECT id, quote_id, description, quantity, unit_price, total, category_id, created_at
		 FROM quote_items WHERE quote_id = $1`,
		quoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.QuoteItem
	for rows.Next() {
		var item models.QuoteItem
		if err := rows.Scan(&item.ID, &item.QuoteID, &item.Description, &item.Quantity,
			&item.UnitPrice, &item.Total, &item.CategoryID, &item.CreatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.QuoteItem{}
	}

	return items, nil
}

// GetQuote returns a single quote by ID
func GetQuote(c *fiber.Ctx) error {
	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	quote, err := helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	// Fetch items
	ctx := context.Background()
	quote.Items, _ = getQuoteItems(ctx, quote.ID)

	return c.JSON(quote)
}

// CreateQuote creates a new quote
func CreateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	var req models.CreateQuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateMaxLength(req.ClientName, "clientName", 100),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateMaxLength(req.Title, "title", 200),
		helpers.ValidateMaxLength(req.Description, "description", 1000),
		helpers.ValidateMaxLength(req.Notes, "notes", 2000),
		helpers.ValidateNoScriptTags(req.ClientName, "clientName"),
		helpers.ValidateNoScriptTags(req.Title, "title"),
		helpers.ValidateDiscountValue(req.Discount, req.DiscountType, "discount"),
	)

	if len(req.Items) == 0 {
		errors = append(errors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

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

	// Calculate totals
	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.Quantity * item.UnitPrice
	}

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
	validUntil := time.Now().AddDate(0, 0, 30)
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	// Parse templateID
	var templateID *uuid.UUID
	if req.TemplateID != "" {
		tid, err := uuid.Parse(req.TemplateID)
		if err == nil {
			templateID = &tid
		}
	}

	// Generate quote number
	quoteNumber, err := generateQuoteNumber(ctx, companyUUID)
	if err != nil {
		return helpers.QuoteNumberGenerationFailed(c, err)
	}

	now := time.Now()
	quoteID := uuid.New()

	// Use transaction for atomic insert
	err = database.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert quote
		_, err := tx.Exec(ctx,
			`INSERT INTO quotes (id, company_id, number, client_name, client_email, client_phone, client_document,
			 client_address, client_city, client_state, client_zip_code, title, description,
			 subtotal, discount, discount_type, total, status, valid_until, notes, template_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`,
			quoteID, companyUUID, quoteNumber, req.ClientName, req.ClientEmail, req.ClientPhone, req.ClientDocument,
			req.ClientAddress, req.ClientCity, req.ClientState, req.ClientZipCode, req.Title, req.Description,
			subtotal, req.Discount, discountType, total, models.QuoteDraft, validUntil, req.Notes, templateID, now, now)
		if err != nil {
			return err
		}

		// Insert items
		for _, item := range req.Items {
			itemTotal := item.Quantity * item.UnitPrice
			var categoryID *uuid.UUID
			if item.CategoryID != "" {
				cid, err := uuid.Parse(item.CategoryID)
				if err == nil {
					categoryID = &cid
				}
			}

			_, err := tx.Exec(ctx,
				`INSERT INTO quote_items (id, quote_id, description, quantity, unit_price, total, category_id, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				uuid.New(), quoteID, helpers.SanitizeString(item.Description), item.Quantity, item.UnitPrice, itemTotal, categoryID, now)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return helpers.QuoteCreateFailed(c, err)
	}

	// Build response
	quote := models.Quote{
		ID:             quoteID,
		CompanyID:      companyUUID,
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

	quote.Items, _ = getQuoteItems(ctx, quoteID)

	return c.Status(201).JSON(quote)
}

// UpdateQuote updates an existing quote
func UpdateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	existingQuote, err := helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	if existingQuote.Status == models.QuoteExecuted {
		return helpers.QuoteAlreadyExecuted(c)
	}

	var req models.UpdateQuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.ClientName, "clientName"),
		helpers.ValidateRequired(req.Title, "title"),
		helpers.ValidateDiscountValue(req.Discount, req.DiscountType, "discount"),
	)

	if len(req.Items) == 0 {
		errors = append(errors, helpers.ValidationError{
			Field:   "items",
			Message: "O orcamento deve ter pelo menos um item",
		})
	}

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Calculate totals
	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.Quantity * item.UnitPrice
	}

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

	validUntil := existingQuote.ValidUntil
	if req.ValidUntil != "" {
		parsed, err := time.Parse("2006-01-02", req.ValidUntil)
		if err == nil {
			validUntil = parsed
		}
	}

	var templateID *uuid.UUID
	if req.TemplateID != "" {
		tid, err := uuid.Parse(req.TemplateID)
		if err == nil {
			templateID = &tid
		}
	}

	now := time.Now()

	// Use transaction
	err = database.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Update quote
		_, err := tx.Exec(ctx,
			`UPDATE quotes SET client_name = $1, client_email = $2, client_phone = $3, client_document = $4,
			 client_address = $5, client_city = $6, client_state = $7, client_zip_code = $8, title = $9,
			 description = $10, subtotal = $11, discount = $12, discount_type = $13, total = $14,
			 valid_until = $15, notes = $16, template_id = $17, updated_at = $18 WHERE id = $19`,
			helpers.SanitizeString(req.ClientName), helpers.SanitizeString(req.ClientEmail),
			helpers.SanitizeString(req.ClientPhone), helpers.SanitizeString(req.ClientDocument),
			helpers.SanitizeString(req.ClientAddress), helpers.SanitizeString(req.ClientCity),
			helpers.SanitizeString(req.ClientState), helpers.SanitizeString(req.ClientZipCode),
			helpers.SanitizeString(req.Title), helpers.SanitizeString(req.Description),
			subtotal, req.Discount, discountType, total, validUntil, helpers.SanitizeString(req.Notes),
			templateID, now, quoteID)
		if err != nil {
			return err
		}

		// Delete old items
		_, err = tx.Exec(ctx, `DELETE FROM quote_items WHERE quote_id = $1`, quoteID)
		if err != nil {
			return err
		}

		// Insert new items
		for _, item := range req.Items {
			itemTotal := item.Quantity * item.UnitPrice
			var categoryID *uuid.UUID
			if item.CategoryID != "" {
				cid, err := uuid.Parse(item.CategoryID)
				if err == nil {
					categoryID = &cid
				}
			}

			_, err := tx.Exec(ctx,
				`INSERT INTO quote_items (id, quote_id, description, quantity, unit_price, total, category_id, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				uuid.New(), quoteID, helpers.SanitizeString(item.Description), item.Quantity, item.UnitPrice, itemTotal, categoryID, now)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return helpers.QuoteUpdateFailed(c, err)
	}

	// Fetch updated quote
	updatedQuote, _ := helpers.ValidateQuoteOwnership(c, quoteID)
	updatedQuote.Items, _ = getQuoteItems(ctx, quoteID)

	return c.JSON(updatedQuote)
}

// DeleteQuote deletes a quote
func DeleteQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	// CASCADE will handle quote_items
	_, err = database.Pool.Exec(ctx, `DELETE FROM quotes WHERE id = $1`, quoteID)
	if err != nil {
		return helpers.QuoteDeleteFailed(c, err)
	}

	return c.SendStatus(204)
}

// DuplicateQuote duplicates an existing quote
func DuplicateQuote(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	originalQuote, err := helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	originalQuote.Items, _ = getQuoteItems(ctx, originalQuote.ID)

	// Generate new quote number
	quoteNumber, err := generateQuoteNumber(ctx, originalQuote.CompanyID)
	if err != nil {
		return helpers.QuoteNumberGenerationFailed(c, err)
	}

	now := time.Now()
	newQuoteID := uuid.New()

	// Use transaction
	err = database.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert new quote
		_, err := tx.Exec(ctx,
			`INSERT INTO quotes (id, company_id, number, client_name, client_email, client_phone, client_document,
			 client_address, client_city, client_state, client_zip_code, title, description,
			 subtotal, discount, discount_type, total, status, valid_until, notes, template_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`,
			newQuoteID, originalQuote.CompanyID, quoteNumber, originalQuote.ClientName, originalQuote.ClientEmail,
			originalQuote.ClientPhone, originalQuote.ClientDocument, originalQuote.ClientAddress, originalQuote.ClientCity,
			originalQuote.ClientState, originalQuote.ClientZipCode, originalQuote.Title+" (Copia)", originalQuote.Description,
			originalQuote.Subtotal, originalQuote.Discount, originalQuote.DiscountType, originalQuote.Total,
			models.QuoteDraft, now.AddDate(0, 0, 30), originalQuote.Notes, originalQuote.TemplateID, now, now)
		if err != nil {
			return err
		}

		// Copy items
		for _, item := range originalQuote.Items {
			_, err := tx.Exec(ctx,
				`INSERT INTO quote_items (id, quote_id, description, quantity, unit_price, total, category_id, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				uuid.New(), newQuoteID, item.Description, item.Quantity, item.UnitPrice, item.Total, item.CategoryID, now)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return helpers.QuoteDuplicateFailed(c, err)
	}

	// Build response
	newQuote := models.Quote{
		ID:             newQuoteID,
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
		Title:          originalQuote.Title + " (Copia)",
		Description:    originalQuote.Description,
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

	newQuote.Items, _ = getQuoteItems(ctx, newQuoteID)

	return c.Status(201).JSON(newQuote)
}

// UpdateQuoteStatus updates the status of a quote
func UpdateQuoteStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	var req models.UpdateQuoteStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	validStatuses := map[string]bool{
		string(models.QuoteDraft):    true,
		string(models.QuoteSent):     true,
		string(models.QuoteApproved): true,
		string(models.QuoteRejected): true,
		string(models.QuoteExecuted): true,
	}

	if !validStatuses[req.Status] {
		return helpers.BadRequest(c, helpers.ErrCodeValidationFailed, "Status invalido", helpers.ErrorDetails{
			"field":         "status",
			"allowedValues": []string{"DRAFT", "SENT", "APPROVED", "REJECTED", "EXECUTED"},
		})
	}

	now := time.Now()
	_, err = database.Pool.Exec(ctx,
		`UPDATE quotes SET status = $1, updated_at = $2 WHERE id = $3`,
		req.Status, now, quoteID)
	if err != nil {
		return helpers.QuoteUpdateFailed(c, err)
	}

	// Fetch updated quote
	updatedQuote, _ := helpers.ValidateQuoteOwnership(c, quoteID)
	updatedQuote.Items, _ = getQuoteItems(ctx, quoteID)

	return c.JSON(updatedQuote)
}

// GetQuoteComparison returns comparison between quoted and executed amounts
func GetQuoteComparison(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id := c.Params("id")
	quoteID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	quote, err := helpers.ValidateQuoteOwnership(c, quoteID)
	if err != nil {
		return err
	}

	quote.Items, _ = getQuoteItems(ctx, quote.ID)

	var comparisonItems []models.QuoteComparisonItem
	var executedTotal float64

	for _, item := range quote.Items {
		compItem := models.QuoteComparisonItem{
			Description: item.Description,
			Quoted:      item.Total,
			Executed:    0,
			Variance:    0,
		}

		if item.CategoryID != nil {
			compItem.CategoryID = item.CategoryID.String()

			// Fetch transactions for this category
			var executed float64
			database.QueryRow(ctx,
				`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE company_id = $1 AND category = $2`,
				quote.CompanyID, item.CategoryID.String()).Scan(&executed)
			compItem.Executed = executed
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
		QuoteID:         quote.ID.String(),
		QuotedTotal:     quote.Total,
		ExecutedTotal:   executedTotal,
		Variance:        variance,
		VariancePercent: variancePercent,
		Items:           comparisonItems,
	}

	return c.JSON(comparison)
}
