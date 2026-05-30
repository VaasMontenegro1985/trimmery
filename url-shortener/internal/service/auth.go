package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"url-shortener/internal/domain"
)

const (
	minPasswordLength      = 8
	maxBcryptPasswordBytes = 72
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
)

type UserStorage interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type AuthResult struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
	User        domain.User
}

type TokenClaims struct {
	UserID int64
	Email  string
}

type AuthService struct {
	storage      UserStorage
	jwtSecret    string
	jwtAccessTTL time.Duration
	logger       *zap.Logger
	now          func() time.Time
}

func NewAuthService(storage UserStorage, jwtSecret string, jwtAccessTTL time.Duration, logger *zap.Logger) *AuthService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AuthService{
		storage:      storage,
		jwtSecret:    jwtSecret,
		jwtAccessTTL: jwtAccessTTL,
		logger:       logger,
		now:          time.Now,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (AuthResult, error) {
	s.logger.Debug("user registration started")

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		s.logger.Warn("user registration rejected: invalid email")
		return AuthResult{}, err
	}

	if err := validatePassword(password); err != nil {
		s.logger.Warn("user registration rejected: invalid password", zap.String("email", normalizedEmail))
		return AuthResult{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("password hash failed", zap.Error(err), zap.String("email", normalizedEmail))
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.storage.Create(ctx, domain.User{
		Email:        normalizedEmail,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			s.logger.Warn("user registration rejected: email already exists", zap.String("email", normalizedEmail))
			return AuthResult{}, err
		}

		s.logger.Error("user registration storage failed", zap.Error(err), zap.String("email", normalizedEmail))
		return AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	result, err := s.issueAuthResult(user)
	if err != nil {
		s.logger.Error("access token issue failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return AuthResult{}, err
	}

	s.logger.Info("user registered", zap.Int64("user_id", user.ID), zap.String("email", user.Email))

	return result, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (AuthResult, error) {
	s.logger.Debug("user login started")

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		s.logger.Warn("user login rejected: invalid email")
		return AuthResult{}, err
	}

	if err := validatePassword(password); err != nil {
		s.logger.Warn("user login rejected: invalid password", zap.String("email", normalizedEmail))
		return AuthResult{}, err
	}

	user, err := s.storage.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Warn("user login rejected: invalid credentials", zap.String("email", normalizedEmail))
			return AuthResult{}, domain.ErrInvalidCredentials
		}

		s.logger.Error("user login storage failed", zap.Error(err), zap.String("email", normalizedEmail))
		return AuthResult{}, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("user login rejected: invalid credentials", zap.String("email", normalizedEmail))
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	result, err := s.issueAuthResult(user)
	if err != nil {
		s.logger.Error("access token issue failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return AuthResult{}, err
	}

	s.logger.Info("user logged in", zap.Int64("user_id", user.ID), zap.String("email", user.Email))

	return result, nil
}

func (s *AuthService) ParseAccessToken(rawToken string) (TokenClaims, error) {
	var claims accessTokenClaims

	token, err := jwt.ParseWithClaims(
		rawToken,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}

			return []byte(s.jwtSecret), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.now),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return TokenClaims{}, fmt.Errorf("parse access token: %w", ErrInvalidToken)
	}

	if token == nil || !token.Valid {
		return TokenClaims{}, fmt.Errorf("validate access token: %w", ErrInvalidToken)
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return TokenClaims{}, fmt.Errorf("parse token subject: %w", ErrInvalidToken)
	}

	if claims.Email == "" {
		return TokenClaims{}, fmt.Errorf("validate token email: %w", ErrInvalidToken)
	}

	return TokenClaims{
		UserID: userID,
		Email:  claims.Email,
	}, nil
}

func (s *AuthService) issueAuthResult(user domain.User) (AuthResult, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.jwtAccessTTL)

	claims := accessTokenClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return AuthResult{}, fmt.Errorf("sign access token: %w", err)
	}

	s.logger.Debug("access token issued", zap.Int64("user_id", user.ID), zap.Time("expires_at", expiresAt))

	return AuthResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user,
	}, nil
}

type accessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func normalizeEmail(email string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(email))
	if value == "" {
		return "", ErrInvalidEmail
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", ErrInvalidEmail
	}

	return value, nil
}

func validatePassword(password string) error {
	passwordLength := len([]byte(password))
	if passwordLength < minPasswordLength || passwordLength > maxBcryptPasswordBytes {
		return ErrInvalidPassword
	}

	return nil
}
