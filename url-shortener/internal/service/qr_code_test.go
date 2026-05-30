package service

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

func TestGenerateLinkQRForbiddenForOtherUser(t *testing.T) {
	otherUserID := int64(99)
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
			UserID:      &otherUserID,
		},
	}

	svc := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	_, err := svc.GenerateLinkQR(context.Background(), 42, "abc123", 256)
	if !errors.Is(err, domain.ErrLinkForbidden) {
		t.Fatalf("GenerateLinkQR() error = %v, want ErrLinkForbidden", err)
	}
}

func TestGenerateLinkQRReturnsPNG(t *testing.T) {
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
			UserID:      int64Ptr(42),
		},
	}

	svc := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	png, err := svc.GenerateLinkQR(context.Background(), 42, "abc123", 256)
	if err != nil {
		t.Fatalf("GenerateLinkQR() error = %v", err)
	}

	if len(png) < 8 || png[0] != 0x89 || png[1] != 0x50 || png[2] != 0x4E || png[3] != 0x47 {
		t.Fatalf("GenerateLinkQR() invalid PNG header, len=%d", len(png))
	}
}

func TestNormalizeQRSize(t *testing.T) {
	got, err := normalizeQRSize(0)
	if err != nil || got != defaultQRSize {
		t.Fatalf("normalizeQRSize(0) = (%d, %v), want (%d, nil)", got, err, defaultQRSize)
	}
}

func TestNormalizeQRSizeOutOfRange(t *testing.T) {
	_, err := normalizeQRSize(64)
	if !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("normalizeQRSize(64) error = %v, want ErrInvalidSize", err)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
