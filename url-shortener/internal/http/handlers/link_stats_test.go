package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/http/middleware"
	"url-shortener/internal/service"
)

func TestListUserLinksReturnsClicksCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkHandler := NewLinkHandler(&fakeLinkService{}, zap.NewNop())
	router := gin.New()
	router.GET("/api/links", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.ListUserLinks)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	if !strings.Contains(response.Body.String(), `"clicks_count":5`) {
		t.Fatalf("body = %s, want clicks_count field", response.Body.String())
	}
}

func TestGetLinkStatsReturnsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkHandler := NewLinkHandler(&fakeLinkService{}, zap.NewNop())
	router := gin.New()
	router.GET("/api/links/:code/stats", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.GetLinkStats)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/links/forbidden/stats", nil)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestGetLinkStatsReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkHandler := NewLinkHandler(&fakeLinkService{}, zap.NewNop())
	router := gin.New()
	router.GET("/api/links/:code/stats", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.GetLinkStats)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/links/abc123/stats", nil)
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
}

func TestRedirectReturnsFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkHandler := NewLinkHandler(&fakeLinkService{}, zap.NewNop())
	router := gin.New()
	router.GET("/:code", linkHandler.Redirect)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
}

func TestUpdateLinkReturnsBadRequestOnNothingToUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkHandler := NewLinkHandler(&fakeLinkService{}, zap.NewNop())
	router := gin.New()
	router.PATCH("/api/links/:code", middleware.AuthRequired(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 1, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.UpdateLink)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/links/abc123", bytes.NewBufferString(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer token")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
