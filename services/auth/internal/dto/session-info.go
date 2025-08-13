package dto

import "github.com/google/uuid"

type SessionInfo struct {
	UserID       int64
	SessionToken uuid.UUID
}
