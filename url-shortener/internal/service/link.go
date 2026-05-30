package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

const (
	codeLength        = 6
	maxCreateAttempts = 5
	maxCodeLength     = 32
	maxURLLength      = 2048
)

var (
	ErrInvalidURL            = errors.New("invalid url")
	ErrInvalidCode           = errors.New("invalid code")
	ErrShortCodeCreateFailed = errors.New("short code create failed")
	ErrInvalidLimit          = errors.New("invalid limit")
)

var codePattern = regexp.MustCompile(`^[0-9a-zA-Z]+$`)

type LinkStorage interface {
	Create(ctx context.Context, link domain.Link) (domain.Link, error)
	FindByCode(ctx context.Context, code string) (domain.Link, error)
	FindActiveByCode(ctx context.Context, code string) (domain.Link, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.Link, error)
	UpdateLink(ctx context.Context, userID int64, currentCode string, newOriginalURL *string, newCode *string) (domain.Link, error)
	SoftDeleteLink(ctx context.Context, userID int64, code string) error
	RecordVisit(ctx context.Context, linkID int64, ip string, userAgent *string) error
	ListVisitsByLinkID(ctx context.Context, linkID int64, limit int) ([]domain.Visit, error)
}

type ShortLink struct {
	Code        string
	ShortURL    string
	OriginalURL string
	ClicksCount int64
	QRURL       string
	CreatedAt   time.Time
}

type VisitEntry struct {
	VisitedAt time.Time
	IP        *string
	UserAgent *string
}

type LinkStats struct {
	Code        string
	ShortURL    string
	OriginalURL string
	ClicksCount int64
	Visits      []VisitEntry
}

type LinkService struct {
	storage LinkStorage
	baseURL string
	logger  *zap.Logger
}

func NewLinkService(storage LinkStorage, baseURL string, logger *zap.Logger) *LinkService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LinkService{
		storage: storage,
		baseURL: baseURL,
		logger:  logger,
	}
}

func (s *LinkService) CreateShortLink(ctx context.Context, rawURL string, userID *int64) (ShortLink, error) {
	s.logger.Debug("short link creation started")

	originalURL, err := normalizeURL(rawURL)
	if err != nil {
		return ShortLink{}, err
	}

	parsedURL, _ := url.Parse(originalURL)
	s.logger.Debug("url validated",
		zap.String("url_scheme", parsedURL.Scheme),
		zap.String("url_host", parsedURL.Host),
	)

	for attempt := 1; attempt <= maxCreateAttempts; attempt++ {
		code, err := generateBase62Code(codeLength)
		if err != nil {
			return ShortLink{}, fmt.Errorf("generate short code: %w", err)
		}

		s.logger.Debug("short code generated",
			zap.String("code", code),
			zap.Int("attempt", attempt),
		)

		link, err := s.storage.Create(ctx, domain.Link{
			Code:        code,
			OriginalURL: originalURL,
			UserID:      userID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrCodeAlreadyExists) {
				s.logger.Debug("short code collision",
					zap.String("code", code),
					zap.Int("attempt", attempt),
				)
				continue
			}

			return ShortLink{}, fmt.Errorf("create link in storage: %w", err)
		}

		shortLink := ShortLink{
			Code:        link.Code,
			ShortURL:    strings.TrimRight(s.baseURL, "/") + "/" + link.Code,
			OriginalURL: link.OriginalURL,
			ClicksCount: link.ClicksCount,
			CreatedAt:   link.CreatedAt,
		}

		logFields := []zap.Field{
			zap.Int64("link_id", link.ID),
			zap.String("code", link.Code),
			zap.String("short_url", shortLink.ShortURL),
		}
		if link.UserID != nil {
			logFields = append(logFields, zap.Int64("user_id", *link.UserID))
		}

		s.logger.Info("short link created", logFields...)

		return shortLink, nil
	}

	return ShortLink{}, fmt.Errorf("create unique short code after %d attempts: %w", maxCreateAttempts, ErrShortCodeCreateFailed)
}

func (s *LinkService) CreateCustomShortLink(ctx context.Context, rawURL string, customCode string, userID *int64) (ShortLink, error) {
	s.logger.Debug("custom short link creation started")

	originalURL, err := normalizeURL(rawURL)
	if err != nil {
		return ShortLink{}, err
	}

	normalizedCode, err := normalizeCode(customCode)
	if err != nil {
		return ShortLink{}, err
	}

	link, err := s.storage.Create(ctx, domain.Link{
		Code:        normalizedCode,
		OriginalURL: originalURL,
		UserID:      userID,
	})
	if err != nil {
		return ShortLink{}, fmt.Errorf("create custom link in storage: %w", err)
	}

	shortLink := ShortLink{
		Code:        link.Code,
		ShortURL:    strings.TrimRight(s.baseURL, "/") + "/" + link.Code,
		OriginalURL: link.OriginalURL,
		ClicksCount: link.ClicksCount,
		CreatedAt:   link.CreatedAt,
	}

	s.logger.Info("custom short link created",
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
	)

	return shortLink, nil
}

func (s *LinkService) FindLinkByCode(ctx context.Context, code string) (domain.Link, error) {
	normalizedCode, err := normalizeCode(code)
	if err != nil {
		return domain.Link{}, err
	}

	s.logger.Debug("short link lookup started", zap.String("code", normalizedCode))

	link, err := s.storage.FindByCode(ctx, normalizedCode)
	if err != nil {
		return domain.Link{}, fmt.Errorf("find link by code in storage: %w", err)
	}

	s.logger.Debug("short link lookup completed",
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
	)

	return link, nil
}

func (s *LinkService) RedirectByCode(ctx context.Context, code string, meta VisitMeta) (domain.Link, error) {
	normalizedCode, err := normalizeCode(code)
	if err != nil {
		return domain.Link{}, err
	}

	link, err := s.storage.FindActiveByCode(ctx, normalizedCode)
	if err != nil {
		return domain.Link{}, fmt.Errorf("find active link by code in storage: %w", err)
	}

	ip, userAgent := normalizeVisitMeta(meta, s.logger)
	if err := s.storage.RecordVisit(ctx, link.ID, ip, userAgent); err != nil {
		s.logger.Warn("visit recording failed",
			zap.Error(err),
			zap.Int64("link_id", link.ID),
			zap.String("code", link.Code),
		)
	} else {
		link.ClicksCount++
		s.logger.Info("visit recorded",
			zap.Int64("link_id", link.ID),
			zap.String("code", link.Code),
			zap.Int64("clicks_count", link.ClicksCount),
		)
	}

	return link, nil
}

func (s *LinkService) GetLinkStats(ctx context.Context, userID int64, code string, limit int) (LinkStats, error) {
	normalizedLimit, err := normalizeVisitLimit(limit)
	if err != nil {
		return LinkStats{}, err
	}

	normalizedCode, err := normalizeCode(code)
	if err != nil {
		return LinkStats{}, err
	}

	s.logger.Debug("link stats lookup started",
		zap.Int64("user_id", userID),
		zap.String("code", normalizedCode),
	)

	link, err := s.storage.FindActiveByCode(ctx, normalizedCode)
	if err != nil {
		return LinkStats{}, fmt.Errorf("find active link by code in storage: %w", err)
	}

	if link.UserID == nil || *link.UserID != userID {
		s.logger.Warn("link stats rejected: forbidden",
			zap.Int64("user_id", userID),
			zap.String("code", normalizedCode),
		)
		return LinkStats{}, domain.ErrLinkForbidden
	}

	visits, err := s.storage.ListVisitsByLinkID(ctx, link.ID, normalizedLimit)
	if err != nil {
		return LinkStats{}, fmt.Errorf("list visits by link id in storage: %w", err)
	}

	entries := make([]VisitEntry, 0, len(visits))
	for _, visit := range visits {
		entries = append(entries, VisitEntry{
			VisitedAt: visit.VisitedAt,
			IP:        visit.IP,
			UserAgent: visit.UserAgent,
		})
	}

	stats := LinkStats{
		Code:        link.Code,
		ShortURL:    strings.TrimRight(s.baseURL, "/") + "/" + link.Code,
		OriginalURL: link.OriginalURL,
		ClicksCount: link.ClicksCount,
		Visits:      entries,
	}

	s.logger.Info("link stats viewed",
		zap.Int64("user_id", userID),
		zap.String("code", link.Code),
		zap.Int64("clicks_count", link.ClicksCount),
		zap.Int("visits_count", len(entries)),
	)

	return stats, nil
}

type LinkUpdate struct {
	OriginalURL *string
	Code        *string
}

var ErrNothingToUpdate = errors.New("nothing to update")

func (s *LinkService) UpdateLink(ctx context.Context, userID int64, currentCode string, patch LinkUpdate) (ShortLink, error) {
	if patch.OriginalURL == nil && patch.Code == nil {
		return ShortLink{}, ErrNothingToUpdate
	}

	normalizedCurrentCode, err := normalizeCode(currentCode)
	if err != nil {
		return ShortLink{}, err
	}

	var newURL *string
	if patch.OriginalURL != nil {
		normalizedURL, err := normalizeURL(*patch.OriginalURL)
		if err != nil {
			return ShortLink{}, err
		}
		newURL = &normalizedURL
	}

	var newCode *string
	if patch.Code != nil {
		normalizedCode, err := normalizeCode(*patch.Code)
		if err != nil {
			return ShortLink{}, err
		}
		newCode = &normalizedCode
	}

	link, err := s.storage.UpdateLink(ctx, userID, normalizedCurrentCode, newURL, newCode)
	if err != nil {
		return ShortLink{}, fmt.Errorf("update link in storage: %w", err)
	}

	s.logger.Info("link updated",
		zap.Int64("user_id", userID),
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
	)

	return ShortLink{
		Code:        link.Code,
		ShortURL:    strings.TrimRight(s.baseURL, "/") + "/" + link.Code,
		OriginalURL: link.OriginalURL,
		ClicksCount: link.ClicksCount,
		CreatedAt:   link.CreatedAt,
	}, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, userID int64, code string) error {
	normalizedCode, err := normalizeCode(code)
	if err != nil {
		return err
	}

	if err := s.storage.SoftDeleteLink(ctx, userID, normalizedCode); err != nil {
		return fmt.Errorf("soft delete link in storage: %w", err)
	}

	s.logger.Info("link soft deleted",
		zap.Int64("user_id", userID),
		zap.String("code", normalizedCode),
	)

	return nil
}

func (s *LinkService) ListUserLinks(ctx context.Context, userID int64) ([]ShortLink, error) {
	s.logger.Debug("user links lookup started", zap.Int64("user_id", userID))

	links, err := s.storage.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list links by user id in storage: %w", err)
	}

	shortLinks := make([]ShortLink, 0, len(links))
	for _, link := range links {
		shortLinks = append(shortLinks, ShortLink{
			Code:        link.Code,
			ShortURL:    strings.TrimRight(s.baseURL, "/") + "/" + link.Code,
			OriginalURL: link.OriginalURL,
			ClicksCount: link.ClicksCount,
			QRURL:       buildQRURL(link.Code),
			CreatedAt:   link.CreatedAt,
		})
	}

	s.logger.Debug("user links lookup completed",
		zap.Int64("user_id", userID),
		zap.Int("links_count", len(shortLinks)),
	)

	return shortLinks, nil
}

func normalizeURL(rawURL string) (string, error) {
	value := strings.TrimSpace(rawURL)
	if value == "" || len(value) > maxURLLength {
		return "", fmt.Errorf("validate url length: %w", ErrInvalidURL)
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", ErrInvalidURL)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("validate url scheme: %w", ErrInvalidURL)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("validate url host: %w", ErrInvalidURL)
	}

	return parsed.String(), nil
}

func generateBase62Code(length int) (string, error) {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		result[i] = alphabet[n.Int64()]
	}

	return string(result), nil
}

func normalizeCode(code string) (string, error) {
	value := strings.TrimSpace(code)
	if value == "" || len(value) > maxCodeLength {
		return "", fmt.Errorf("validate code length: %w", ErrInvalidCode)
	}

	if !codePattern.MatchString(value) {
		return "", fmt.Errorf("validate code charset: %w", ErrInvalidCode)
	}

	return value, nil
}

func normalizeVisitMeta(meta VisitMeta, logger *zap.Logger) (string, *string) {
	ip := strings.TrimSpace(meta.IP)
	if ip != "" && net.ParseIP(ip) == nil {
		logger.Warn("visit meta rejected: invalid ip")
		ip = ""
	}

	userAgent := strings.TrimSpace(meta.UserAgent)
	if userAgent == "" {
		return ip, nil
	}

	if len(userAgent) > maxUserAgentLength {
		logger.Warn("visit meta truncated: user agent too long",
			zap.Int("original_length", len(userAgent)),
			zap.Int("max_length", maxUserAgentLength),
		)
		userAgent = userAgent[:maxUserAgentLength]
	}

	return ip, &userAgent
}

func normalizeVisitLimit(limit int) (int, error) {
	if limit <= 0 {
		return defaultVisitLimit, nil
	}

	if limit > MaxVisitLimit {
		return 0, ErrInvalidLimit
	}

	return limit, nil
}
