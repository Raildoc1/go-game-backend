package models

// UserCreatedEvent represents payload for user-created events.
type UserCreatedEvent struct {
	UserID int64 `json:"user_id"`
}
