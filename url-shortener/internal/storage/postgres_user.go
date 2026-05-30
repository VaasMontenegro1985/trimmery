package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

type PostgresUserStorage struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresUserStorage(db *pgxpool.Pool, logger *zap.Logger) *PostgresUserStorage {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PostgresUserStorage{
		db:     db,
		logger: logger,
	}
}

func (s *PostgresUserStorage) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
	`

	s.logger.Debug("storage user insert started", zap.String("email", user.Email))

	err := s.db.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, fmt.Errorf("insert user: %w", domain.ErrEmailAlreadyExists)
		}

		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	s.logger.Debug("storage user insert completed",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, nil
}

func (s *PostgresUserStorage) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	s.logger.Debug("storage user lookup started", zap.String("email", email))

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("select user by email: %w", domain.ErrUserNotFound)
		}

		return domain.User{}, fmt.Errorf("select user by email: %w", err)
	}

	s.logger.Debug("storage user lookup completed",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, nil
}
