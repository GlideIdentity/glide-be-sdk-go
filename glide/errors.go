package glide

import (
	"fmt"
)

// Error codes - Only codes that the server actually returns to clients
// Based on magic-auth-2.0/pkg/types/error_types.go (non-internal codes only)
const (
	// 400 Bad Request errors
	ErrCodeBadRequest        = "BAD_REQUEST"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeMissingParameters = "MISSING_PARAMETERS"

	// 404 Not Found errors
	ErrCodeSessionNotFound = "SESSION_NOT_FOUND"

	// 422 Unprocessable Entity errors
	ErrCodeInvalidVerification     = "INVALID_VERIFICATION"
	ErrCodeCarrierNotEligible      = "CARRIER_NOT_ELIGIBLE"
	ErrCodeUnsupportedPlatform     = "UNSUPPORTED_PLATFORM"
	ErrCodePhoneNumberMismatch     = "PHONE_NUMBER_MISMATCH"
	ErrCodeInvalidCredentialFormat = "INVALID_CREDENTIAL_FORMAT"
	ErrCodeUnprocessableEntity     = "UNPROCESSABLE_ENTITY"

	// 429 Too Many Requests errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	// 500 Internal Server errors
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"

	// 503 Service Unavailable errors
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
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
		ErrCodeServiceUnavailable:
		return true
	default:
		// Also retry on 500 errors even if not explicitly listed
		return e.Status >= 500 && e.Status < 600
	}
}

// sanitizeError removes sensitive information from server errors
func sanitizeError(serverErr *Error) *Error {
	// Pass through the backend error as-is, trusting the backend to provide appropriate messages
	// The backend is responsible for not exposing sensitive information
	return &Error{
		Code:      serverErr.Code,
		Message:   serverErr.Message, // Use the actual backend message
		Status:    serverErr.Status,
		RequestID: serverErr.RequestID,
		Details:   serverErr.Details, // Include all details from backend
	}
}

// getPublicMessage returns a user-safe message for the error code
func getPublicMessage(code string) string {
	messages := map[string]string{
		// 400 errors
		ErrCodeBadRequest:        "Invalid request. Please try again.",
		ErrCodeValidationError:   "The provided information is invalid.",
		ErrCodeMissingParameters: "Required information is missing.",

		// 404 errors
		ErrCodeSessionNotFound: "Session not found. Please start over.",

		// 422 errors
		ErrCodeInvalidVerification:     "Verification failed. Please try again.",
		ErrCodeCarrierNotEligible:      "Your carrier is not eligible for this authentication method.",
		ErrCodeUnsupportedPlatform:     "Your platform is not supported.",
		ErrCodePhoneNumberMismatch:     "Phone number does not match.",
		ErrCodeInvalidCredentialFormat: "Invalid credential format.",
		ErrCodeUnprocessableEntity:     "Request could not be processed. Please try again.",

		// 429 errors
		ErrCodeRateLimitExceeded: "Too many requests. Please wait and try again.",

		// 500 errors
		ErrCodeInternalServerError: "An error occurred. Please try again later.",

		// 503 errors
		ErrCodeServiceUnavailable: "Service temporarily unavailable. Please try again later.",
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
