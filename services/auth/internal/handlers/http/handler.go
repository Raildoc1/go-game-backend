// Package httphand contains HTTP handlers for the auth service.
package httphand

import (
	"context"
	"go-game-backend/pkg/logging"
	"go-game-backend/services/auth/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

// Logic defines the business logic required by the HTTP handler.
type Logic interface {
	Register(ctx context.Context, req *models.RegisterRequest) (resp *models.LoginRespose, err error)
	Login(ctx context.Context, req *models.LoginRequest) (resp *models.LoginRespose, err error)
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (resp *models.LoginRespose, err error)
}

// Handler provides HTTP endpoints for authentication operations.
type Handler struct {
	logic  Logic
	logger *logging.ZapLogger
}

// New creates a new HTTP handler with the provided logic implementation and logger.
func New(logic Logic, logger *logging.ZapLogger) *Handler {
	return &Handler{
		logic:  logic,
		logger: logger,
	}
}

// Register handles user registration requests.
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

// Login processes user login requests.
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

// RefreshToken handles token refresh requests.
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
