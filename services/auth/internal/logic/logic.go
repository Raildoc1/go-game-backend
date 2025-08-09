package logic

import (
	"context"
	"fmt"
	"go-game-backend/services/auth/pkg/models"
)

type CredsRepository interface {
	ValidateCredentials(ctx context.Context, username, password string) (bool, error)
}

type SessionRepository interface {
	DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error
	SetSessionToken(ctx context.Context, username, token string) error
	SetRefreshToken(ctx context.Context, username, token string) error
}

type Logic struct {
	credsRepository   CredsRepository
	sessionRepository SessionRepository
	tokensFactory     *TokensFactory
}

func NewLogic(
	credsRepository CredsRepository,
	sessionRepository SessionRepository,
	tokensFactory *TokensFactory,
) *Logic {
	return &Logic{
		credsRepository:   credsRepository,
		sessionRepository: sessionRepository,
		tokensFactory:     tokensFactory,
	}
}

func (l *Logic) Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginResponse, err error) {
	valid, err := l.credsRepository.ValidateCredentials(ctx, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("credentials validation failed: %w", err)
	}
	if !valid {
		return nil, ErrValidationCredentials
	}

	sessionToken := l.tokensFactory.CreateSessionToken()
	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(req.Username, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken := l.tokensFactory.CreateRefreshToken()

	err = l.sessionRepository.DoWithTransaction(ctx, func(ctx context.Context) error {
		err := l.sessionRepository.SetSessionToken(ctx, req.Username, sessionToken)
		if err != nil {
			return fmt.Errorf("session token saving failed: %w", err)
		}
		err = l.sessionRepository.SetRefreshToken(ctx, req.Username, sessionToken)
		if err != nil {
			return fmt.Errorf("refresh token saving failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp = &models.LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAtUnix: expiresAt.Unix(),
	}
	return resp, nil
}
