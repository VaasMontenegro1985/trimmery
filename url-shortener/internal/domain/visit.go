package domain

import "time"

type Visit struct {
	ID        int64
	LinkID    int64
	VisitedAt time.Time
	IP        *string
	UserAgent *string
}
