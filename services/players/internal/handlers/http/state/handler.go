package statehand

import (
	"context"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/players/internal/middleware"
	"go-game-backend/services/players/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

type Logic interface {
	GetInitialState(ctx context.Context, userID int64) (*models.InitialState, error)
	GetStateDelta(ctx context.Context, userID int64, fromStateID int) ([]any, error)
}

// Handler provides HTTP endpoints for getting player state.
type Handler struct {
	logic  Logic
	logger *logging.ZapLogger
}

func New(logic Logic,logger *logging.ZapLogger) *Handler {
	return &Handler{logic: logic, logger: logger}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/initial", h.GetInitialState)
	r.GET("/delta/:from_state_id", h.GetInitialState)
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

// GetStateDelta handles state delta requests.
func (h *Handler) GetStateDelta(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDCtxKey)
	fromStateID := c.GetInt("from_state_id")
	deltas, err := h.logic.GetStateDelta(c, userID, fromStateID)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "get state delta", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, deltas)
}