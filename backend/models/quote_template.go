package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuoteTemplate representa um template de orçamento
type QuoteTemplate struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CompanyID    primitive.ObjectID `json:"companyId" bson:"companyId"`
	Name         string             `json:"name" bson:"name"`
	HeaderText   string             `json:"headerText,omitempty" bson:"headerText,omitempty"`
	FooterText   string             `json:"footerText,omitempty" bson:"footerText,omitempty"`
	TermsText    string             `json:"termsText,omitempty" bson:"termsText,omitempty"`
	PrimaryColor string             `json:"primaryColor" bson:"primaryColor"` // Hex color
	LogoURL      string             `json:"logoUrl,omitempty" bson:"logoUrl,omitempty"`
	IsDefault    bool               `json:"isDefault" bson:"isDefault"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
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
