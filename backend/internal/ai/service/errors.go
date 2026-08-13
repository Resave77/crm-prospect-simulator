package service

import "errors"

var (
	ErrAINotConfigured   = errors.New("AI_NOT_CONFIGURED")
	ErrAIUnavailable     = errors.New("AI_UNAVAILABLE")
	ErrAITimeout         = errors.New("AI_TIMEOUT")
	ErrAIRateLimited     = errors.New("AI_RATE_LIMITED")
	ErrAIInvalidResponse = errors.New("AI_INVALID_RESPONSE")
	ErrAIAuthentication  = errors.New("AI_AUTHENTICATION_FAILED")
	ErrAIRequestRejected = errors.New("AI_REQUEST_REJECTED")
)
