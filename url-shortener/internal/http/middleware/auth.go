package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/service"
)

const (
	authorizationHeader = "Authorization"
	userIDContextKey    = "user_id"
	userEmailContextKey = "user_email"
)

type TokenParser interface {
	ParseAccessToken(rawToken string) (service.TokenClaims, error)
}

func AuthRequired(authService TokenParser, logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader(authorizationHeader))
		if !ok {
			logger.Warn("authorization required",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
			writeUnauthorized(c)
			return
		}

		claims, err := authService.ParseAccessToken(token)
		if err != nil {
			logger.Warn("authorization rejected: invalid token",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
			writeUnauthorized(c)
			return
		}

		setCurrentUser(c, claims)
		c.Next()
	}
}

func OptionalAuth(authService TokenParser, logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeader)
		if strings.TrimSpace(header) == "" {
			c.Next()
			return
		}

		token, ok := bearerToken(header)
		if !ok {
			logger.Warn("optional authorization rejected: malformed bearer token",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
			writeUnauthorized(c)
			return
		}

		claims, err := authService.ParseAccessToken(token)
		if err != nil {
			logger.Warn("optional authorization rejected: invalid token",
				zap.Error(err),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
			writeUnauthorized(c)
			return
		}

		setCurrentUser(c, claims)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}

func CurrentUserEmail(c *gin.Context) (string, bool) {
	value, ok := c.Get(userEmailContextKey)
	if !ok {
		return "", false
	}

	email, ok := value.(string)
	return email, ok
}

func setCurrentUser(c *gin.Context, claims service.TokenClaims) {
	c.Set(userIDContextKey, claims.UserID)
	c.Set(userEmailContextKey, claims.Email)
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func writeUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "unauthorized",
			"message": "Valid bearer token is required",
		},
	})
}
