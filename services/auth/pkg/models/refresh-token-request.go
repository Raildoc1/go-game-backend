package models

import "github.com/google/uuid"

type RefreshTokenRequest struct {
	RefreshToken uuid.UUID `json:"refresh_token"`
}
