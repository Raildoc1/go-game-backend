// Package tknfactory provides utilities for creating access, refresh, and
// session tokens for the auth service.
package tknfactory

import (
	"fmt"
	"go-game-backend/pkg/jwtfactory"
	"time"

	"github.com/google/uuid"
)

// Config specifies the TTLs for the tokens produced by TokensFactory.
type Config struct {
	AccessTokenTTL  time.Duration `yaml:"access-token-ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh-token-ttl"`
}

// TokensFactory generates access, refresh, and session tokens.
type TokensFactory struct {
	cfg          *Config
	tokenFactory *jwtfactory.Factory
}

// New creates a TokensFactory using the provided JWT factory and configuration.
func New(tokenFactory *jwtfactory.Factory, config *Config) *TokensFactory {
	return &TokensFactory{
		cfg:          config,
		tokenFactory: tokenFactory,
	}
}

// CreateAccessToken generates a JWT access token for a user and session.
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

// CreateRefreshToken creates a refresh token that expires after the configured
// TTL.
func (f *TokensFactory) CreateRefreshToken(issueTime time.Time) (tkn uuid.UUID, expiresAt time.Time) {
	return uuid.New(), issueTime.Add(f.cfg.RefreshTokenTTL)
}

// CreateSessionToken creates a session token that expires after the refresh
// token TTL.
func (f *TokensFactory) CreateSessionToken(issueTime time.Time) (tkn uuid.UUID, expiresAt time.Time) {
	return uuid.New(), issueTime.Add(f.cfg.RefreshTokenTTL)
}
