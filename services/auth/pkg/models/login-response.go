package models

import "github.com/google/uuid"

// LoginRespose contains tokens returned after a successful authentication.
type LoginRespose struct {
	AccessToken   string    `json:"access_token"`
	RefreshToken  uuid.UUID `json:"refresh_token"`
	ExpiresAtUnix int64     `json:"expires_at"`
}
