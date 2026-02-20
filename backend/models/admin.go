package models

import (
	"time"

	"github.com/google/uuid"
)

// AdminStats representa as estatísticas do dashboard administrativo
type AdminStats struct {
	TotalUsers        int64       `json:"totalUsers"`
	TotalCompanies    int64       `json:"totalCompanies"`
	TotalTransactions int64       `json:"totalTransactions"`
	TotalRevenue      float64     `json:"totalRevenue"`
	RecentUsers       []User      `json:"recentUsers"`
	RecentTransactions []AdminTransaction `json:"recentTransactions"`
}

// AdminTransaction representa uma transação na visão administrativa
type AdminTransaction struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"companyId"`
	CompanyName string    `json:"companyName"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Date        string    `json:"date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AdminUserResponse representa um usuário na visão administrativa
type AdminUserResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	CompanyCount  int       `json:"companyCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AdminCompanyResponse representa uma empresa na visão administrativa
type AdminCompanyResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"userId"`
	UserName        string    `json:"userName"`
	UserEmail       string    `json:"userEmail"`
	Name            string    `json:"name"`
	Cnpj            *string    `json:"cnpj,omitempty"`
	TransactionCount int      `json:"transactionCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UpdateUserRequest representa a requisição para atualizar um usuário
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Pagination representa a paginação de resultados
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PaginatedResponse representa uma resposta paginada
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// UserFilter representa os filtros para listagem de usuários
type UserFilter struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// CompanyFilter representa os filtros para listagem de empresas
type CompanyFilter struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// TransactionFilter representa os filtros para listagem de transações
type TransactionFilter struct {
	Search    string `query:"search"`
	Status    string `query:"status"`
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
}
