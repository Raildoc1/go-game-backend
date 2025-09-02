// Package dto contains internal data transfer objects used by the auth service.
package dto

import "github.com/google/uuid"

// SessionInfo contains session token information associated with a user.
type SessionInfo struct {
	UserID       int64
	SessionToken uuid.UUID
}
