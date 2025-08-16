package authserv

import (
	"context"
	"fmt"
	"go-game-backend/pkg/futils"
	"go-game-backend/services/auth/internal/dto"
	"go-game-backend/services/auth/internal/services/token"
	"go-game-backend/services/auth/pkg/models"
	"time"

	"github.com/google/uuid"
)

// Config holds configuration for the authentication service.
type Config struct {
	PlayerLockTTL time.Duration `yaml:"player-lock-ttl"`
}

// PlayerStorageGateway defines the required operations for interacting with
// the player storage service.
type PlayerStorageGateway interface {
	AddUser(ctx context.Context, loginToken uuid.UUID) (userID int64, err error)
	FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (userID int64, err error)
}

// SessionRepository describes the persistence layer for session information.
type SessionRepository interface {
	DoWithTransaction(ctx context.Context, f futils.CtxF) error
	DoWithPlayerLock(ctx context.Context, userID int64, ttl time.Duration, f futils.CtxF) error
	SetSessionToken(ctx context.Context, userID int64, token uuid.UUID, expiresAt time.Time) error
	SetRefreshToken(ctx context.Context, token uuid.UUID, sessionInfo dto.SessionInfo, expiresAt time.Time) error
	RemoveRefreshToken(ctx context.Context, token uuid.UUID) error
	GetSessionInfo(ctx context.Context, refreshToken uuid.UUID) (dto.SessionInfo, error)
}

// Service provides authentication related operations such as registration,
// login and token refresh.
type Service struct {
	cfg               *Config
	userRepository    PlayerStorageGateway
	sessionRepository SessionRepository
	tokensFactory     *tknfactory.TokensFactory
}

// New creates a new Service instance with the supplied dependencies.
func New(
	cfg *Config,
	userRepository PlayerStorageGateway,
	sessionRepository SessionRepository,
	tokensFactory *tknfactory.TokensFactory,
) *Service {
	return &Service{
		cfg:               cfg,
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		tokensFactory:     tokensFactory,
	}
}

// Register creates a new user using the provided login token and returns a
// session with access and refresh tokens.
func (l *Service) Register(ctx context.Context, req *models.RegisterRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.userRepository.AddUser(ctx, req.LoginToken)
	if err != nil {
		return nil, fmt.Errorf("add user failed: %w", err)
	}

	sessionInfo, err := l.startSessionWithPlayerLock(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return sessionInfo, nil
}

// Login authenticates a user and starts a new session.
func (l *Service) Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.userRepository.FindUserByLoginToken(ctx, req.LoginToken)
	if err != nil {
		return nil, fmt.Errorf("find user failed: %w", err)
	}

	sessionInfo, err := l.startSessionWithPlayerLock(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return sessionInfo, nil
}

func (l *Service) startSessionWithPlayerLock(ctx context.Context, userID int64) (resp *models.LoginRespose, err error) {
	err = l.sessionRepository.DoWithPlayerLock(ctx, userID, l.cfg.PlayerLockTTL, func(ctx context.Context) error {
		resp, err = l.startSession(ctx, userID)
		if err != nil {
			return fmt.Errorf("start session failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *Service) startSession(ctx context.Context, userID int64) (resp *models.LoginRespose, err error) {
	utcNow := time.Now().UTC()

	sessionToken, sessionTokenExpiresAt := l.tokensFactory.CreateSessionToken(utcNow)
	accessToken, accessTokenExpiresAt, err := l.tokensFactory.CreateAccessToken(userID, sessionToken, utcNow)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken, refreshTokenExpiresAt := l.tokensFactory.CreateRefreshToken(utcNow)

	err = l.sessionRepository.DoWithTransaction(ctx, func(ctx context.Context) error {
		err := l.sessionRepository.SetSessionToken(ctx, userID, sessionToken, sessionTokenExpiresAt)
		if err != nil {
			return fmt.Errorf("session token saving failed: %w", err)
		}

		sessionInfo := dto.SessionInfo{
			UserID:       userID,
			SessionToken: sessionToken,
		}
		err = l.sessionRepository.SetRefreshToken(ctx, refreshToken, sessionInfo, refreshTokenExpiresAt)
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
		ExpiresAtUnix: accessTokenExpiresAt.Unix(),
	}
	return resp, nil
}

// RefreshToken exchanges a refresh token for a new pair of access and refresh
// tokens.
func (l *Service) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (resp *models.LoginRespose, err error) {
	utcNow := time.Now().UTC()

	sessionInfo, err := l.sessionRepository.GetSessionInfo(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("getting session info failed: %w", err)
	}

	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(sessionInfo.UserID, sessionInfo.SessionToken, utcNow)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken, refreshTokenExpiresAt := l.tokensFactory.CreateRefreshToken(utcNow)

	err = l.sessionRepository.DoWithTransaction(ctx, func(ctx context.Context) error {
		err := l.sessionRepository.RemoveRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return fmt.Errorf("removing refresh token failed: %w", err)
		}

		err = l.sessionRepository.SetRefreshToken(ctx, refreshToken, sessionInfo, refreshTokenExpiresAt)
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
