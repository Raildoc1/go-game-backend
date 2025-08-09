package logic

import (
	"fmt"
	"go-game-backend/pkg/jwtfactory"
	"time"
)

type Config struct {
	AccessTokenTTL  time.Duration `yaml:"access-token-ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh-token-ttl"`
}

type TokensFactory struct {
	config       Config
	tokenFactory *jwtfactory.Factory
}

func NewTokensFactory(tokenFactory *jwtfactory.Factory, config Config) *TokensFactory {
	return &TokensFactory{
		config:       config,
		tokenFactory: tokenFactory,
	}
}

func (f *TokensFactory) CreateAccessToken(username, sessionToken string) (string, error) {
	payload := map[string]string{
		"username": username,
		"session":  sessionToken,
	}
	tkn, err := f.tokenFactory.Generate(f.config.AccessTokenTTL, payload)
	if err != nil {
		return "", fmt.Errorf("jwt token generation failed: %w", err)
	}
	return tkn, nil
}
