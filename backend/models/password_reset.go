package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a password reset token stored in the database
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"createdAt"`
}

// ForgotPasswordRequest is the request body for requesting a password reset
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest is the request body for resetting the password
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// ForgotPasswordResponse is the response for forgot password request
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}
