package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"m2m-backend/database"
	"m2m-backend/models"
)

// GetAdminStats retorna as estatísticas para o dashboard administrativo
func GetAdminStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Contar total de usuários
	var userCount int64
	err := database.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao contar usuários",
			"message": err.Error(),
		})
	}

	// Contar total de empresas
	var companyCount int64
	err = database.QueryRow(ctx, `SELECT COUNT(*) FROM companies`).Scan(&companyCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao contar empresas",
			"message": err.Error(),
		})
	}

	// Contar total de transações
	var transactionCount int64
	err = database.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&transactionCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao contar transações",
			"message": err.Error(),
		})
	}

	// Calcular receita total (soma de todas as transações pagas)
	var totalRevenue float64
	err = database.QueryRow(ctx, `
		SELECT COALESCE(SUM(paid_amount), 0) 
		FROM transactions 
		WHERE status = 'PAGO'
	`).Scan(&totalRevenue)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao calcular receita",
			"message": err.Error(),
		})
	}

	// Buscar usuários recentes (últimos 5)
	rows, err := database.Query(ctx, `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT 5
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao buscar usuários recentes",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var recentUsers []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		recentUsers = append(recentUsers, user)
	}

	// Buscar transações recentes (últimas 5)
	txRows, err := database.Query(ctx, `
		SELECT t.id, t.company_id, c.name as company_name, t.description, t.amount, t.category, 
		       t.year || '-' || t.month || '-' || t.due_day as date, t.status, t.created_at
		FROM transactions t
		LEFT JOIN companies c ON t.company_id = c.id
		ORDER BY t.created_at DESC 
		LIMIT 5
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao buscar transações recentes",
			"message": err.Error(),
		})
	}
	defer txRows.Close()

	var recentTransactions []models.AdminTransaction
	for txRows.Next() {
		var tx models.AdminTransaction
		err := txRows.Scan(&tx.ID, &tx.CompanyID, &tx.CompanyName, &tx.Description, &tx.Amount, &tx.Category, &tx.Date, &tx.Status, &tx.CreatedAt)
		if err != nil {
			continue
		}
		recentTransactions = append(recentTransactions, tx)
	}

	stats := models.AdminStats{
		TotalUsers:         userCount,
		TotalCompanies:     companyCount,
		TotalTransactions:  transactionCount,
		TotalRevenue:       totalRevenue,
		RecentUsers:        recentUsers,
		RecentTransactions: recentTransactions,
	}

	return c.JSON(stats)
}

// GetAllUsers retorna todos os usuários com paginação
func GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 50)
	if limit < 1 || limit > 100 {
		limit = 50
	}
	search := c.Query("search")

	offset := (page - 1) * limit

	var total int64
	var rows interface{ Close() }

	if search != "" {
		// Count with search
		searchPattern := "%" + search + "%"
		err := database.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR email ILIKE $1`,
			searchPattern).Scan(&total)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao contar usuários",
				"message": err.Error(),
			})
		}

		// Fetch with search
		pgRows, err := database.Query(ctx,
			`SELECT id, name, email, created_at, updated_at 
			 FROM users 
			 WHERE name ILIKE $1 OR email ILIKE $1
			 ORDER BY created_at DESC 
			 LIMIT $2 OFFSET $3`,
			searchPattern, limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao buscar usuários",
				"message": err.Error(),
			})
		}
		rows = pgRows
	} else {
		// Count all
		err := database.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao contar usuários",
				"message": err.Error(),
			})
		}

		// Fetch all
		pgRows, err := database.Query(ctx,
			`SELECT id, name, email, created_at, updated_at 
			 FROM users 
			 ORDER BY created_at DESC 
			 LIMIT $1 OFFSET $2`,
			limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao buscar usuários",
				"message": err.Error(),
			})
		}
		rows = pgRows
	}
	defer rows.Close()

	var users []models.AdminUserResponse
	pgRows := rows.(interface {
		Close()
		Next() bool
		Scan(dest ...interface{}) error
	})
	
	for pgRows.Next() {
		var user models.AdminUserResponse
		err := pgRows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		
		// Contar empresas do usuário
		var companyCount int64
		database.QueryRow(ctx, `SELECT COUNT(*) FROM companies WHERE user_id = $1`, user.ID).Scan(&companyCount)
		user.CompanyCount = int(companyCount)
		
		users = append(users, user)
	}

	response := models.PaginatedResponse{
		Data: users,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	return c.JSON(response)
}

// UpdateUser atualiza um usuário
func UpdateUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "ID inválido",
			"message": err.Error(),
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Dados inválidos",
			"message": err.Error(),
		})
	}

	result, err := database.Pool.Exec(ctx,
		`UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3`,
		req.Name, req.Email, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao atualizar usuário",
			"message": err.Error(),
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Usuário não encontrado",
			"message": "O usuário especificado não existe",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Usuário atualizado com sucesso",
	})
}

// DeleteUser remove um usuário
func DeleteUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "ID inválido",
			"message": err.Error(),
		})
	}

	// Deletar transações das empresas do usuário
	_, err = database.Pool.Exec(ctx,
		`DELETE FROM transactions WHERE company_id IN (SELECT id FROM companies WHERE user_id = $1)`,
		userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao deletar transações",
			"message": err.Error(),
		})
	}

	// Deletar empresas do usuário
	_, err = database.Pool.Exec(ctx,
		`DELETE FROM companies WHERE user_id = $1`,
		userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao deletar empresas",
			"message": err.Error(),
		})
	}

	// Deletar usuário
	result, err := database.Pool.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
		userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao deletar usuário",
			"message": err.Error(),
		})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Usuário não encontrado",
			"message": "O usuário especificado não existe",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Usuário deletado com sucesso",
	})
}

// GetAllCompanies retorna todas as empresas com paginação
func GetAllCompanies(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 50)
	if limit < 1 || limit > 100 {
		limit = 50
	}
	search := c.Query("search")

	offset := (page - 1) * limit

	var total int64
	var rows interface{ Close() }

	if search != "" {
		searchPattern := "%" + search + "%"
		err := database.QueryRow(ctx,
			`SELECT COUNT(*) FROM companies WHERE name ILIKE $1`,
			searchPattern).Scan(&total)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao contar empresas",
				"message": err.Error(),
			})
		}

		pgRows, err := database.Query(ctx,
			`SELECT c.id, c.user_id, c.name, c.cnpj, c.created_at, c.updated_at,
			        u.name as user_name, u.email as user_email
			 FROM companies c
			 LEFT JOIN users u ON c.user_id = u.id
			 WHERE c.name ILIKE $1
			 ORDER BY c.created_at DESC 
			 LIMIT $2 OFFSET $3`,
			searchPattern, limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao buscar empresas",
				"message": err.Error(),
			})
		}
		rows = pgRows
	} else {
		err := database.QueryRow(ctx, `SELECT COUNT(*) FROM companies`).Scan(&total)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao contar empresas",
				"message": err.Error(),
			})
		}

		pgRows, err := database.Query(ctx,
			`SELECT c.id, c.user_id, c.name, c.cnpj, c.created_at, c.updated_at,
			        u.name as user_name, u.email as user_email
			 FROM companies c
			 LEFT JOIN users u ON c.user_id = u.id
			 ORDER BY c.created_at DESC 
			 LIMIT $1 OFFSET $2`,
			limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Erro ao buscar empresas",
				"message": err.Error(),
			})
		}
		rows = pgRows
	}
	defer rows.Close()

	var companies []models.AdminCompanyResponse
	pgRows := rows.(interface {
		Close()
		Next() bool
		Scan(dest ...interface{}) error
	})
	
	for pgRows.Next() {
		var company models.AdminCompanyResponse
		var cnpj *string
		err := pgRows.Scan(&company.ID, &company.UserID, &company.Name, &cnpj, &company.CreatedAt, &company.UpdatedAt, &company.UserName, &company.UserEmail)
		if err != nil {
			continue
		}
		company.Cnpj = cnpj
		
		// Contar transações da empresa
		var txCount int64
		database.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE company_id = $1`, company.ID).Scan(&txCount)
		company.TransactionCount = int(txCount)
		
		companies = append(companies, company)
	}

	response := models.PaginatedResponse{
		Data: companies,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	return c.JSON(response)
}

// GetAllTransactionsAdmin retorna todas as transações com paginação (para admin)
func GetAllTransactionsAdmin(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 50)
	if limit < 1 || limit > 100 {
		limit = 50
	}
	search := c.Query("search")
	status := c.Query("status")

	offset := (page - 1) * limit

	// Construir query dinamicamente
	query := `SELECT t.id, t.company_id, c.name as company_name, t.description, t.amount, t.category, 
			  t.year || '-' || t.month || '-' || t.due_day as date, t.status, t.created_at
			  FROM transactions t
			  LEFT JOIN companies c ON t.company_id = c.id
			  WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM transactions t WHERE 1=1`
	
	var args []interface{}
	argCount := 0

	if search != "" {
		argCount++
		query += ` AND t.description ILIKE $` + string(rune('0'+argCount))
		countQuery += ` AND t.description ILIKE $` + string(rune('0'+argCount))
		args = append(args, "%"+search+"%")
	}
	
	if status != "" {
		argCount++
		query += ` AND t.status = $` + string(rune('0'+argCount))
		countQuery += ` AND t.status = $` + string(rune('0'+argCount))
		args = append(args, status)
	}

	// Count total
	var total int64
	err := database.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao contar transações",
			"message": err.Error(),
		})
	}

	// Add pagination
	argCount++
	query += ` ORDER BY t.created_at DESC LIMIT $` + string(rune('0'+argCount))
	args = append(args, limit)
	
	argCount++
	query += ` OFFSET $` + string(rune('0'+argCount))
	args = append(args, offset)

	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Erro ao buscar transações",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var transactions []models.AdminTransaction
	for rows.Next() {
		var tx models.AdminTransaction
		err := rows.Scan(&tx.ID, &tx.CompanyID, &tx.CompanyName, &tx.Description, &tx.Amount, &tx.Category, &tx.Date, &tx.Status, &tx.CreatedAt)
		if err != nil {
			continue
		}
		transactions = append(transactions, tx)
	}

	response := models.PaginatedResponse{
		Data: transactions,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	return c.JSON(response)
}
