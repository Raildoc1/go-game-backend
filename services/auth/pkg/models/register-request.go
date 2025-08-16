package models

import "github.com/google/uuid"

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	LoginToken uuid.UUID `json:"login_token"`
}
