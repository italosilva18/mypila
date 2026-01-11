package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	CategorySalary     = "Salário"
	CategoryVacation   = "Férias"
	CategoryAICost     = "Custos de IA"
	CategoryDockerCost = "Custo de Docker"
)

type Status string

const (
	StatusPaid Status = "PAGO"
	StatusOpen Status = "ABERTO"
)

type Transaction struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"companyId"`
	Month       string    `json:"month"`
	Year        int       `json:"year"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Status      Status    `json:"status"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateTransactionRequest struct {
	CompanyID   string  `json:"companyId"`
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Status      Status  `json:"status"`
	Description string  `json:"description,omitempty"`
}

type UpdateTransactionRequest struct {
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Status      Status  `json:"status"`
	Description string  `json:"description,omitempty"`
}

type Stats struct {
	Paid  float64 `json:"paid"`
	Open  float64 `json:"open"`
	Total float64 `json:"total"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// PaginatedTransactions contains paginated transaction results
type PaginatedTransactions struct {
	Data       []Transaction      `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}
