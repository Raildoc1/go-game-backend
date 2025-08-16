package services

import "errors"

// ErrValidationCredentials is returned when provided credentials are invalid.
var ErrValidationCredentials = errors.New("invalid credentials")
