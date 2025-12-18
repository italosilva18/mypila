package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecurringTransaction struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CompanyID   primitive.ObjectID `json:"companyId" bson:"companyId"`
	Description string             `json:"description" bson:"description"`
	Amount      float64            `json:"amount" bson:"amount"`
	Category    string             `json:"category" bson:"category"`
	DayOfMonth  int                `json:"dayOfMonth" bson:"dayOfMonth"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

type CreateRecurringRequest struct {
	CompanyID   string  `json:"companyId"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	DayOfMonth  int     `json:"dayOfMonth"`
}
