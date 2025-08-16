package models

import "github.com/google/uuid"

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken uuid.UUID `json:"refresh_token"`
}
