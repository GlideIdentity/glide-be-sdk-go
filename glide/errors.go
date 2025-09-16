package glide

import (
	"fmt"
)

// Error codes - matching the server's error codes but sanitized for client use
const (
	// 400 Bad Request errors
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeValidationError    = "VALIDATION_ERROR"
	ErrCodeInvalidParameters  = "INVALID_PARAMETERS"
	ErrCodeMissingParameters  = "MISSING_PARAMETERS"
	ErrCodeInvalidPhoneNumber = "INVALID_PHONE_NUMBER"
	ErrCodeInvalidMCCMNC      = "INVALID_MCC_MNC"

	// 401 Unauthorized errors
	ErrCodeUnauthorized           = "UNAUTHORIZED"
	ErrCodeInvalidCredentials     = "INVALID_CREDENTIALS"
	ErrCodeExpiredToken           = "EXPIRED_TOKEN"
	ErrCodeTokenAcquisitionFailed = "TOKEN_ACQUISITION_FAILED"
	ErrCodeInvalidAPIKey          = "INVALID_API_KEY"
	ErrCodeMissingAuthHeader      = "MISSING_AUTH_HEADER"

	// 403 Forbidden errors
	ErrCodeForbidden               = "FORBIDDEN"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrCodeAccessDenied            = "ACCESS_DENIED"

	// 404 Not Found errors
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeSessionNotFound  = "SESSION_NOT_FOUND"
	ErrCodeCarrierNotFound  = "CARRIER_NOT_FOUND"
	ErrCodeEndpointNotFound = "ENDPOINT_NOT_FOUND"

	// 409 Conflict errors
	ErrCodeConflict               = "CONFLICT"
	ErrCodeResourceAlreadyExists  = "RESOURCE_ALREADY_EXISTS"
	ErrCodeDuplicateSession       = "DUPLICATE_SESSION"
	ErrCodeConcurrentModification = "CONCURRENT_MODIFICATION"

	// 422 Unprocessable Entity errors
	ErrCodeUnprocessableEntity         = "UNPROCESSABLE_ENTITY"
	ErrCodeUnsupportedVerification     = "UNSUPPORTED_VERIFICATION"
	ErrCodeInvalidVerification         = "INVALID_VERIFICATION"
	ErrCodeVerificationFailed          = "VERIFICATION_FAILED"
	ErrCodeOTPExpired                  = "OTP_EXPIRED"
	ErrCodeOTPInvalid                  = "OTP_INVALID"
	ErrCodeRCSUnavailable              = "RCS_UNAVAILABLE"
	ErrCodeCarrierIdentificationFailed = "CARRIER_IDENTIFICATION_FAILED"
	ErrCodeCarrierNotEligible          = "CARRIER_NOT_ELIGIBLE"
	ErrCodeUnsupportedCarrier          = "UNSUPPORTED_CARRIER"
	ErrCodeUnsupportedPlatform         = "UNSUPPORTED_PLATFORM"
	ErrCodeUnsupportedStrategy         = "UNSUPPORTED_STRATEGY"
	ErrCodeInvalidSessionState         = "INVALID_SESSION_STATE"
	ErrCodePhoneNumberMismatch         = "PHONE_NUMBER_MISMATCH"
	ErrCodeInvalidCredentialFormat     = "INVALID_CREDENTIAL_FORMAT"

	// 429 Too Many Requests errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   = "TOO_MANY_REQUESTS"
	ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"

	// 500 Internal Server errors
	ErrCodeInternalServerError              = "INTERNAL_SERVER_ERROR"
	ErrCodeCircuitBreakerConfigurationError = "CIRCUIT_BREAKER_CONFIGURATION_ERROR"
	ErrCodeDatabaseError                    = "DATABASE_ERROR"
	ErrCodeCacheError                       = "CACHE_ERROR"
	ErrCodeSerializationError               = "SERIALIZATION_ERROR"
	ErrCodeCryptoError                      = "CRYPTO_ERROR"

	// 502 Bad Gateway errors
	ErrCodeBadGateway      = "BAD_GATEWAY"
	ErrCodeUpstreamError   = "UPSTREAM_ERROR"
	ErrCodeInvalidResponse = "INVALID_RESPONSE"

	// 503 Service Unavailable errors
	ErrCodeServiceUnavailable     = "SERVICE_UNAVAILABLE"
	ErrCodeDownstreamServiceError = "DOWNSTREAM_SERVICE_ERROR"
	ErrCodeProviderError          = "PROVIDER_ERROR"
	ErrCodeCircuitBreakerOpen     = "CIRCUIT_BREAKER_OPEN"
	ErrCodeMaintenanceMode        = "MAINTENANCE_MODE"

	// 504 Gateway Timeout errors
	ErrCodeGatewayTimeout   = "GATEWAY_TIMEOUT"
	ErrCodeRequestTimeout   = "REQUEST_TIMEOUT"
	ErrCodeUpstreamTimeout  = "UPSTREAM_TIMEOUT"
	ErrCodeDeadlineExceeded = "DEADLINE_EXCEEDED"
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
		ErrCodeQuotaExceeded,
		ErrCodeServiceUnavailable,
		ErrCodeDownstreamServiceError,
		ErrCodeProviderError,
		ErrCodeCircuitBreakerOpen,
		ErrCodeMaintenanceMode,
		ErrCodeGatewayTimeout,
		ErrCodeRequestTimeout,
		ErrCodeUpstreamTimeout,
		ErrCodeDeadlineExceeded,
		ErrCodeBadGateway,
		ErrCodeUpstreamError:
		return true
	default:
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
		ErrCodeBadRequest:         "Invalid request. Please try again.",
		ErrCodeValidationError:    "The provided information is invalid.",
		ErrCodeInvalidParameters:  "Invalid parameters provided.",
		ErrCodeMissingParameters:  "Required information is missing.",
		ErrCodeInvalidPhoneNumber: "Please enter a valid phone number.",
		ErrCodeInvalidMCCMNC:      "Invalid network information.",

		// 401 errors
		ErrCodeUnauthorized:           "Authentication required.",
		ErrCodeInvalidCredentials:     "Invalid credentials provided.",
		ErrCodeExpiredToken:           "Your session has expired. Please authenticate again.",
		ErrCodeTokenAcquisitionFailed: "Unable to acquire authentication token.",
		ErrCodeInvalidAPIKey:          "Invalid API key.",
		ErrCodeMissingAuthHeader:      "Authentication information is missing.",

		// 403 errors
		ErrCodeForbidden:               "Access denied.",
		ErrCodeInsufficientPermissions: "You do not have permission to perform this action.",
		ErrCodeAccessDenied:            "Access denied.",

		// 404 errors
		ErrCodeNotFound:         "Resource not found.",
		ErrCodeResourceNotFound: "The requested resource was not found.",
		ErrCodeSessionNotFound:  "Session not found. Please start over.",
		ErrCodeCarrierNotFound:  "Carrier information not available.",
		ErrCodeEndpointNotFound: "The requested endpoint was not found.",

		// 409 errors
		ErrCodeConflict:               "A conflict occurred. Please try again.",
		ErrCodeResourceAlreadyExists:  "This resource already exists.",
		ErrCodeDuplicateSession:       "A session already exists.",
		ErrCodeConcurrentModification: "The resource was modified. Please try again.",

		// 422 errors
		ErrCodeUnprocessableEntity:         "Unable to process the request.",
		ErrCodeUnsupportedVerification:     "This verification method is not supported.",
		ErrCodeInvalidVerification:         "Verification failed. Please try again.",
		ErrCodeVerificationFailed:          "Verification could not be completed.",
		ErrCodeOTPExpired:                  "The verification code has expired.",
		ErrCodeOTPInvalid:                  "Invalid verification code.",
		ErrCodeRCSUnavailable:              "RCS service is not available.",
		ErrCodeCarrierIdentificationFailed: "Unable to identify carrier.",
		ErrCodeCarrierNotEligible:          "Your carrier is not eligible for this authentication method.",
		ErrCodeUnsupportedCarrier:          "Your carrier is not supported.",
		ErrCodeUnsupportedPlatform:         "Your platform is not supported.",
		ErrCodeUnsupportedStrategy:         "This authentication method is not supported.",
		ErrCodeInvalidSessionState:         "Invalid session state.",
		ErrCodePhoneNumberMismatch:         "Phone number does not match.",
		ErrCodeInvalidCredentialFormat:     "Invalid credential format.",

		// 429 errors
		ErrCodeRateLimitExceeded: "Too many requests. Please wait and try again.",
		ErrCodeTooManyRequests:   "Too many requests. Please wait and try again.",
		ErrCodeQuotaExceeded:     "Request quota exceeded.",

		// 500 errors
		ErrCodeInternalServerError:              "An error occurred. Please try again later.",
		ErrCodeCircuitBreakerConfigurationError: "Service configuration error.",
		ErrCodeDatabaseError:                    "Database error occurred.",
		ErrCodeCacheError:                       "Cache error occurred.",
		ErrCodeSerializationError:               "Data processing error.",
		ErrCodeCryptoError:                      "Encryption error occurred.",

		// 502 errors
		ErrCodeBadGateway:      "Gateway error. Please try again.",
		ErrCodeUpstreamError:   "Upstream service error.",
		ErrCodeInvalidResponse: "Invalid response received.",

		// 503 errors
		ErrCodeServiceUnavailable:     "Service temporarily unavailable. Please try again later.",
		ErrCodeDownstreamServiceError: "External service error.",
		ErrCodeProviderError:          "Provider service error.",
		ErrCodeCircuitBreakerOpen:     "Service temporarily unavailable.",
		ErrCodeMaintenanceMode:        "Service under maintenance.",

		// 504 errors
		ErrCodeGatewayTimeout:   "Request timed out. Please try again.",
		ErrCodeRequestTimeout:   "Request timed out. Please try again.",
		ErrCodeUpstreamTimeout:  "Upstream service timeout.",
		ErrCodeDeadlineExceeded: "Request deadline exceeded.",
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
