package models

import (
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	Expense CategoryType = "EXPENSE"
	Income  CategoryType = "INCOME"
)

type Category struct {
	ID        uuid.UUID    `json:"id"`
	CompanyID uuid.UUID    `json:"companyId"`
	Name      string       `json:"name"`
	Type      CategoryType `json:"type"`
	Color     string       `json:"color"`
	Budget    float64      `json:"budget"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type CreateCategoryRequest struct {
	Name   string       `json:"name"`
	Type   CategoryType `json:"type"`
	Color  string       `json:"color"`
	Budget float64      `json:"budget"`
}

type UpdateCategoryRequest struct {
	Name   string       `json:"name"`
	Type   CategoryType `json:"type"`
	Color  string       `json:"color"`
	Budget float64      `json:"budget"`
}
