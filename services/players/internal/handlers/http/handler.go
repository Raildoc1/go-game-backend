package httphand

import (
	"context"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/players/internal/middleware"
	"go-game-backend/services/players/internal/ws"
	"go-game-backend/services/players/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
)

// Logic defines business logic required by handler.
type Logic interface {
	GetInitialState(ctx context.Context, userID int64) (*models.InitialState, error)
}

// Handler provides HTTP endpoints for players service.
type Handler struct {
	logic  Logic
	hub    *ws.Hub
	logger *logging.ZapLogger
}

// New creates new handler.
func New(logic Logic, hub *ws.Hub, logger *logging.ZapLogger) *Handler {
	return &Handler{logic: logic, hub: hub, logger: logger}
}

// GetInitialState handles initial state requests.
func (h *Handler) GetInitialState(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDCtxKey)
	state, err := h.logic.GetInitialState(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "get initial state", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, state)
}

// WS upgrades connection to websocket and registers it.
func (h *Handler) WS(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDCtxKey)
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Add(userID, conn)
}
