package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSessionExpired     = errors.New("session expired")
	ErrUserInactive       = errors.New("user inactive")
	ErrForbidden          = errors.New("forbidden")
	ErrPasswordTooWeak    = errors.New("password too weak")
	ErrPasswordMismatch   = errors.New("password confirmation mismatch")
	ErrPasswordSame       = errors.New("new password must differ from current password")
	ErrMissingFields      = errors.New("missing required fields")
)
