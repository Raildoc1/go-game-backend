package logic

import (
	"context"
	"fmt"
	"go-game-backend/services/auth/internal/dto"
	"go-game-backend/services/auth/pkg/models"
)

type UserRepository interface {
	AddUser(ctx context.Context, loginToken string) (userID int64, err error)
	FindUserByLoginToken(ctx context.Context, loginToken string) (userID int64, err error)
	FindUserByRefreshToken(ctx context.Context, loginToken string) (userID int64, err error)
}

type SessionRepository interface {
	DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error
	SetSessionToken(ctx context.Context, userID int64, token string) error
	SetRefreshToken(ctx context.Context, token string, sessionInfo dto.SessionInfo) error
	RemoveRefreshToken(ctx context.Context, token string) error
	GetSessionInfo(ctx context.Context, refreshToken string) (sessionInfo dto.SessionInfo, err error)
}

type Logic struct {
	userRepository    UserRepository
	sessionRepository SessionRepository
	tokensFactory     *TokensFactory
}

func NewLogic(
	userRepository UserRepository,
	sessionRepository SessionRepository,
	tokensFactory *TokensFactory,
) *Logic {
	return &Logic{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		tokensFactory:     tokensFactory,
	}
}

func (l *Logic) Register(ctx context.Context, req *models.RegisterRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.userRepository.AddUser(ctx, req.LoginToken)
	if err != nil {
		return nil, fmt.Errorf("add user failed: %w", err)
	}

	sessionInfo, err := l.startSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return sessionInfo, nil
}

func (l *Logic) Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.userRepository.FindUserByLoginToken(ctx, req.LoginToken)
	if err != nil {
		return nil, fmt.Errorf("find user failed: %w", err)
	}

	sessionInfo, err := l.startSession(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return sessionInfo, nil
}

func (l *Logic) startSession(ctx context.Context, userID int64) (resp *models.LoginRespose, err error) {
	sessionToken := l.tokensFactory.CreateSessionToken()
	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(userID, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken := l.tokensFactory.CreateRefreshToken()

	err = l.sessionRepository.DoWithTransaction(ctx, func(ctx context.Context) error {
		err := l.sessionRepository.SetSessionToken(ctx, userID, sessionToken)
		if err != nil {
			return fmt.Errorf("session token saving failed: %w", err)
		}

		sessionInfo := dto.SessionInfo{
			UserID:       userID,
			SessionToken: sessionToken,
		}
		err = l.sessionRepository.SetRefreshToken(ctx, refreshToken, sessionInfo)
		if err != nil {
			return fmt.Errorf("refresh token saving failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp = &models.LoginRespose{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAtUnix: expiresAt.Unix(),
	}
	return resp, nil
}

func (l *Logic) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (resp *models.LoginRespose, err error) {
	sessionInfo, err := l.sessionRepository.GetSessionInfo(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("getting session info failed: %w", err)
	}

	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(sessionInfo.UserID, sessionInfo.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken := l.tokensFactory.CreateRefreshToken()

	err = l.sessionRepository.DoWithTransaction(ctx, func(ctx context.Context) error {
		err := l.sessionRepository.RemoveRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return fmt.Errorf("removing refresh token failed: %w", err)
		}

		err = l.sessionRepository.SetRefreshToken(ctx, refreshToken, sessionInfo)
		if err != nil {
			return fmt.Errorf("refresh token saving failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp = &models.LoginRespose{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAtUnix: expiresAt.Unix(),
	}
	return resp, nil
}
