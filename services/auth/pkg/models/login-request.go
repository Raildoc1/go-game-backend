package models

import "github.com/google/uuid"

// LoginRequest represents a request to log in using a login token.
type LoginRequest struct {
	LoginToken uuid.UUID `json:"login_token"`
}
