// Package services defines shared components and errors for the auth service.
package services

import "errors"

// ErrValidationCredentials is returned when provided credentials are invalid.
var ErrValidationCredentials = errors.New("invalid credentials")
