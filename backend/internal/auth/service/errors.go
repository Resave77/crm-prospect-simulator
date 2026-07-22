package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSessionExpired     = errors.New("session expired")
	ErrUserInactive       = errors.New("user inactive")
	ErrForbidden          = errors.New("forbidden")
)
