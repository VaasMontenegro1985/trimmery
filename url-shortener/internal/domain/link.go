package domain

import "time"

type Link struct {
	ID          int64
	Code        string
	OriginalURL string
	UserID      *int64
	ClicksCount int64
	DeletedAt   *time.Time
	CreatedAt   time.Time
}
