package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

const uniqueViolationCode = "23505"

type PostgresLinkStorage struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresLinkStorage(db *pgxpool.Pool, logger *zap.Logger) *PostgresLinkStorage {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PostgresLinkStorage{
		db:     db,
		logger: logger,
	}
}

func (s *PostgresLinkStorage) Create(ctx context.Context, link domain.Link) (domain.Link, error) {
	const query = `
		INSERT INTO links (code, original_url, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, code, original_url, user_id, clicks_count, deleted_at, created_at
	`

	s.logger.Debug("storage link insert started", zap.String("code", link.Code))

	var userID sql.NullInt64
	var deletedAt sql.NullTime
	err := s.db.QueryRow(ctx, query, link.Code, link.OriginalURL, link.UserID).Scan(
		&link.ID,
		&link.Code,
		&link.OriginalURL,
		&userID,
		&link.ClicksCount,
		&deletedAt,
		&link.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Link{}, fmt.Errorf("insert link: %w", domain.ErrCodeAlreadyExists)
		}

		return domain.Link{}, fmt.Errorf("insert link: %w", err)
	}

	link.UserID = nullableInt64Ptr(userID)
	link.DeletedAt = nullableTimePtr(deletedAt)

	s.logger.Debug("storage link insert completed",
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
	)

	return link, nil
}

func (s *PostgresLinkStorage) FindByCode(ctx context.Context, code string) (domain.Link, error) {
	const query = `
		SELECT id, code, original_url, user_id, clicks_count, deleted_at, created_at
		FROM links
		WHERE code = $1
	`

	var link domain.Link
	var userID sql.NullInt64
	var deletedAt sql.NullTime
	s.logger.Debug("storage link lookup started", zap.String("code", code))

	err := s.db.QueryRow(ctx, query, code).Scan(
		&link.ID,
		&link.Code,
		&link.OriginalURL,
		&userID,
		&link.ClicksCount,
		&deletedAt,
		&link.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Link{}, fmt.Errorf("select link by code: %w", domain.ErrLinkNotFound)
		}

		return domain.Link{}, fmt.Errorf("select link by code: %w", err)
	}

	link.UserID = nullableInt64Ptr(userID)
	link.DeletedAt = nullableTimePtr(deletedAt)

	s.logger.Debug("storage link lookup completed",
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
	)

	return link, nil
}

func (s *PostgresLinkStorage) ListByUserID(ctx context.Context, userID int64) ([]domain.Link, error) {
	const query = `
		SELECT id, code, original_url, user_id, clicks_count, deleted_at, created_at
		FROM links
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	s.logger.Debug("storage user links lookup started", zap.Int64("user_id", userID))

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("select links by user id: %w", err)
	}
	defer rows.Close()

	links := make([]domain.Link, 0)
	for rows.Next() {
		var link domain.Link
		var nullableUserID sql.NullInt64
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&link.ID,
			&link.Code,
			&link.OriginalURL,
			&nullableUserID,
			&link.ClicksCount,
			&deletedAt,
			&link.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan link by user id: %w", err)
		}

		link.UserID = nullableInt64Ptr(nullableUserID)
		link.DeletedAt = nullableTimePtr(deletedAt)
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate links by user id: %w", err)
	}

	s.logger.Debug("storage user links lookup completed",
		zap.Int64("user_id", userID),
		zap.Int("links_count", len(links)),
	)

	return links, nil
}

func (s *PostgresLinkStorage) FindActiveByCode(ctx context.Context, code string) (domain.Link, error) {
	const query = `
		SELECT id, code, original_url, user_id, clicks_count, deleted_at, created_at
		FROM links
		WHERE code = $1 AND deleted_at IS NULL
	`

	var link domain.Link
	var userID sql.NullInt64
	var deletedAt sql.NullTime
	s.logger.Debug("storage active link lookup started", zap.String("code", code))

	err := s.db.QueryRow(ctx, query, code).Scan(
		&link.ID,
		&link.Code,
		&link.OriginalURL,
		&userID,
		&link.ClicksCount,
		&deletedAt,
		&link.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Link{}, fmt.Errorf("select active link by code: %w", domain.ErrLinkNotFound)
		}

		return domain.Link{}, fmt.Errorf("select active link by code: %w", err)
	}

	link.UserID = nullableInt64Ptr(userID)
	link.DeletedAt = nullableTimePtr(deletedAt)

	return link, nil
}

func (s *PostgresLinkStorage) UpdateLink(ctx context.Context, userID int64, currentCode string, newOriginalURL *string, newCode *string) (domain.Link, error) {
	const query = `
		UPDATE links
		SET
			original_url = COALESCE($1, original_url),
			code = COALESCE($2, code)
		WHERE code = $3
			AND deleted_at IS NULL
			AND user_id = $4
		RETURNING id, code, original_url, user_id, clicks_count, deleted_at, created_at
	`

	var link domain.Link
	var nullableUserID sql.NullInt64
	var deletedAt sql.NullTime

	err := s.db.QueryRow(ctx, query, newOriginalURL, newCode, currentCode, userID).Scan(
		&link.ID,
		&link.Code,
		&link.OriginalURL,
		&nullableUserID,
		&link.ClicksCount,
		&deletedAt,
		&link.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Link{}, fmt.Errorf("update link: %w", domain.ErrCodeAlreadyExists)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Link{}, fmt.Errorf("update link: %w", domain.ErrLinkNotFound)
		}

		return domain.Link{}, fmt.Errorf("update link: %w", err)
	}

	link.UserID = nullableInt64Ptr(nullableUserID)
	link.DeletedAt = nullableTimePtr(deletedAt)

	return link, nil
}

func (s *PostgresLinkStorage) SoftDeleteLink(ctx context.Context, userID int64, code string) error {
	const query = `
		UPDATE links
		SET deleted_at = now()
		WHERE code = $1
			AND deleted_at IS NULL
			AND user_id = $2
	`

	tag, err := s.db.Exec(ctx, query, code, userID)
	if err != nil {
		return fmt.Errorf("soft delete link: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft delete link: %w", domain.ErrLinkNotFound)
	}

	return nil
}

func (s *PostgresLinkStorage) RecordVisit(ctx context.Context, linkID int64, ip string, userAgent *string) error {
	s.logger.Debug("storage visit insert started", zap.Int64("link_id", linkID))

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin visit transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const insertVisitQuery = `
		INSERT INTO visits (link_id, ip, user_agent)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var visitID int64
	err = tx.QueryRow(ctx, insertVisitQuery, linkID, ipValueForStorage(ip), userAgent).Scan(&visitID)
	if err != nil {
		return fmt.Errorf("insert visit: %w", err)
	}

	const updateClicksQuery = `
		UPDATE links
		SET clicks_count = clicks_count + 1
		WHERE id = $1
	`

	if _, err := tx.Exec(ctx, updateClicksQuery, linkID); err != nil {
		return fmt.Errorf("increment clicks count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit visit transaction: %w", err)
	}

	s.logger.Debug("storage visit insert completed",
		zap.Int64("link_id", linkID),
		zap.Int64("visit_id", visitID),
	)

	return nil
}

func (s *PostgresLinkStorage) ListVisitsByLinkID(ctx context.Context, linkID int64, limit int) ([]domain.Visit, error) {
	const query = `
		SELECT id, link_id, visited_at, ip::text, user_agent
		FROM visits
		WHERE link_id = $1
		ORDER BY visited_at DESC
		LIMIT $2
	`

	s.logger.Debug("storage visits lookup started",
		zap.Int64("link_id", linkID),
		zap.Int("limit", limit),
	)

	rows, err := s.db.Query(ctx, query, linkID, limit)
	if err != nil {
		return nil, fmt.Errorf("select visits by link id: %w", err)
	}
	defer rows.Close()

	visits := make([]domain.Visit, 0)
	for rows.Next() {
		var visit domain.Visit
		var ip sql.NullString
		var userAgent sql.NullString

		if err := rows.Scan(
			&visit.ID,
			&visit.LinkID,
			&visit.VisitedAt,
			&ip,
			&userAgent,
		); err != nil {
			return nil, fmt.Errorf("scan visit: %w", err)
		}

		visit.IP = nullableStringPtr(ip)
		visit.UserAgent = nullableStringPtr(userAgent)
		visits = append(visits, visit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visits by link id: %w", err)
	}

	s.logger.Debug("storage visits lookup completed",
		zap.Int64("link_id", linkID),
		zap.Int("visits_count", len(visits)),
	)

	return visits, nil
}

func ipValueForStorage(ip string) any {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil
	}

	return parsed
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

func nullableTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func nullableInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64
	return &result
}
