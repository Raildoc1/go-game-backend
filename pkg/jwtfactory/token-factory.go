package jwtfactory

import (
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type Factory struct {
	auth *jwtauth.JWTAuth
}

func New(auth *jwtauth.JWTAuth) *Factory {
	return &Factory{
		auth: auth,
	}
}

func (tf *Factory) Generate(
	ttl time.Duration,
	extraClaims map[string]any,
) (tkn string, expiresAt time.Time, err error) {
	timeNow := time.Now()
	expiresAt = timeNow.Add(ttl)
	claims := map[string]any{
		"exp": expiresAt.Unix(),
		"iat": timeNow.Unix(),
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
