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
	ErrValidation          = errors.New("magicauth: validation error")
	ErrNilRequest          = fmt.Errorf("%w: request is nil", ErrValidation)
	ErrMissingNonce        = errors.New("magicauth: nonce is required")
	ErrMissingUseCase      = errors.New("magicauth: use_case is required")
	ErrMissingSessionKey   = errors.New("magicauth: session_key is required")
	ErrMissingCredential   = errors.New("magicauth: credential is required")
	ErrMissingSessionID    = errors.New("magicauth: session_id is required")
	ErrMissingPhoneNum     = errors.New("magicauth: phone_number is required for VerifyPhoneNumber use case")
	ErrNoPhoneNumbers      = errors.New("magicauth: phone_numbers must contain at least one number")
	ErrTooManyPhoneNumbers = errors.New("magicauth: phone_numbers exceeds the maximum of 10")
	ErrBlankPhoneNumber    = errors.New("magicauth: phone_numbers must not contain blank entries")

	// API errors (match API error codes)
	ErrBadRequest            = errors.New("magicauth: bad request")
	ErrUnauthorized          = errors.New("magicauth: unauthorized")
	ErrForbidden             = errors.New("magicauth: forbidden")
	ErrSessionNotFound       = errors.New("magicauth: session not found")
	ErrCarrierNotEligible    = errors.New("magicauth: carrier not eligible")
	ErrCarrierRejected       = errors.New("magicauth: carrier rejected the authentication request")
	ErrCarrierAppUnavailable = errors.New("magicauth: carrier app unavailable on the phone that scanned the QR code")
	ErrUnsupportedPlatform   = errors.New("magicauth: unsupported platform")
	ErrPhoneNumberMismatch   = errors.New("magicauth: phone number mismatch")
	ErrInvalidCredential     = errors.New("magicauth: invalid credential format")
	ErrBrowserMismatch       = errors.New("magicauth: browser mismatch")
	ErrSessionNotEligible    = errors.New("magicauth: session not eligible")
	ErrSessionExpired        = errors.New("magicauth: session expired")
	ErrRateLimit             = errors.New("magicauth: rate limit exceeded")
	ErrInternalServer        = errors.New("magicauth: internal server error")
	ErrServiceUnavailable    = errors.New("magicauth: service unavailable")

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
	CodeCarrierRejected          ErrorCode = "CARRIER_REJECTED"
	CodeCarrierAppUnavailable    ErrorCode = "CARRIER_APP_UNAVAILABLE"
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

	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support.
	RequestID string

	// SessionKey is the session the failed request was about, once the server
	// had read it from the request. Empty when the request failed before naming
	// a session, on operations that have none (eligibility, prepare), and from
	// servers older than the release that added it.
	SessionKey string

	// TraceID is the distributed trace ID
	TraceID string

	// SpanID is the span ID for tracing
	SpanID string

	// Timestamp is the ISO 8601 timestamp of when the error occurred on the server
	Timestamp string

	// Service is the name of the service that produced this error
	Service string

	// Details contains additional error details flattened to string values.
	//
	// Both the nested "fields" map (validation errors) and any top-level
	// detail keys are preserved. For CARRIER_REJECTED this carries
	// "carrier_error", "carrier_error_description", and (when present)
	// "stage" and "error_code", so callers can inspect why a device-bound
	// (link/desktop) authentication was rejected by the carrier.
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
	CodeCarrierRejected:          ErrCarrierRejected,
	CodeCarrierAppUnavailable:    ErrCarrierAppUnavailable,
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

// ErrVerificationRecord matches every *VerificationRecordError via errors.Is.
var ErrVerificationRecord = errors.New("verification record invalid")

// Values of VerificationRecordError.Code (identical strings in every Glide backend SDK).
const (
	VerificationRecordCodeMalformed          = "MALFORMED"
	VerificationRecordCodeDisallowedHeader   = "DISALLOWED_HEADER"
	VerificationRecordCodeUnknownKey         = "UNKNOWN_KEY"
	VerificationRecordCodeBadSignature       = "BAD_SIGNATURE"
	VerificationRecordCodeWrongIssuer        = "WRONG_ISSUER"
	VerificationRecordCodeWrongAudience      = "WRONG_AUDIENCE"
	VerificationRecordCodeExpired            = "EXPIRED"
	VerificationRecordCodeNotYetValid        = "NOT_YET_VALID"
	VerificationRecordCodeNonceMismatch      = "NONCE_MISMATCH"
	VerificationRecordCodeOriginMismatch     = "ORIGIN_MISMATCH"
	VerificationRecordCodeUnsupportedVersion = "UNSUPPORTED_VERSION"
	VerificationRecordCodeRevokedKey         = "REVOKED_KEY"
	VerificationRecordCodeRecordUnavailable  = "RECORD_UNAVAILABLE"
)

// VerificationRecordError is returned when a verification record fails
// verification. Code is one of the VerificationRecordCode constants.
type VerificationRecordError struct {
	Code    string
	Message string
}

func (e *VerificationRecordError) Error() string {
	return fmt.Sprintf("verification record: %s: %s", e.Code, e.Message)
}

// Is reports whether target is ErrVerificationRecord.
func (e *VerificationRecordError) Is(target error) bool {
	return target == ErrVerificationRecord
}

func recordError(code, message string) *VerificationRecordError {
	return &VerificationRecordError{Code: code, Message: message}
}

// newAPIError creates an APIError from an error response.
func newAPIError(resp *errorResponse, status int, retryAfter int) *APIError {
	code := ErrorCode(resp.Code)
	details := flattenErrorDetails(resp.Details)

	return &APIError{
		Err:        codeToSentinel(code),
		Code:       code,
		Message:    resp.Message,
		Status:     status,
		RequestID:  resp.RequestID,
		SessionKey: resp.SessionKey,
		TraceID:    resp.TraceID,
		SpanID:     resp.SpanID,
		Timestamp:  resp.Timestamp,
		Service:    resp.Service,
		Details:    details,
		RetryAfter: retryAfter,
	}
}

// flattenErrorDetails converts the raw API "details" object into a flat
// map[string]string. The nested "fields" map (validation errors) is merged in
// alongside any top-level scalar detail keys, so callers never silently lose
// detail keys the server sent. This is what preserves the CARRIER_REJECTED
// carrier_error/carrier_error_description/stage/error_code keys that the
// previous fields-only parsing dropped. Composite top-level values (arrays and
// non-"fields" objects) are skipped because they have no string representation.
func flattenErrorDetails(raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for key, value := range raw {
		switch val := value.(type) {
		case string:
			out[key] = val
		case bool, float64:
			out[key] = fmt.Sprintf("%v", val)
		case map[string]any:
			if key == "fields" {
				for fieldKey, fieldValue := range val {
					if s, ok := fieldValue.(string); ok {
						out[fieldKey] = s
					} else if fieldValue != nil {
						out[fieldKey] = fmt.Sprintf("%v", fieldValue)
					}
				}
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
