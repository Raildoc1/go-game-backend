// Package authsvc contains the core authentication service logic.
package authsvc

import (
	"context"
	"fmt"
	"go-game-backend/pkg/futils"
	"go-game-backend/services/auth/internal/dto"
	"go-game-backend/services/auth/internal/services/token"
	"go-game-backend/services/auth/pkg/models"
	"time"

	postgresstore "go-game-backend/pkg/postgres"
	redisstore "go-game-backend/pkg/redis"

	postgresrepo "go-game-backend/services/auth/internal/repository/postgres"
	redisrepo "go-game-backend/services/auth/internal/repository/redis"
)

// Config holds configuration for the authentication service.
type Config struct {
	PlayerLockTTL    time.Duration `yaml:"player-lock-ttl"`
	UserCreatedTopic string        `yaml:"user-created-topic"`
}

type playerLocker interface {
	DoWithPlayerLock(ctx context.Context, userID int64, f futils.CtxF) error
}

// Service provides authentication related operations such as registration,
// login and token refresh.
type Service struct {
	cfg           *Config
	pgStore       *postgresstore.Storage[postgresrepo.Repos]
	rxStore       *redisstore.Storage[redisrepo.Repos]
	playerLocker  playerLocker
	tokensFactory *tknfactory.TokensFactory
}

// New creates a new Service instance with the supplied dependencies.
func New(
	cfg *Config,
	pgStore *postgresstore.Storage[postgresrepo.Repos],
	rxStore *redisstore.Storage[redisrepo.Repos],
	playerLocker playerLocker,
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

	err = l.pgStore.DoTx(ctx, func(ctx context.Context, r *postgresrepo.Repos) error {
		userID, err = r.User().AddUser(ctx, req.LoginToken)
		if err != nil {
			return fmt.Errorf("add user failed: %w", err)
		}

		ev := models.UserCreatedEvent{UserID: userID}
		if err := r.Outbox().AddJSON(ctx, l.cfg.UserCreatedTopic, ev); err != nil {
			return fmt.Errorf("save outbox event: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pg transaction: %w", err)
	}

	sessionInfo, err := l.startSessionWithPlayerLock(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("start session with player lock: %w", err)
	}

	return sessionInfo, nil
}

// Login authenticates a user and starts a new session.
func (l *Service) Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error) {
	userID, err := l.pgStore.Raw().User().FindUserByLoginToken(ctx, req.LoginToken)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	sessionInfo, err := l.startSessionWithPlayerLock(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("start session with player lock: %w", err)
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

	err = l.rxStore.DoTx(ctx, func(ctx context.Context, r *redisrepo.Repos) error {
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
		return nil, fmt.Errorf("get session info: %w", err)
	}

	accessToken, expiresAt, err := l.tokensFactory.CreateAccessToken(sessionInfo.UserID, sessionInfo.SessionToken, utcNow)
	if err != nil {
		return nil, fmt.Errorf("creation access token: %w", err)
	}
	refreshToken, refreshTokenExpiresAt := l.tokensFactory.CreateRefreshToken(utcNow)

	err = l.rxStore.DoTx(ctx, func(ctx context.Context, r *redisrepo.Repos) error {
		err := r.Session().RemoveRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			return fmt.Errorf("remove refresh token: %w", err)
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
		ExpiresAtUnix: expiresAt.Unix(),
	}
	return resp, nil
}
