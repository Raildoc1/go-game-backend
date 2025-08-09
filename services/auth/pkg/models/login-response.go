package models

import "time"

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAtUTC time.Time `json:"expires_at"`
}
