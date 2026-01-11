package models

import (
	"time"

	"github.com/google/uuid"
)

type RecurringTransaction struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"companyId"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	DayOfMonth  int       `json:"dayOfMonth"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateRecurringRequest struct {
	CompanyID   string  `json:"companyId"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	DayOfMonth  int     `json:"dayOfMonth"`
}
