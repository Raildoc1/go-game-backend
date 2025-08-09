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

func (tf *Factory) Generate(ttl time.Duration, extraClaims map[string]string) (string, error) {
	timeNow := time.Now()
	claims := map[string]any{
		"exp": timeNow.Add(ttl).Unix(),
		"iat": timeNow.Unix(),
	}
	for k, v := range extraClaims {
		claims[k] = v
	}
	_, tokenString, err := tf.auth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, nil
}
