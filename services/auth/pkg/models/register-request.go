package models

import "github.com/google/uuid"

type RegisterRequest struct {
	LoginToken uuid.UUID `json:"login_token"`
}
