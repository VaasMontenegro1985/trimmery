package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/http/middleware"
	"url-shortener/internal/service"
)

func TestGetLinkQRReturnsPNG(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeLinkService{}
	linkHandler := NewLinkHandler(fake, zap.NewNop())
	router := gin.New()
	router.GET("/api/links/:code/qr", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.GetLinkQR)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/links/abc123/qr", nil)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	if ct := response.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
}

func TestGetLinkQRInvalidSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := &fakeLinkService{}
	linkHandler := NewLinkHandler(fake, zap.NewNop())
	router := gin.New()
	router.GET("/api/links/:code/qr", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.GetLinkQR)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/links/abc123/qr?size=10", nil)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
