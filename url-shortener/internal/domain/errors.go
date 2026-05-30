package domain

import "errors"

var (
	ErrLinkNotFound       = errors.New("link not found")
	ErrCodeAlreadyExists  = errors.New("link code already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLinkForbidden      = errors.New("link forbidden")
	ErrLinkDeleted        = errors.New("link deleted")
)
