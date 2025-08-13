package models

import "github.com/google/uuid"

type LoginRequest struct {
	LoginToken uuid.UUID `json:"login_token"`
}
