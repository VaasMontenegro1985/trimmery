package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

const (
	defaultQRSize = 256
	minQRSize     = 128
	maxQRSize     = 1024
)

var ErrInvalidSize = errors.New("invalid qr size")

func (s *LinkService) GenerateLinkQR(ctx context.Context, userID int64, code string, size int) ([]byte, error) {
	normalizedCode, err := normalizeCode(code)
	if err != nil {
		return nil, err
	}

	normalizedSize, err := normalizeQRSize(size)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("qr generation started",
		zap.Int64("user_id", userID),
		zap.String("code", normalizedCode),
		zap.Int("size", normalizedSize),
	)

	link, err := s.storage.FindActiveByCode(ctx, normalizedCode)
	if err != nil {
		return nil, fmt.Errorf("find active link for qr: %w", err)
	}

	if link.UserID == nil || *link.UserID != userID {
		s.logger.Warn("qr generation rejected: forbidden",
			zap.Int64("user_id", userID),
			zap.String("code", normalizedCode),
		)
		return nil, domain.ErrLinkForbidden
	}

	shortURL := strings.TrimRight(s.baseURL, "/") + "/" + link.Code

	png, err := qrcode.Encode(shortURL, qrcode.Medium, normalizedSize)
	if err != nil {
		s.logger.Warn("qr encoding failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("code", normalizedCode),
		)
		return nil, fmt.Errorf("encode qr code: %w", err)
	}

	s.logger.Info("qr generated",
		zap.Int64("user_id", userID),
		zap.Int64("link_id", link.ID),
		zap.String("code", link.Code),
		zap.Int("size", normalizedSize),
	)

	return png, nil
}

func normalizeQRSize(size int) (int, error) {
	if size <= 0 {
		return defaultQRSize, nil
	}

	if size < minQRSize || size > maxQRSize {
		return 0, ErrInvalidSize
	}

	return size, nil
}

func buildQRURL(code string) string {
	return "/api/links/" + code + "/qr"
}
