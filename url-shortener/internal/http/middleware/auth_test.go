package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"url-shortener/internal/service"
)

func TestAuthRequiredAcceptsValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/protected", AuthRequired(fakeTokenParser{
		claims: service.TokenClaims{UserID: 7, Email: "user@example.com"},
	}, zap.NewNop()), func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			t.Fatal("CurrentUserID() ok = false, want true")
		}

		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestAuthRequiredRejectsMissingBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/protected", AuthRequired(fakeTokenParser{}, zap.NewNop()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestOptionalAuthAllowsAnonymousRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/public", OptionalAuth(fakeTokenParser{}, zap.NewNop()), func(c *gin.Context) {
		if _, ok := CurrentUserID(c); ok {
			t.Fatal("CurrentUserID() ok = true, want false")
		}

		c.Status(http.StatusOK)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/public", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestOptionalAuthRejectsInvalidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/public", OptionalAuth(fakeTokenParser{err: service.ErrInvalidToken}, zap.NewNop()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/public", nil)
	request.Header.Set("Authorization", "Bearer bad-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

type fakeTokenParser struct {
	claims service.TokenClaims
	err    error
}

func (p fakeTokenParser) ParseAccessToken(_ string) (service.TokenClaims, error) {
	if p.err != nil {
		return service.TokenClaims{}, errors.Join(service.ErrInvalidToken, p.err)
	}

	return p.claims, nil
}
