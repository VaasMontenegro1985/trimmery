package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

const testJWTSecret = "01234567890123456789012345678901"

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	storage := newFakeUserStorage()
	authService := NewAuthService(storage, testJWTSecret, time.Hour, zap.NewNop())
	authService.now = func() time.Time {
		return time.Date(2026, 5, 29, 9, 0, 0, 0, time.UTC)
	}

	registerResult, err := authService.Register(context.Background(), " User@Example.COM ", "strong-password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if registerResult.User.Email != "user@example.com" {
		t.Fatalf("registered email = %q, want normalized email", registerResult.User.Email)
	}

	if registerResult.AccessToken == "" {
		t.Fatal("Register() returned empty access token")
	}

	claims, err := authService.ParseAccessToken(registerResult.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if claims.UserID != registerResult.User.ID || claims.Email != registerResult.User.Email {
		t.Fatalf("claims = %+v, want registered user", claims)
	}

	loginResult, err := authService.Login(context.Background(), "user@example.com", "strong-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if loginResult.User.ID != registerResult.User.ID {
		t.Fatalf("login user id = %d, want %d", loginResult.User.ID, registerResult.User.ID)
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	storage := newFakeUserStorage()
	authService := NewAuthService(storage, testJWTSecret, time.Hour, zap.NewNop())

	_, err := authService.Login(context.Background(), "missing@example.com", "strong-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthServiceParseAccessTokenRejectsExpiredToken(t *testing.T) {
	storage := newFakeUserStorage()
	authService := NewAuthService(storage, testJWTSecret, time.Hour, zap.NewNop())
	authService.now = func() time.Time {
		return time.Date(2026, 5, 29, 9, 0, 0, 0, time.UTC)
	}

	result, err := authService.issueAuthResult(domain.User{ID: 1, Email: "user@example.com"})
	if err != nil {
		t.Fatalf("issueAuthResult() error = %v", err)
	}

	authService.now = func() time.Time {
		return time.Date(2026, 5, 29, 11, 0, 0, 0, time.UTC)
	}

	_, err = authService.ParseAccessToken(result.AccessToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{name: "valid", password: "12345678"},
		{name: "too short", password: "1234567", wantError: true},
		{name: "too long", password: string(make([]byte, maxBcryptPasswordBytes+1)), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if (err != nil) != tt.wantError {
				t.Fatalf("validatePassword() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

type fakeUserStorage struct {
	nextID int64
	users  map[string]domain.User
}

func newFakeUserStorage() *fakeUserStorage {
	return &fakeUserStorage{
		nextID: 1,
		users:  make(map[string]domain.User),
	}
}

func (s *fakeUserStorage) Create(_ context.Context, user domain.User) (domain.User, error) {
	if _, ok := s.users[user.Email]; ok {
		return domain.User{}, domain.ErrEmailAlreadyExists
	}

	user.ID = s.nextID
	user.CreatedAt = time.Now()
	s.nextID++
	s.users[user.Email] = user

	return user, nil
}

func (s *fakeUserStorage) FindByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := s.users[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}
