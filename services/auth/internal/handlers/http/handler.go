package httphand

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/auth/pkg/models"
	"go.uber.org/zap"
	"net/http"
)

type Logic interface {
	Register(ctx context.Context, req *models.RegisterRequest) (resp *models.LoginRespose, err error)
	Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error)
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (resp *models.LoginRespose, err error)
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

func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		return
	}

	resp, err := h.logic.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "failed to register", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		return
	}

	resp, err := h.logic.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "failed to login", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.BindJSON(&req); err != nil {
		return
	}

	resp, err := h.logic.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorCtx(c.Request.Context(), "failed to refresh token", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, resp)
}
