package glide

import (
	"fmt"
)

// Error codes - matching the server's error codes but sanitized for client use
const (
	// 400 errors
	ErrCodeBadRequest        = "BAD_REQUEST"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeInvalidParameters = "INVALID_PARAMETERS"

	// 401 errors
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"

	// 403 errors
	ErrCodeForbidden               = "FORBIDDEN"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// 404 errors
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeSessionNotFound  = "SESSION_NOT_FOUND"

	// 409 errors
	ErrCodeConflict              = "CONFLICT"
	ErrCodeResourceAlreadyExists = "RESOURCE_ALREADY_EXISTS"

	// 422 errors
	ErrCodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
	ErrCodeCarrierNotEligible  = "CARRIER_NOT_ELIGIBLE"
	ErrCodeVerificationFailed  = "VERIFICATION_FAILED"
	ErrCodeSessionExpired      = "SESSION_EXPIRED"

	// 429 errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   = "TOO_MANY_REQUESTS"

	// 500 errors
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeUnexpectedError     = "UNEXPECTED_ERROR"

	// 503 errors
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeProviderError      = "PROVIDER_ERROR"
)

// Error represents an error returned by the Glide API
type Error struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Status    int                    `json:"status,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s: %s (request_id: %s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsCode checks if the error matches a specific error code
func (e *Error) IsCode(code string) bool {
	return e.Code == code
}

// NewError creates a new Error with the given code and message
func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithStatus creates a new Error with status code
func NewErrorWithStatus(code, message string, status int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// IsRetryable returns true if the error is retryable
func (e *Error) IsRetryable() bool {
	switch e.Code {
	case ErrCodeRateLimitExceeded,
		ErrCodeTooManyRequests,
		ErrCodeServiceUnavailable,
		ErrCodeProviderError:
		return true
	default:
		return e.Status >= 500 && e.Status < 600
	}
}

// sanitizeError removes sensitive information from server errors
func sanitizeError(serverErr *Error) *Error {
	// Create a clean error without sensitive details
	sanitized := &Error{
		Code:      serverErr.Code,
		Message:   getPublicMessage(serverErr.Code),
		Status:    serverErr.Status,
		RequestID: serverErr.RequestID,
	}

	// Only include safe details
	if serverErr.Details != nil {
		sanitized.Details = make(map[string]interface{})
		// Add only non-sensitive details
		for k, v := range serverErr.Details {
			if isSafeDetail(k) {
				sanitized.Details[k] = v
			}
		}
	}

	return sanitized
}

// getPublicMessage returns a user-safe message for the error code
func getPublicMessage(code string) string {
	messages := map[string]string{
		ErrCodeCarrierNotEligible: "This verification method is not available for your device",
		ErrCodeSessionExpired:     "Session has expired, please try again",
		ErrCodeRateLimitExceeded:  "Too many requests, please wait before trying again",
		ErrCodeVerificationFailed: "Verification could not be completed",
		ErrCodeServiceUnavailable: "Service temporarily unavailable",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "An error occurred processing your request"
}

// isSafeDetail checks if a detail field is safe to expose to clients
func isSafeDetail(key string) bool {
	// List of safe detail keys that don't expose sensitive info
	safeKeys := map[string]bool{
		"retry_after": true,
		"session_id":  true,
		"use_case":    true,
	}
	return safeKeys[key]
}
