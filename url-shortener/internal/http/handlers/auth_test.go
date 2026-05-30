package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
	"url-shortener/internal/service"
)

func TestRegisterReturnsCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := NewAuthHandler(fakeAuthService{}, zap.NewNop())
	router := gin.New()
	router.POST("/auth/register", authHandler.Register)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"strong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
}

func TestLoginReturnsUnauthorizedForInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := NewAuthHandler(fakeAuthService{err: domain.ErrInvalidCredentials}, zap.NewNop())
	router := gin.New()
	router.POST("/auth/login", authHandler.Login)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRegisterReturnsBadRequestForInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := NewAuthHandler(fakeAuthService{}, zap.NewNop())
	router := gin.New()
	router.POST("/auth/register", authHandler.Register)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{bad-json}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

type fakeAuthService struct {
	err error
}

func (s fakeAuthService) Register(_ context.Context, email string, _ string) (service.AuthResult, error) {
	if s.err != nil {
		return service.AuthResult{}, s.err
	}

	return authResultForTest(email), nil
}

func (s fakeAuthService) Login(_ context.Context, email string, _ string) (service.AuthResult, error) {
	if s.err != nil {
		return service.AuthResult{}, errors.Join(domain.ErrInvalidCredentials, s.err)
	}

	return authResultForTest(email), nil
}

func authResultForTest(email string) service.AuthResult {
	return service.AuthResult{
		AccessToken: "token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Date(2026, 5, 30, 9, 0, 0, 0, time.UTC),
		User: domain.User{
			ID:    1,
			Email: email,
		},
	}
}
