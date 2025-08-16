package jwtfactory

import (
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

// Factory wraps jwtauth.JWTAuth and provides helper methods for creating
// JSON Web Tokens.
type Factory struct {
	auth *jwtauth.JWTAuth
}

// New creates a Factory using the provided JWTAuth instance.
func New(auth *jwtauth.JWTAuth) *Factory {
	return &Factory{
		auth: auth,
	}
}

// Generate creates a new signed JWT token with the specified TTL and issue
// time. Additional claims may be supplied via extraClaims. It returns the
// generated token string and its expiration time.
func (tf *Factory) Generate(
	ttl time.Duration,
	issueTime time.Time,
	extraClaims map[string]any,
) (tkn string, expiresAt time.Time, err error) {
	expiresAt = issueTime.Add(ttl)
	claims := map[string]any{
		"exp": expiresAt.Unix(),
		"iat": issueTime.Unix(),
	}
	for k, v := range extraClaims {
		claims[k] = v
	}
	_, tokenString, err := tf.auth.Encode(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, expiresAt, nil
}
