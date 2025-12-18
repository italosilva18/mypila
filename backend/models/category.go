package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryType string

const (
	Expense CategoryType = "EXPENSE"
	Income  CategoryType = "INCOME"
)

type Category struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CompanyID primitive.ObjectID `json:"companyId" bson:"companyId"`
	Name      string             `json:"name" bson:"name"`
	Type      CategoryType       `json:"type" bson:"type"`
	Color     string             `json:"color" bson:"color"`
	Budget    float64            `json:"budget" bson:"budget"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
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
