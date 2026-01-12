package handlers

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

// GetAllTransactions returns paginated transactions (filtered by company with ownership validation)
func GetAllTransactions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse pagination parameters
	page := parsePage(c.Query("page", "1"))
	limit := parseLimit(c.Query("limit", "50"))
	offset := (page - 1) * limit

	var total int64
	var rows interface{ Close() }

	companyIDStr := c.Query("companyId")

	if companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return helpers.InvalidIDFormat(c, "companyId")
		}

		// Validate ownership
		_, err = helpers.ValidateCompanyOwnership(c, companyID)
		if err != nil {
			return err
		}

		// Count total
		err = database.QueryRow(ctx,
			`SELECT COUNT(*) FROM transactions WHERE company_id = $1`,
			companyID).Scan(&total)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}

		// Fetch transactions
		pgRows, err := database.Query(ctx,
			`SELECT id, company_id, category, description, amount, month, year, due_day, status, created_at, updated_at
			 FROM transactions WHERE company_id = $1
			 ORDER BY year DESC, month DESC, due_day DESC
			 LIMIT $2 OFFSET $3`,
			companyID, limit, offset)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}
		rows = pgRows
	} else {
		// Get all transactions for all user's companies
		// Count total
		err = database.QueryRow(ctx,
			`SELECT COUNT(*) FROM transactions t
			 INNER JOIN companies c ON t.company_id = c.id
			 WHERE c.user_id = $1`,
			userID).Scan(&total)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}

		// Fetch transactions
		pgRows, err := database.Query(ctx,
			`SELECT t.id, t.company_id, t.category, t.description, t.amount, t.month, t.year, t.due_day, t.status, t.created_at, t.updated_at
			 FROM transactions t
			 INNER JOIN companies c ON t.company_id = c.id
			 WHERE c.user_id = $1
			 ORDER BY t.year DESC, t.month DESC, t.due_day DESC
			 LIMIT $2 OFFSET $3`,
			userID, limit, offset)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}
		rows = pgRows
	}

	defer rows.Close()

	// Scan transactions
	var transactions []models.Transaction
	pgRows := rows.(interface {
		Close()
		Next() bool
		Scan(dest ...interface{}) error
	})

	for pgRows.Next() {
		var t models.Transaction
		var description *string
		var dueDay *int
		if err := pgRows.Scan(&t.ID, &t.CompanyID, &t.Category, &description, &t.Amount, &t.Month, &t.Year, &dueDay, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "scan_transaction", err)
		}
		if description != nil {
			t.Description = *description
		}
		if dueDay != nil {
			t.DueDay = *dueDay
		} else {
			t.DueDay = 1
		}
		transactions = append(transactions, t)
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedTransactions{
		Data: transactions,
		Pagination: models.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return c.JSON(response)
}

// parsePage parses the page query parameter with validation
func parsePage(pageStr string) int {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// parseLimit parses the limit query parameter with validation (max 100)
func parseLimit(limitStr string) int {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 50 // Default
	}
	if limit > 100 {
		return 100 // Maximum
	}
	return limit
}

// GetTransaction returns a single transaction by ID (ownership validated)
func GetTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	transaction, err := helpers.ValidateTransactionOwnership(c, transactionID)
	if err != nil {
		return err
	}

	// Fetch full transaction
	var t models.Transaction
	var description *string
	var dueDay *int
	err = database.QueryRow(ctx,
		`SELECT id, company_id, category, description, amount, month, year, due_day, status, created_at, updated_at
		 FROM transactions WHERE id = $1`,
		transactionID).Scan(&t.ID, &t.CompanyID, &t.Category, &description, &t.Amount, &t.Month, &t.Year, &dueDay, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return helpers.TransactionNotFound(c)
	}
	if description != nil {
		t.Description = *description
	}
	if dueDay != nil {
		t.DueDay = *dueDay
	} else {
		t.DueDay = 1
	}

	_ = transaction // ownership validation returns the transaction

	return c.JSON(t)
}

// CreateTransaction creates a new transaction
func CreateTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateAmount(req.Amount, "amount"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateMonth(req.Month),
		helpers.ValidateRange(req.Year, 2000, 2100, "year"),
		helpers.ValidateStatus(string(req.Status)),
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

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return helpers.InvalidIDFormat(c, "companyId")
	}

	// Validate company ownership before creating transaction
	_, err = helpers.ValidateCompanyOwnership(c, companyID)
	if err != nil {
		return err
	}

	// Default due day to 1 if not provided
	dueDay := req.DueDay
	if dueDay < 1 || dueDay > 31 {
		dueDay = 1
	}

	now := time.Now()
	transaction := models.Transaction{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Month:       req.Month,
		Year:        req.Year,
		DueDay:      dueDay,
		Amount:      req.Amount,
		Category:    req.Category,
		Status:      req.Status,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO transactions (id, company_id, category, description, amount, month, year, due_day, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transaction.ID, transaction.CompanyID, transaction.Category, transaction.Description,
		transaction.Amount, transaction.Month, transaction.Year, transaction.DueDay, transaction.Status,
		transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		return helpers.TransactionCreateFailed(c, err)
	}

	return c.Status(201).JSON(transaction)
}

// UpdateTransaction updates an existing transaction (ownership validated)
func UpdateTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateTransactionOwnership(c, transactionID)
	if err != nil {
		return err
	}

	var req models.UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateAmount(req.Amount, "amount"),
		helpers.ValidateMaxLength(req.Description, "description", 200),
		helpers.ValidateRequired(req.Category, "category"),
		helpers.ValidateMonth(req.Month),
		helpers.ValidateRange(req.Year, 2000, 2100, "year"),
		helpers.ValidateStatus(string(req.Status)),
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

	// Default due day to 1 if not provided
	dueDay := req.DueDay
	if dueDay < 1 || dueDay > 31 {
		dueDay = 1
	}

	now := time.Now()
	result, err := database.Pool.Exec(ctx,
		`UPDATE transactions SET month = $1, year = $2, due_day = $3, amount = $4, category = $5, status = $6, description = $7, updated_at = $8
		 WHERE id = $9`,
		req.Month, req.Year, dueDay, req.Amount, req.Category, req.Status, req.Description, now, transactionID)
	if err != nil {
		return helpers.TransactionUpdateFailed(c, err)
	}

	if result.RowsAffected() == 0 {
		return helpers.TransactionNotFound(c)
	}

	// Fetch updated transaction
	var t models.Transaction
	var description *string
	var dueDayPtr *int
	err = database.QueryRow(ctx,
		`SELECT id, company_id, category, description, amount, month, year, due_day, status, created_at, updated_at
		 FROM transactions WHERE id = $1`,
		transactionID).Scan(&t.ID, &t.CompanyID, &t.Category, &description, &t.Amount, &t.Month, &t.Year, &dueDayPtr, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return helpers.DatabaseError(c, "fetch_transaction", err)
	}
	if description != nil {
		t.Description = *description
	}
	if dueDayPtr != nil {
		t.DueDay = *dueDayPtr
	} else {
		t.DueDay = 1
	}

	return c.JSON(t)
}

// DeleteTransaction deletes a transaction (ownership validated)
func DeleteTransaction(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership
	_, err = helpers.ValidateTransactionOwnership(c, transactionID)
	if err != nil {
		return err
	}

	result, err := database.Pool.Exec(ctx,
		`DELETE FROM transactions WHERE id = $1`,
		transactionID)
	if err != nil {
		return helpers.TransactionDeleteFailed(c, err)
	}

	if result.RowsAffected() == 0 {
		return helpers.TransactionNotFound(c)
	}

	return c.JSON(fiber.Map{"message": "Transacao excluida com sucesso"})
}

// GetStats returns financial statistics (ownership validated)
func GetStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	companyIDStr := c.Query("companyId")

	var stats models.Stats

	if companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return helpers.InvalidIDFormat(c, "companyId")
		}

		// Validate ownership
		_, err = helpers.ValidateCompanyOwnership(c, companyID)
		if err != nil {
			return err
		}

		// Get stats for specific company
		err = database.QueryRow(ctx,
			`SELECT
				COALESCE(SUM(CASE WHEN status = 'PAGO' THEN amount ELSE 0 END), 0) as paid,
				COALESCE(SUM(CASE WHEN status = 'ABERTO' THEN amount ELSE 0 END), 0) as open,
				COALESCE(SUM(amount), 0) as total
			 FROM transactions WHERE company_id = $1`,
			companyID).Scan(&stats.Paid, &stats.Open, &stats.Total)
	} else {
		// Get stats for all user's companies
		err = database.QueryRow(ctx,
			`SELECT
				COALESCE(SUM(CASE WHEN t.status = 'PAGO' THEN t.amount ELSE 0 END), 0) as paid,
				COALESCE(SUM(CASE WHEN t.status = 'ABERTO' THEN t.amount ELSE 0 END), 0) as open,
				COALESCE(SUM(t.amount), 0) as total
			 FROM transactions t
			 INNER JOIN companies c ON t.company_id = c.id
			 WHERE c.user_id = $1`,
			userID).Scan(&stats.Paid, &stats.Open, &stats.Total)
	}

	if err != nil {
		return helpers.TransactionFetchFailed(c, err)
	}

	return c.JSON(stats)
}

// ToggleStatus toggles the status of a transaction (ownership validated)
func ToggleStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return helpers.InvalidIDFormat(c, "id")
	}

	// Validate ownership and get current transaction
	transaction, err := helpers.ValidateTransactionOwnership(c, transactionID)
	if err != nil {
		return err
	}

	// Toggle status
	newStatus := models.StatusOpen
	if transaction.Status == models.StatusOpen {
		newStatus = models.StatusPaid
	}

	now := time.Now()
	_, err = database.Pool.Exec(ctx,
		`UPDATE transactions SET status = $1, updated_at = $2 WHERE id = $3`,
		newStatus, now, transactionID)
	if err != nil {
		return helpers.TransactionUpdateFailed(c, err)
	}

	transaction.Status = newStatus
	transaction.UpdatedAt = now
	return c.JSON(transaction)
}

// SeedTransactions seeds initial data if collection is empty
func SeedTransactions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if already has data
	var count int64
	err := database.QueryRow(ctx, "SELECT COUNT(*) FROM transactions").Scan(&count)
	if err != nil {
		return helpers.DatabaseError(c, "count_transactions", err)
	}

	if count > 0 {
		return c.JSON(fiber.Map{"message": "Data already seeded", "count": count})
	}

	// Get current user ID for seeding
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Check for company
	var companyID uuid.UUID
	err = database.QueryRow(ctx, "SELECT id FROM companies WHERE user_id = $1 LIMIT 1", userID).Scan(&companyID)
	if err != nil {
		// Create company
		companyID = uuid.New()
		now := time.Now()
		_, err = database.Pool.Exec(ctx,
			`INSERT INTO companies (id, user_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
			companyID, userID, "MyPila", now, now)
		if err != nil {
			return helpers.CompanyCreateFailed(c, err)
		}
	}

	// Insert sample transactions
	now := time.Now()
	sampleData := []models.Transaction{
		{ID: uuid.New(), CompanyID: companyID, Month: "Janeiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), CompanyID: companyID, Month: "Fevereiro", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), CompanyID: companyID, Month: "Marco", Year: 2024, Amount: 3500, Category: models.CategorySalary, Status: models.StatusPaid, CreatedAt: now, UpdatedAt: now},
	}

	for _, t := range sampleData {
		_, err = database.Pool.Exec(ctx,
			`INSERT INTO transactions (id, company_id, category, description, amount, month, year, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			t.ID, t.CompanyID, t.Category, t.Description, t.Amount, t.Month, t.Year, t.Status, t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return helpers.DatabaseError(c, "seed_transactions", err)
		}
	}

	return c.Status(201).JSON(fiber.Map{"message": "Data seeded successfully", "count": len(sampleData)})
}

// GetUpcomingTransactions returns transactions that are due soon (within specified days)
func GetUpcomingTransactions(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user ID
	userID, err := helpers.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// Parse days parameter (default: 7 days)
	daysStr := c.Query("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	companyIDStr := c.Query("companyId")

	// Month name to number mapping
	monthToNum := map[string]int{
		"Janeiro": 1, "Fevereiro": 2, "Marco": 3, "Abril": 4,
		"Maio": 5, "Junho": 6, "Julho": 7, "Agosto": 8,
		"Setembro": 9, "Outubro": 10, "Novembro": 11, "Dezembro": 12,
	}

	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	var rows interface {
		Close()
		Next() bool
		Scan(dest ...interface{}) error
	}

	if companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return helpers.InvalidIDFormat(c, "companyId")
		}

		// Validate ownership
		_, err = helpers.ValidateCompanyOwnership(c, companyID)
		if err != nil {
			return err
		}

		// Fetch open transactions for this company
		pgRows, err := database.Query(ctx,
			`SELECT id, company_id, category, description, amount, month, year, due_day, status, created_at, updated_at
			 FROM transactions
			 WHERE company_id = $1 AND status = 'ABERTO'
			 ORDER BY year, month, due_day`,
			companyID)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}
		rows = pgRows
	} else {
		// Fetch open transactions for all user's companies
		pgRows, err := database.Query(ctx,
			`SELECT t.id, t.company_id, t.category, t.description, t.amount, t.month, t.year, t.due_day, t.status, t.created_at, t.updated_at
			 FROM transactions t
			 INNER JOIN companies c ON t.company_id = c.id
			 WHERE c.user_id = $1 AND t.status = 'ABERTO'
			 ORDER BY t.year, t.month, t.due_day`,
			userID)
		if err != nil {
			return helpers.TransactionFetchFailed(c, err)
		}
		rows = pgRows
	}

	defer rows.Close()

	var upcomingTransactions []models.UpcomingTransaction

	for rows.Next() {
		var t models.Transaction
		var description *string
		var dueDay *int
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.Category, &description, &t.Amount, &t.Month, &t.Year, &dueDay, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return helpers.DatabaseError(c, "scan_transaction", err)
		}
		if description != nil {
			t.Description = *description
		}
		if dueDay != nil {
			t.DueDay = *dueDay
		} else {
			t.DueDay = 1
		}

		// Calculate due date
		monthNum, ok := monthToNum[t.Month]
		if !ok {
			continue
		}

		dueDate := time.Date(t.Year, time.Month(monthNum), t.DueDay, 0, 0, 0, 0, time.Local)

		// Check if due date is within the range (today to futureDate)
		if dueDate.After(now.AddDate(0, 0, -1)) && !dueDate.After(futureDate) {
			daysUntilDue := int(dueDate.Sub(now).Hours() / 24)
			if daysUntilDue < 0 {
				daysUntilDue = 0 // Already overdue
			}

			upcomingTransactions = append(upcomingTransactions, models.UpcomingTransaction{
				Transaction:  t,
				DaysUntilDue: daysUntilDue,
			})
		}
	}

	if upcomingTransactions == nil {
		upcomingTransactions = []models.UpcomingTransaction{}
	}

	return c.JSON(fiber.Map{
		"upcoming": upcomingTransactions,
		"count":    len(upcomingTransactions),
		"days":     days,
	})
}
