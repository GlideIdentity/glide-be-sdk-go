package glide

import (
	"fmt"
)

// Error codes.
const (
	ErrCodeConfigurationError   = "CONFIGURATION_ERROR"
	ErrCodeInvalidPhoneNumber   = "INVALID_PHONE_NUMBER"
	ErrCodeMissingRequiredField = "MISSING_REQUIRED_FIELD"
	ErrCodeInvalidUseCase       = "INVALID_USE_CASE"
	ErrCodeInvalidSession       = "INVALID_SESSION"
	ErrCodeSessionExpired       = "SESSION_EXPIRED"
	ErrCodeCarrierNotEligible   = "CARRIER_NOT_ELIGIBLE"
	ErrCodeUnsupportedPlatform  = "UNSUPPORTED_PLATFORM"
	ErrCodePhoneNumberMismatch  = "PHONE_NUMBER_MISMATCH"
	ErrCodeInvalidCredential    = "INVALID_CREDENTIAL"
	ErrCodeVerificationFailed   = "VERIFICATION_FAILED"
	ErrCodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalServerError  = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
)

// MagicalAuthError represents an API error.
type MagicalAuthError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Status    int                    `json:"-"`
}

// Error implements the error interface.
func (e *MagicalAuthError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s: %s (request_id: %s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsCode checks if the error matches a specific code.
func (e *MagicalAuthError) IsCode(code string) bool {
	return e.Code == code
}

// NewMagicalAuthError creates a new error.
func NewMagicalAuthError(code, message string) *MagicalAuthError {
	return &MagicalAuthError{
		Code:    code,
		Message: message,
	}
}

// NewMagicalAuthErrorWithStatus creates a new error with HTTP status.
func NewMagicalAuthErrorWithStatus(code, message string, status int) *MagicalAuthError {
	return &MagicalAuthError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}
