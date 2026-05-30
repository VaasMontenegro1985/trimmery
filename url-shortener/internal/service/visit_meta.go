package service

const (
	MaxVisitLimit      = 100
	maxUserAgentLength = 1024
	defaultVisitLimit  = 50
)

type VisitMeta struct {
	IP        string
	UserAgent string
}
