package models

import (
	"time"

	"github.com/google/uuid"
)

// QuoteTemplate representa um template de orçamento
type QuoteTemplate struct {
	ID           uuid.UUID `json:"id"`
	CompanyID    uuid.UUID `json:"companyId"`
	Name         string    `json:"name"`
	HeaderText   string    `json:"headerText,omitempty"`
	FooterText   string    `json:"footerText,omitempty"`
	TermsText    string    `json:"termsText,omitempty"`
	PrimaryColor string    `json:"primaryColor"` // Hex color
	LogoURL      string    `json:"logoUrl,omitempty"`
	IsDefault    bool      `json:"isDefault"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CreateQuoteTemplateRequest para criar template
type CreateQuoteTemplateRequest struct {
	Name         string `json:"name"`
	HeaderText   string `json:"headerText,omitempty"`
	FooterText   string `json:"footerText,omitempty"`
	TermsText    string `json:"termsText,omitempty"`
	PrimaryColor string `json:"primaryColor"`
	LogoURL      string `json:"logoUrl,omitempty"`
	IsDefault    bool   `json:"isDefault"`
}

// UpdateQuoteTemplateRequest para atualizar template
type UpdateQuoteTemplateRequest struct {
	Name         string `json:"name"`
	HeaderText   string `json:"headerText,omitempty"`
	FooterText   string `json:"footerText,omitempty"`
	TermsText    string `json:"termsText,omitempty"`
	PrimaryColor string `json:"primaryColor"`
	LogoURL      string `json:"logoUrl,omitempty"`
	IsDefault    bool   `json:"isDefault"`
}
