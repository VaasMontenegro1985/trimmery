package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
	"url-shortener/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, email string, password string) (service.AuthResult, error)
	Login(ctx context.Context, email string, password string) (service.AuthResult, error)
}

type AuthHandler struct {
	service AuthService
	logger  *zap.Logger
}

func NewAuthHandler(service AuthService, logger *zap.Logger) *AuthHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        userResponse `json:"user"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	h.logger.Debug("register handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)

	var request authRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	result, err := h.service.Register(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toAuthResponse(result))
}

func (h *AuthHandler) Login(c *gin.Context) {
	h.logger.Debug("login handler started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)

	var request authRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	result, err := h.service.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAuthResponse(result))
}

func (h *AuthHandler) writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(c, http.StatusBadRequest, "invalid_email", "Email must be valid")
	case errors.Is(err, service.ErrInvalidPassword):
		writeError(c, http.StatusBadRequest, "invalid_password", "Password must be 8-72 bytes")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		writeError(c, http.StatusConflict, "email_already_exists", "Email already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect")
	default:
		h.logger.Error("auth request failed",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}

func toAuthResponse(result service.AuthResult) authResponse {
	return authResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresAt:   result.ExpiresAt,
		User: userResponse{
			ID:    result.User.ID,
			Email: result.User.Email,
		},
	}
}
