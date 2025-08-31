// Package authserv contains the core authentication service logic.
package authserv

import (
	"context"
	"encoding/json"
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
	PlayerLockTTL    time.Duration `yaml:"player-lock-ttl"`
	UserCreatedTopic string        `yaml:"user-created-topic"`
}

type PgStore interface {
	DoTx(ctx context.Context, f futils.CtxFT[PgRepos]) error
	Raw() PgRepos
}

type PgRepos interface {
	User() UserRepo
	Outbox() OutboxRepo
}

type UserRepo interface {
	AddUser(ctx context.Context, loginToken uuid.UUID) (userID int64, err error)
	FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (userID int64, err error)
}

type OutboxRepo interface {
	Add(ctx context.Context, topic string, payload []byte) error
}

type RxStore interface {
	DoTx(ctx context.Context, f futils.CtxFT[RxRepos]) error
	Raw() RxRepos
}

type PlayerLocker interface {
	DoWithPlayerLock(ctx context.Context, userID int64, f futils.CtxF) error
}

type RxRepos interface {
	Session() SessionRepo
}

type SessionRepo interface {
	SetSessionToken(ctx context.Context, userID int64, token uuid.UUID, expiresAt time.Time) error
	SetRefreshToken(ctx context.Context, token uuid.UUID, sessionInfo dto.SessionInfo, expiresAt time.Time) error
	RemoveRefreshToken(ctx context.Context, token uuid.UUID) error
	GetSessionInfo(ctx context.Context, refreshToken uuid.UUID) (dto.SessionInfo, error)
}

// Service provides authentication related operations such as registration,
// login and token refresh.
type Service struct {
	cfg           *Config
	pgStore       PgStore
	rxStore       RxStore
	playerLocker  PlayerLocker
	tokensFactory *tknfactory.TokensFactory
}

// New creates a new Service instance with the supplied dependencies.
func New(
	cfg *Config,
	pgStore PgStore,
	rxStore RxStore,
	playerLocker PlayerLocker,
	tokensFactory *tknfactory.TokensFactory,
) *Service {
	return &Service{
		cfg:           cfg,
		pgStore:       pgStore,
		rxStore:       rxStore,
		playerLocker:  playerLocker,
		tokensFactory: tokensFactory,
	}
}

// Register creates a new user using the provided login token and returns a
// session with access and refresh tokens.
func (l *Service) Register(ctx context.Context, req *models.RegisterRequest) (resp *models.LoginRespose, err error) {
	var userID int64

	err = l.pgStore.DoTx(ctx, func(ctx context.Context, r PgRepos) error {
		userID, err = r.User().AddUser(ctx, req.LoginToken)
		if err != nil {
			return fmt.Errorf("add user failed: %w", err)
		}
		payload, err := json.Marshal(models.UserCreatedEvent{UserID: userID})
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}
		if err := r.Outbox().Add(ctx, l.cfg.UserCreatedTopic, payload); err != nil {
			return fmt.Errorf("save outbox event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sessionInfo, err := l.startSessionWithPlayerLock(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return sessionInfo, nil
}

// Login authenticates a user and starts a new session.
func (l *Service) Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.pgStore.Raw().User().FindUserByLoginToken(ctx, req.LoginToken)
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
	err = l.playerLocker.DoWithPlayerLock(ctx, userID, func(ctx context.Context) error {
		resp, err = l.startSession(ctx, userID)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
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

	err = l.rxStore.DoTx(ctx, func(ctx context.Context, r RxRepos) error {
		err := r.Session().SetSessionToken(ctx, userID, sessionToken, sessionTokenExpiresAt)
		if err != nil {
			return fmt.Errorf("set session token: %w", err)
		}

		sessionInfo := dto.SessionInfo{
			UserID:       userID,
			SessionToken: sessionToken,
		}
		err = r.Session().SetRefreshToken(ctx, refreshToken, sessionInfo, refreshTokenExpiresAt)
		if err != nil {
			return fmt.Errorf("set refresh token: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("rx transaction: %w", err)
	}

	resp = &models.LoginRespose{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAtUnix: accessTokenExpiresAt.Unix(),
	}
	return resp, nil
}

// RefreshToken exchanges a refresh token for a new pair of access and refresh tokens.
func (l *Service) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (resp *models.LoginRespose, err error) {
	utcNow := time.Now().UTC()

	sessionInfo, err := l.rxStore.Raw().Session().GetSessionInfo(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("getting session info failed: %w", err)
	}

	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(sessionInfo.UserID, sessionInfo.SessionToken, utcNow)
	if err != nil {
		return nil, fmt.Errorf("token creation failed: %w", err)
	}
	refreshToken, refreshTokenExpiresAt := l.tokensFactory.CreateRefreshToken(utcNow)

	err = l.rxStore.DoTx(ctx, func(ctx context.Context, r RxRepos) error {
		err := r.Session().RemoveRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return fmt.Errorf("removing refresh token failed: %w", err)
		}

		err = r.Session().SetRefreshToken(ctx, refreshToken, sessionInfo, refreshTokenExpiresAt)
		if err != nil {
			return fmt.Errorf("refresh token saving failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token transaction failed: %w", err)
	}

	resp = &models.LoginRespose{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAtUnix: expiresAt.Unix(),
	}
	return resp, nil
}
