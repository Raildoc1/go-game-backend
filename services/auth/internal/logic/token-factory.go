package logic

import (
	"fmt"
	"github.com/google/uuid"
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

func (f *TokensFactory) CreateAccessToken(userID int64, sessionToken string) (tkn string, expiresAt time.Time, err error) {
	payload := map[string]any{
		"userID":  userID,
		"session": sessionToken,
	}
	tkn, expiresAt, err = f.tokenFactory.Generate(f.config.AccessTokenTTL, payload)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("jwt token generation failed: %w", err)
	}
	return tkn, expiresAt, nil
}

func (f *TokensFactory) CreateRefreshToken() string {
	return uuid.New().String()
}

func (f *TokensFactory) CreateSessionToken() string {
	return uuid.New().String()
}
