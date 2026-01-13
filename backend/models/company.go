package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	CNPJ      string    `json:"cnpj,omitempty"`
	LegalName string    `json:"legalName,omitempty"`
	TradeName string    `json:"tradeName,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	City      string    `json:"city,omitempty"`
	State     string    `json:"state,omitempty"`
	ZipCode   string    `json:"zipCode,omitempty"`
	LogoURL   string    `json:"logoUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateCompanyRequest struct {
	Name      string `json:"name"`
	CNPJ      string `json:"cnpj,omitempty"`
	LegalName string `json:"legalName,omitempty"`
	TradeName string `json:"tradeName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	ZipCode   string `json:"zipCode,omitempty"`
	LogoURL   string `json:"logoUrl,omitempty"`
}

type UpdateCompanyRequest struct {
	Name      string `json:"name,omitempty"`
	CNPJ      string `json:"cnpj,omitempty"`
	LegalName string `json:"legalName,omitempty"`
	TradeName string `json:"tradeName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	ZipCode   string `json:"zipCode,omitempty"`
	LogoURL   string `json:"logoUrl,omitempty"`
}
