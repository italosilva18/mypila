package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token stored in the database
// The actual token is hashed before storage for security
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	TokenHash string     `json:"-"`                      // SHA256 hash of the token
	UserID    uuid.UUID  `json:"userId"`                 // Reference to the user
	ExpiresAt time.Time  `json:"expiresAt"`              // When the token expires
	CreatedAt time.Time  `json:"createdAt"`              // When the token was created
	IsRevoked bool       `json:"isRevoked"`              // Whether the token has been revoked
	RevokedAt *time.Time `json:"revokedAt,omitempty"`    // When it was revoked
}

// RefreshTokenRequest is the request body for refreshing tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// TokenPair represents both access and refresh tokens returned to the client
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // Access token expiry in seconds
}

// AuthResponseWithTokens is the response for login/register with token pair
type AuthResponseWithTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // Access token expiry in seconds
	User         User   `json:"user"`
}

// LogoutRequest is the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// LogoutAllRequest can be used to logout from all devices
type LogoutAllRequest struct {
	// Empty - just revokes all tokens for the authenticated user
}
