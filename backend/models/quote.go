package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuoteStatus string

const (
	QuoteDraft    QuoteStatus = "DRAFT"    // Rascunho
	QuoteSent     QuoteStatus = "SENT"     // Enviado ao cliente
	QuoteApproved QuoteStatus = "APPROVED" // Aprovado
	QuoteRejected QuoteStatus = "REJECTED" // Rejeitado
	QuoteExecuted QuoteStatus = "EXECUTED" // Executado
)

type DiscountType string

const (
	DiscountPercent DiscountType = "PERCENT"
	DiscountValue   DiscountType = "VALUE"
)

// QuoteItem representa um item do orçamento
type QuoteItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Description string             `json:"description" bson:"description"`
	Quantity    float64            `json:"quantity" bson:"quantity"`
	UnitPrice   float64            `json:"unitPrice" bson:"unitPrice"`
	Total       float64            `json:"total" bson:"total"`
	CategoryID  primitive.ObjectID `json:"categoryId,omitempty" bson:"categoryId,omitempty"`
}

// Quote representa um orçamento comercial
type Quote struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CompanyID      primitive.ObjectID `json:"companyId" bson:"companyId"`
	Number         string             `json:"number" bson:"number"` // Auto: "ORC-2024-001"
	// Dados do Cliente (cadastro completo)
	ClientName     string             `json:"clientName" bson:"clientName"`
	ClientEmail    string             `json:"clientEmail,omitempty" bson:"clientEmail,omitempty"`
	ClientPhone    string             `json:"clientPhone,omitempty" bson:"clientPhone,omitempty"`
	ClientDocument string             `json:"clientDocument,omitempty" bson:"clientDocument,omitempty"` // CPF ou CNPJ
	ClientAddress  string             `json:"clientAddress,omitempty" bson:"clientAddress,omitempty"`
	ClientCity     string             `json:"clientCity,omitempty" bson:"clientCity,omitempty"`
	ClientState    string             `json:"clientState,omitempty" bson:"clientState,omitempty"`
	ClientZipCode  string             `json:"clientZipCode,omitempty" bson:"clientZipCode,omitempty"`
	// Dados do Orçamento
	Title        string             `json:"title" bson:"title"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty"`
	Items        []QuoteItem        `json:"items" bson:"items"`
	Subtotal     float64            `json:"subtotal" bson:"subtotal"`
	Discount     float64            `json:"discount" bson:"discount"`
	DiscountType DiscountType       `json:"discountType" bson:"discountType"`
	Total        float64            `json:"total" bson:"total"`
	Status       QuoteStatus        `json:"status" bson:"status"`
	ValidUntil   time.Time          `json:"validUntil" bson:"validUntil"`
	Notes        string             `json:"notes,omitempty" bson:"notes,omitempty"`
	TemplateID   primitive.ObjectID `json:"templateId,omitempty" bson:"templateId,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// CreateQuoteItemRequest para criar item do orçamento
type CreateQuoteItemRequest struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	CategoryID  string  `json:"categoryId,omitempty"`
}

// CreateQuoteRequest para criar orçamento
type CreateQuoteRequest struct {
	// Dados do Cliente
	ClientName     string `json:"clientName"`
	ClientEmail    string `json:"clientEmail,omitempty"`
	ClientPhone    string `json:"clientPhone,omitempty"`
	ClientDocument string `json:"clientDocument,omitempty"`
	ClientAddress  string `json:"clientAddress,omitempty"`
	ClientCity     string `json:"clientCity,omitempty"`
	ClientState    string `json:"clientState,omitempty"`
	ClientZipCode  string `json:"clientZipCode,omitempty"`
	// Dados do Orçamento
	Title        string                   `json:"title"`
	Description  string                   `json:"description,omitempty"`
	Items        []CreateQuoteItemRequest `json:"items"`
	Discount     float64                  `json:"discount"`
	DiscountType string                   `json:"discountType"`
	ValidUntil   string                   `json:"validUntil"` // ISO date string
	Notes        string                   `json:"notes,omitempty"`
	TemplateID   string                   `json:"templateId,omitempty"`
}

// UpdateQuoteRequest para atualizar orçamento
type UpdateQuoteRequest struct {
	// Dados do Cliente
	ClientName     string `json:"clientName"`
	ClientEmail    string `json:"clientEmail,omitempty"`
	ClientPhone    string `json:"clientPhone,omitempty"`
	ClientDocument string `json:"clientDocument,omitempty"`
	ClientAddress  string `json:"clientAddress,omitempty"`
	ClientCity     string `json:"clientCity,omitempty"`
	ClientState    string `json:"clientState,omitempty"`
	ClientZipCode  string `json:"clientZipCode,omitempty"`
	// Dados do Orçamento
	Title        string                   `json:"title"`
	Description  string                   `json:"description,omitempty"`
	Items        []CreateQuoteItemRequest `json:"items"`
	Discount     float64                  `json:"discount"`
	DiscountType string                   `json:"discountType"`
	ValidUntil   string                   `json:"validUntil"`
	Notes        string                   `json:"notes,omitempty"`
	TemplateID   string                   `json:"templateId,omitempty"`
}

// UpdateQuoteStatusRequest para atualizar status
type UpdateQuoteStatusRequest struct {
	Status string `json:"status"`
}

// QuoteComparisonItem para comparativo de item
type QuoteComparisonItem struct {
	Description string  `json:"description"`
	CategoryID  string  `json:"categoryId,omitempty"`
	Quoted      float64 `json:"quoted"`
	Executed    float64 `json:"executed"`
	Variance    float64 `json:"variance"`
}

// QuoteComparison para comparativo orçado vs realizado
type QuoteComparison struct {
	QuoteID         string                `json:"quoteId"`
	QuotedTotal     float64               `json:"quotedTotal"`
	ExecutedTotal   float64               `json:"executedTotal"`
	Variance        float64               `json:"variance"`
	VariancePercent float64               `json:"variancePercent"`
	Items           []QuoteComparisonItem `json:"items"`
}
