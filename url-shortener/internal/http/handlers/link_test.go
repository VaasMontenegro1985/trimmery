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
	"url-shortener/internal/http/middleware"
	"url-shortener/internal/service"
)

func TestShortenAllowsAnonymousRequest(t *testing.T) {
	recorder := performShortenRequest(t, "")

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestShortenAcceptsValidBearerToken(t *testing.T) {
	recorder := performShortenRequest(t, "Bearer valid-token")

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestShortenRejectsInvalidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	linkService := &fakeLinkService{}
	linkHandler := NewLinkHandler(linkService, zap.NewNop())
	router := gin.New()
	router.POST("/shorten", middleware.OptionalAuth(fakeLinkTokenParser{err: service.ErrInvalidToken}, zap.NewNop()), linkHandler.Shorten)

	request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"https://example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer bad-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func performShortenRequest(t *testing.T, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	linkService := &fakeLinkService{}
	linkHandler := NewLinkHandler(linkService, zap.NewNop())
	router := gin.New()
	router.POST("/shorten", middleware.OptionalAuth(fakeLinkTokenParser{
		claims: service.TokenClaims{UserID: 42, Email: "user@example.com"},
	}, zap.NewNop()), linkHandler.Shorten)

	request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"https://example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

type fakeLinkService struct{}

func (s *fakeLinkService) CreateShortLink(_ context.Context, rawURL string, _ *int64) (service.ShortLink, error) {
	if rawURL == "" {
		return service.ShortLink{}, service.ErrInvalidURL
	}

	return service.ShortLink{
		Code:        "abc123",
		ShortURL:    "http://localhost:8080/abc123",
		OriginalURL: rawURL,
	}, nil
}

func (s *fakeLinkService) CreateCustomShortLink(_ context.Context, rawURL string, customCode string, _ *int64) (service.ShortLink, error) {
	if customCode == "" {
		return service.ShortLink{}, service.ErrInvalidCode
	}

	return service.ShortLink{
		Code:        customCode,
		ShortURL:    "http://localhost:8080/" + customCode,
		OriginalURL: rawURL,
	}, nil
}

func (s *fakeLinkService) RedirectByCode(_ context.Context, _ string, _ service.VisitMeta) (domain.Link, error) {
	return domain.Link{
		ID:          1,
		Code:        "abc123",
		OriginalURL: "https://example.com",
	}, nil
}

func (s *fakeLinkService) GetLinkStats(_ context.Context, userID int64, code string, _ int) (service.LinkStats, error) {
	if code == "forbidden" {
		return service.LinkStats{}, domain.ErrLinkForbidden
	}

	return service.LinkStats{
		Code:        code,
		ShortURL:    "http://localhost:8080/" + code,
		OriginalURL: "https://example.com",
		ClicksCount: 2,
		Visits:      []service.VisitEntry{{VisitedAt: time.Now()}},
	}, nil
}

func (s *fakeLinkService) ListUserLinks(_ context.Context, _ int64) ([]service.ShortLink, error) {
	return []service.ShortLink{{
		Code:        "abc123",
		ShortURL:    "http://localhost:8080/abc123",
		OriginalURL: "https://example.com",
		ClicksCount: 5,
	}}, nil
}

func (s *fakeLinkService) UpdateLink(_ context.Context, _ int64, _ string, patch service.LinkUpdate) (service.ShortLink, error) {
	if patch.OriginalURL == nil && patch.Code == nil {
		return service.ShortLink{}, service.ErrNothingToUpdate
	}

	return service.ShortLink{
		Code:        "newCode",
		ShortURL:    "http://localhost:8080/newCode",
		OriginalURL: "https://example.com",
		ClicksCount: 0,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *fakeLinkService) DeleteLink(_ context.Context, _ int64, _ string) error {
	return nil
}

func (s *fakeLinkService) GenerateLinkQR(_ context.Context, _ int64, code string, size int) ([]byte, error) {
	if size > 0 && size < 128 {
		return nil, service.ErrInvalidSize
	}

	return []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, nil
}

type fakeLinkTokenParser struct {
	claims service.TokenClaims
	err    error
}

func (p fakeLinkTokenParser) ParseAccessToken(_ string) (service.TokenClaims, error) {
	if p.err != nil {
		return service.TokenClaims{}, errors.Join(service.ErrInvalidToken, p.err)
	}

	return p.claims, nil
}
