package tknfactory

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
	cfg          *Config
	tokenFactory *jwtfactory.Factory
}

func New(tokenFactory *jwtfactory.Factory, config *Config) *TokensFactory {
	return &TokensFactory{
		cfg:          config,
		tokenFactory: tokenFactory,
	}
}

func (f *TokensFactory) CreateAccessToken(
	userID int64,
	sessionToken uuid.UUID,
	issueTime time.Time,
) (tkn string, expiresAt time.Time, err error) {
	payload := map[string]any{
		"userID":  userID,
		"session": sessionToken,
	}
	tkn, expiresAt, err = f.tokenFactory.Generate(f.cfg.AccessTokenTTL, issueTime, payload)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("jwt token generation failed: %w", err)
	}
	return tkn, expiresAt, nil
}

func (f *TokensFactory) CreateRefreshToken(issueTime time.Time) (tkn uuid.UUID, expiresAt time.Time) {
	return uuid.New(), issueTime.Add(f.cfg.RefreshTokenTTL)
}

func (f *TokensFactory) CreateSessionToken(issueTime time.Time) (tkn uuid.UUID, expiresAt time.Time) {
	return uuid.New(), issueTime.Add(f.cfg.RefreshTokenTTL)
}
