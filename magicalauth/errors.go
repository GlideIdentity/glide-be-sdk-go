package magicalauth

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
// Use errors.Is() to check for these errors.
var (
	// Configuration errors
	ErrMissingClientID     = errors.New("magicauth: client_id is required")
	ErrMissingClientSecret = errors.New("magicauth: client_secret is required")

	// Validation errors
	ErrValidation        = errors.New("magicauth: validation error")
	ErrNilRequest        = fmt.Errorf("%w: request is nil", ErrValidation)
	ErrMissingNonce      = errors.New("magicauth: nonce is required")
	ErrMissingUseCase    = errors.New("magicauth: use_case is required")
	ErrMissingSessionKey = errors.New("magicauth: session_key is required")
	ErrMissingCredential = errors.New("magicauth: credential is required")
	ErrMissingSessionID  = errors.New("magicauth: session_id is required")
	ErrMissingPhoneNum   = errors.New("magicauth: phone_number is required for VerifyPhoneNumber use case")

	// API errors (match API error codes)
	ErrBadRequest          = errors.New("magicauth: bad request")
	ErrUnauthorized        = errors.New("magicauth: unauthorized")
	ErrForbidden           = errors.New("magicauth: forbidden")
	ErrSessionNotFound     = errors.New("magicauth: session not found")
	ErrCarrierNotEligible  = errors.New("magicauth: carrier not eligible")
	ErrUnsupportedPlatform = errors.New("magicauth: unsupported platform")
	ErrPhoneNumberMismatch = errors.New("magicauth: phone number mismatch")
	ErrInvalidCredential   = errors.New("magicauth: invalid credential format")
	ErrBrowserMismatch     = errors.New("magicauth: browser mismatch")
	ErrSessionNotEligible  = errors.New("magicauth: session not eligible")
	ErrSessionExpired      = errors.New("magicauth: session expired")
	ErrRateLimit           = errors.New("magicauth: rate limit exceeded")
	ErrInternalServer      = errors.New("magicauth: internal server error")
	ErrServiceUnavailable  = errors.New("magicauth: service unavailable")

	// Device binding errors (SDK-side validation)
	ErrDeviceBindingFailed  = errors.New("magicauth: device binding validation failed")
	ErrMissingBindingCookie = errors.New("magicauth: binding cookie missing or invalid")
)

// ErrorCode represents API error codes.
type ErrorCode string

const (
	CodeBadRequest               ErrorCode = "BAD_REQUEST"
	CodeUnauthorized             ErrorCode = "UNAUTHORIZED"
	CodeForbidden                ErrorCode = "FORBIDDEN"
	CodeValidationError          ErrorCode = "VALIDATION_ERROR"
	CodeMissingParameters        ErrorCode = "MISSING_PARAMETERS"
	CodeSessionNotFound          ErrorCode = "SESSION_NOT_FOUND"
	CodeInvalidVerification      ErrorCode = "INVALID_VERIFICATION"
	CodeCarrierNotEligible       ErrorCode = "CARRIER_NOT_ELIGIBLE"
	CodeUnsupportedPlatform      ErrorCode = "UNSUPPORTED_PLATFORM"
	CodePhoneNumberMismatch      ErrorCode = "PHONE_NUMBER_MISMATCH"
	CodeInvalidCredential        ErrorCode = "INVALID_CREDENTIAL_FORMAT" // #nosec G101 -- error code, not a credential
	CodeAudienceValidationFailed ErrorCode = "AUDIENCE_VALIDATION_FAILED"
	CodeBrowserMismatch          ErrorCode = "BROWSER_MISMATCH"
	CodeQRAlreadyScanned         ErrorCode = "QR_ALREADY_SCANNED"
	CodeSessionNotEligible       ErrorCode = "SESSION_NOT_ELIGIBLE"
	CodeSessionExpired           ErrorCode = "SESSION_EXPIRED"
	CodeRateLimitExceeded        ErrorCode = "RATE_LIMIT_EXCEEDED"
	CodeInternalServerError      ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable       ErrorCode = "SERVICE_UNAVAILABLE"

	// SDK-side error codes for device binding
	CodeDeviceBindingFailed  ErrorCode = "DEVICE_BINDING_FAILED"
	CodeMissingBindingCookie ErrorCode = "MISSING_BINDING_COOKIE"
)

// APIError represents an error response from the Magic Auth API.
// Use errors.As() to extract detailed error information.
type APIError struct {
	// Err is the sentinel error for use with errors.Is()
	Err error

	// Code is the API error code
	Code ErrorCode

	// Message is the human-readable error message
	Message string

	// Status is the HTTP status code
	Status int

	// RequestID is the request tracking ID
	RequestID string

	// TraceID is the distributed trace ID
	TraceID string

	// SpanID is the span ID for tracing
	SpanID string

	// Timestamp is the ISO 8601 timestamp of when the error occurred on the server
	Timestamp string

	// Service is the name of the service that produced this error
	Service string

	// Details contains additional error details
	Details map[string]string

	// RetryAfter is the number of seconds to wait before retrying (for rate limits)
	RetryAfter int
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("magicauth: %s: %s (request_id: %s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("magicauth: %s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying sentinel error for use with errors.Is().
func (e *APIError) Unwrap() error {
	return e.Err
}

// errorCodeMap maps API error codes to sentinel errors for O(1) lookup.
var errorCodeMap = map[ErrorCode]error{
	CodeBadRequest:               ErrBadRequest,
	CodeMissingParameters:        ErrBadRequest,
	CodeUnauthorized:             ErrUnauthorized,
	CodeForbidden:                ErrForbidden,
	CodeValidationError:          ErrValidation,
	CodeSessionNotFound:          ErrSessionNotFound,
	CodeInvalidVerification:      ErrBadRequest,
	CodeCarrierNotEligible:       ErrCarrierNotEligible,
	CodeUnsupportedPlatform:      ErrUnsupportedPlatform,
	CodePhoneNumberMismatch:      ErrPhoneNumberMismatch,
	CodeInvalidCredential:        ErrInvalidCredential,
	CodeAudienceValidationFailed: ErrValidation,
	CodeBrowserMismatch:          ErrBrowserMismatch,
	CodeQRAlreadyScanned:         ErrBadRequest,
	CodeSessionNotEligible:       ErrSessionNotEligible,
	CodeSessionExpired:           ErrSessionExpired,
	CodeDeviceBindingFailed:      ErrDeviceBindingFailed,
	CodeMissingBindingCookie:     ErrMissingBindingCookie,
	CodeRateLimitExceeded:        ErrRateLimit,
	CodeInternalServerError:      ErrInternalServer,
	CodeServiceUnavailable:       ErrServiceUnavailable,
}

// codeToSentinel maps API error codes to sentinel errors.
func codeToSentinel(code ErrorCode) error {
	if sentinel, ok := errorCodeMap[code]; ok {
		return sentinel
	}
	return ErrInternalServer
}

// DesktopWaitError is returned when the server responds with HTTP 202,
// indicating the desktop QR session is still waiting for mobile authentication.
// The caller should inspect Retry and SessionExpiresInSeconds to decide whether
// to re-issue the request.
type DesktopWaitError struct {
	Response *DesktopWaitResponse
}

func (e *DesktopWaitError) Error() string {
	return fmt.Sprintf("magicauth: desktop wait: %s (retry=%v, expires_in=%ds)",
		e.Response.Message, e.Response.Retry, e.Response.SessionExpiresInSeconds)
}

// newAPIError creates an APIError from an error response.
func newAPIError(resp *errorResponse, status int, retryAfter int) *APIError {
	code := ErrorCode(resp.Code)
	var details map[string]string
	if resp.Details != nil {
		details = resp.Details.Fields
	}

	return &APIError{
		Err:        codeToSentinel(code),
		Code:       code,
		Message:    resp.Message,
		Status:     status,
		RequestID:  resp.RequestID,
		TraceID:    resp.TraceID,
		SpanID:     resp.SpanID,
		Timestamp:  resp.Timestamp,
		Service:    resp.Service,
		Details:    details,
		RetryAfter: retryAfter,
	}
}

// newOAuthError creates an APIError from an OAuth error response.
func newOAuthError(resp *oauthError, status int) *APIError {
	var sentinel error
	var code ErrorCode

	switch resp.Error {
	case "invalid_client":
		sentinel = ErrUnauthorized
		code = CodeUnauthorized
	default:
		sentinel = ErrBadRequest
		code = CodeBadRequest
	}

	message := resp.ErrorDescription
	if message == "" {
		message = resp.Error
	}

	return &APIError{
		Err:     sentinel,
		Code:    code,
		Message: message,
		Status:  status,
	}
}
