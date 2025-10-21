package domain

import (
	"errors"
	"time"
)

var (
	// ErrAlreadyExists is an error for already existing entity
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrNotFound is an error for not found entity
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidArgument is an error for invalid argument
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrUnauthorized is an error for unauthorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is an error for forbidden
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is an error for internal server error
	ErrInternal = errors.New("internal server error")

	// ErrTooManyRequests is an error for too many requests
	ErrTooManyRequests = errors.New("too many requests")
)

// ErrorType — типизация видов доменных ошибок для UI/маппинга статусов
type ErrorType string

const (
	ErrorTypeInternal            ErrorType = "internal"
	ErrorTypeRateLimit           ErrorType = "rate_limit"
	ErrorTypeQuotaExceeded       ErrorType = "quota_exceeded"
	ErrorTypeProviderUnavailable ErrorType = "provider_unavailable"
)

// TypedError — расширяемая обёртка для доменных ошибок с типом/кодом и метаданными
type TypedError struct {
	Err   error
	Type  ErrorType // e.g. rate_limit, quota_exceeded, provider_unavailable, internal
	Code  string // optional provider or http/grpc code
	RetryAfter *time.Duration // optional suggested delay
}

func (e *TypedError) Error() string { return e.Err.Error() }
func (e *TypedError) Unwrap() error { return e.Err }

// Helpers to construct common typed errors
func RateLimitError(base error, retryAfter *time.Duration) *TypedError {
	if base == nil { base = ErrTooManyRequests }
	return &TypedError{Err: base, Type: ErrorTypeRateLimit, Code: "429", RetryAfter: retryAfter}
}

func QuotaExceededError(base error) *TypedError {
	if base == nil { base = ErrTooManyRequests }
	return &TypedError{Err: base, Type: ErrorTypeQuotaExceeded}
}

func ProviderUnavailableError(base error) *TypedError {
	if base == nil { base = ErrInternal }
	return &TypedError{Err: base, Type: ErrorTypeProviderUnavailable, Code: "unavailable"}
}

func InternalTypedError(base error) *TypedError {
	if base == nil { base = ErrInternal }
	return &TypedError{Err: base, Type: ErrorTypeInternal}
}
