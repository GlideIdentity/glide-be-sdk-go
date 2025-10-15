// Package glide provides a Go SDK for the Glide Identity API.
//
// The SDK provides easy-to-use methods for phone authentication via SIM cards,
// including carrier verification and phone number retrieval.
//
// Basic usage:
//
//	import "github.com/glideidentity/glide-go-sdk"
//
//	// Create client
//	client := glide.New(
//	    glide.WithAPIKey("your-api-key"),
//	)
//
//	// Prepare authentication
//	resp, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
//	    UseCase: glide.UseCaseVerifyPhoneNumber,
//	    PhoneNumber: "+1234567890",
//	})
package glide

// Re-export everything from the glide package for root-level access
// This allows users to import "github.com/glideidentity/glide-go-sdk"
// instead of "github.com/glideidentity/glide-go-sdk/glide"

import (
	"github.com/glideidentity/glide-go-sdk/glide"
)

// Client and configuration types
type (
	Client = glide.Client
	Config = glide.Config
	Option = glide.Option
)

// Service interfaces
type (
	MagicAuthService    = glide.MagicAuthService
	SimSwapService      = glide.SimSwapService
	NumberVerifyService = glide.NumberVerifyService
	KYCService          = glide.KYCService
)

// MagicAuth types
type (
	PrepareRequest            = glide.PrepareRequest
	PrepareResponse           = glide.PrepareResponse
	VerifyPhoneNumberRequest  = glide.VerifyPhoneNumberRequest
	VerifyPhoneNumberResponse = glide.VerifyPhoneNumberResponse
	GetPhoneNumberRequest     = glide.GetPhoneNumberRequest
	GetPhoneNumberResponse    = glide.GetPhoneNumberResponse
	SessionInfo               = glide.SessionInfo
	PLMN                      = glide.PLMN
	ConsentData               = glide.ConsentData
	ClientInfo                = glide.ClientInfo
)

// SimSwap types
type (
	SimSwapCheckRequest  = glide.SimSwapCheckRequest
	SimSwapCheckResponse = glide.SimSwapCheckResponse
	SimSwapDateRequest   = glide.SimSwapDateRequest
	SimSwapDateResponse  = glide.SimSwapDateResponse
)

// NumberVerify types
type (
	NumberVerifyRequest  = glide.NumberVerifyRequest
	NumberVerifyResponse = glide.NumberVerifyResponse
)

// KYC types
type (
	KYCMatchRequest  = glide.KYCMatchRequest
	KYCMatchResponse = glide.KYCMatchResponse
	MatchResult      = glide.MatchResult
	Address          = glide.Address
)

// Logger types
type (
	Logger    = glide.Logger
	LogLevel  = glide.LogLevel
	LogFormat = glide.LogFormat
	Field     = glide.Field
)

// Error type
type Error = glide.Error

// Constants - Use Cases
const (
	UseCaseGetPhoneNumber    = glide.UseCaseGetPhoneNumber
	UseCaseVerifyPhoneNumber = glide.UseCaseVerifyPhoneNumber
)

// Constants - Log Formats
const (
	LogFormatPretty = glide.LogFormatPretty
	LogFormatJSON   = glide.LogFormatJSON
	LogFormatSimple = glide.LogFormatSimple
)

// Constants - Authentication Strategies
const (
	AuthenticationStrategyTS43 = glide.AuthenticationStrategyTS43
	AuthenticationStrategyLink = glide.AuthenticationStrategyLink
)

// Constants - Log Levels
const (
	LogLevelSilent = glide.LogLevelSilent
	LogLevelError  = glide.LogLevelError
	LogLevelWarn   = glide.LogLevelWarn
	LogLevelInfo   = glide.LogLevelInfo
	LogLevelDebug  = glide.LogLevelDebug
)

// Constants - Error Codes (Only codes the server actually returns)
const (
	// 400 Bad Request errors
	ErrCodeBadRequest        = glide.ErrCodeBadRequest
	ErrCodeValidationError   = glide.ErrCodeValidationError
	ErrCodeMissingParameters = glide.ErrCodeMissingParameters

	// 404 Not Found errors
	ErrCodeSessionNotFound = glide.ErrCodeSessionNotFound

	// 422 Unprocessable Entity errors
	ErrCodeInvalidVerification     = glide.ErrCodeInvalidVerification
	ErrCodeCarrierNotEligible      = glide.ErrCodeCarrierNotEligible
	ErrCodeUnsupportedPlatform     = glide.ErrCodeUnsupportedPlatform
	ErrCodePhoneNumberMismatch     = glide.ErrCodePhoneNumberMismatch
	ErrCodeInvalidCredentialFormat = glide.ErrCodeInvalidCredentialFormat
	ErrCodeUnprocessableEntity     = glide.ErrCodeUnprocessableEntity

	// 429 Too Many Requests errors
	ErrCodeRateLimitExceeded = glide.ErrCodeRateLimitExceeded

	// 500 Internal Server errors
	ErrCodeInternalServerError = glide.ErrCodeInternalServerError

	// 503 Service Unavailable errors
	ErrCodeServiceUnavailable = glide.ErrCodeServiceUnavailable
)

// Functions

// New creates a new Glide client with the given options
var New = glide.New

// Option functions
var (
	WithAPIKey      = glide.WithAPIKey
	WithBaseURL     = glide.WithBaseURL
	WithTimeout     = glide.WithTimeout
	WithHTTPClient  = glide.WithHTTPClient
	WithRetry       = glide.WithRetry
	WithRateLimit   = glide.WithRateLimit
	WithNoRateLimit = glide.WithNoRateLimit
	WithDebug       = glide.WithDebug
	WithLogLevel    = glide.WithLogLevel
	WithLogFormat   = glide.WithLogFormat
	WithLogger      = glide.WithLogger
)

// Error constructors
var (
	NewError           = glide.NewError
	NewErrorWithStatus = glide.NewErrorWithStatus
)

// Validation functions
var (
	ValidatePhoneNumber         = glide.ValidatePhoneNumber
	ValidatePLMN                = glide.ValidatePLMN
	ValidateConsentData         = glide.ValidateConsentData
	ValidateUseCaseRequirements = glide.ValidateUseCaseRequirements
)

// Logger constructors
var (
	NewDefaultLogger = glide.NewDefaultLogger
	NewNoopLogger    = glide.NewNoopLogger
	ParseLogLevel    = glide.ParseLogLevel
)
