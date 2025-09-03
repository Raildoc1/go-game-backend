package middleware

import (
	"go-game-backend/pkg/logging"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	playersvc "go-game-backend/services/players/internal/services/player"
)

// JWTConfig defines JWT verification parameters.
type JWTConfig struct {
	Algorithm string
	Secret    string
}

// SessionValidator validates player sessions.
type SessionValidator struct {
	cfg      *JWTConfig
	sessions playersvc.SessionRepository
	logger   *logging.ZapLogger
}

// NewSessionValidator creates new validator.
func NewSessionValidator(cfg *JWTConfig, sessions playersvc.SessionRepository, logger *logging.ZapLogger) *SessionValidator {
	return &SessionValidator{cfg: cfg, sessions: sessions, logger: logger}
}

// UserIDCtxKey is context key storing authenticated user ID.
const UserIDCtxKey = "userID"

// Middleware returns gin middleware for session validation.
func (m *SessionValidator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(m.cfg.Secret), nil
		}, jwt.WithValidMethods([]string{m.cfg.Algorithm}))
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uidVal, ok := claims["userID"].(float64)
		sessionStr, ok2 := claims["session"].(string)
		if !ok || !ok2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID := int64(uidVal)
		sessionID, err := uuid.Parse(sessionStr)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		stored, err := m.sessions.GetSessionToken(c.Request.Context(), userID)
		if err != nil || stored != sessionID {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(UserIDCtxKey, userID)
		c.Next()
	}
}
