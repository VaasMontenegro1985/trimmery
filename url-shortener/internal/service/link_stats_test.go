package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"url-shortener/internal/domain"
)

func TestGetLinkStatsForbiddenForAnonymousLink(t *testing.T) {
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
			UserID:      nil,
			ClicksCount: 1,
		},
	}

	linkService := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	_, err := linkService.GetLinkStats(context.Background(), 42, "abc123", 0)
	if !errors.Is(err, domain.ErrLinkForbidden) {
		t.Fatalf("GetLinkStats() error = %v, want ErrLinkForbidden", err)
	}
}

func TestGetLinkStatsForbiddenForOtherUser(t *testing.T) {
	otherUserID := int64(99)
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
			UserID:      &otherUserID,
			ClicksCount: 1,
		},
	}

	linkService := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	_, err := linkService.GetLinkStats(context.Background(), 42, "abc123", 0)
	if !errors.Is(err, domain.ErrLinkForbidden) {
		t.Fatalf("GetLinkStats() error = %v, want ErrLinkForbidden", err)
	}
}

func TestRedirectByCodeRecordsVisit(t *testing.T) {
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
			ClicksCount: 0,
		},
	}

	linkService := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	link, err := linkService.RedirectByCode(context.Background(), "abc123", VisitMeta{
		IP:        "203.0.113.1",
		UserAgent: "TestAgent/1.0",
	})
	if err != nil {
		t.Fatalf("RedirectByCode() error = %v", err)
	}

	if !storage.recordVisitCalled {
		t.Fatal("RecordVisit was not called")
	}

	if link.ClicksCount != 1 {
		t.Fatalf("link.ClicksCount = %d, want 1", link.ClicksCount)
	}
}

func TestRedirectByCodeStillReturnsLinkWhenRecordVisitFails(t *testing.T) {
	storage := &fakeLinkStorage{
		link: domain.Link{
			ID:          1,
			Code:        "abc123",
			OriginalURL: "https://example.com",
		},
		recordVisitErr: errors.New("db down"),
	}

	linkService := NewLinkService(storage, "http://localhost:8080", zap.NewNop())

	link, err := linkService.RedirectByCode(context.Background(), "abc123", VisitMeta{})
	if err != nil {
		t.Fatalf("RedirectByCode() error = %v", err)
	}

	if link.Code != "abc123" {
		t.Fatalf("link.Code = %q, want abc123", link.Code)
	}
}

func TestNormalizeVisitLimit(t *testing.T) {
	got, err := normalizeVisitLimit(0)
	if err != nil || got != defaultVisitLimit {
		t.Fatalf("normalizeVisitLimit(0) = (%d, %v), want (%d, nil)", got, err, defaultVisitLimit)
	}

	_, err = normalizeVisitLimit(101)
	if !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("normalizeVisitLimit(101) error = %v, want ErrInvalidLimit", err)
	}
}

type fakeLinkStorage struct {
	link              domain.Link
	visits            []domain.Visit
	recordVisitCalled bool
	recordVisitErr    error
}

func (s *fakeLinkStorage) Create(_ context.Context, link domain.Link) (domain.Link, error) {
	return link, nil
}

func (s *fakeLinkStorage) FindByCode(_ context.Context, _ string) (domain.Link, error) {
	return s.link, nil
}

func (s *fakeLinkStorage) FindActiveByCode(_ context.Context, _ string) (domain.Link, error) {
	return s.link, nil
}

func (s *fakeLinkStorage) ListByUserID(_ context.Context, _ int64) ([]domain.Link, error) {
	return nil, nil
}

func (s *fakeLinkStorage) UpdateLink(_ context.Context, _ int64, _ string, _ *string, _ *string) (domain.Link, error) {
	return s.link, nil
}

func (s *fakeLinkStorage) SoftDeleteLink(_ context.Context, _ int64, _ string) error {
	return nil
}

func (s *fakeLinkStorage) RecordVisit(_ context.Context, linkID int64, _ string, _ *string) error {
	s.recordVisitCalled = true
	if s.recordVisitErr != nil {
		return s.recordVisitErr
	}

	s.link.ClicksCount++
	return nil
}

func (s *fakeLinkStorage) ListVisitsByLinkID(_ context.Context, _ int64, _ int) ([]domain.Visit, error) {
	if len(s.visits) > 0 {
		return s.visits, nil
	}

	return []domain.Visit{{
		VisitedAt: time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC),
	}}, nil
}
