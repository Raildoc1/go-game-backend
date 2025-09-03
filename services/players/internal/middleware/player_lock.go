package middleware

import (
	"context"
	"go-game-backend/pkg/futils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type playerLocker interface {
	DoWithPlayerLock(ctx context.Context, userID int64, f futils.CtxF) error
}

type PlayerLock struct {
	locker playerLocker
}

func NewPlayerLock(locker playerLocker) *PlayerLock {
	return &PlayerLock{locker: locker}
}

func (m *PlayerLock) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		idVal, ok := c.Get(UserIDCtxKey)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, ok := idVal.(int64)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		if err := m.locker.DoWithPlayerLock(c.Request.Context(), userID, func(ctx context.Context) error {
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return nil
		}); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}
