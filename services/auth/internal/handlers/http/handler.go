package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/auth/pkg/models"
	"go.uber.org/zap"
	"net/http"
)

type Logic interface {
	Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginResponse, err error)
}

type Handler struct {
	logic  Logic
	logger *logging.ZapLogger
}

func New(logic Logic, logger *logging.ZapLogger) *Handler {
	return &Handler{
		logic:  logic,
		logger: logger,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		return
	}

	resp, err := h.logic.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "failed to put item", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, resp)
}
