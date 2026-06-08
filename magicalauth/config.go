// Package magicalauth provides a Go SDK for the Magic Auth Aggregator API.
//
// The SDK supports phone number verification and retrieval using digital credentials
// through carrier-based authentication (TS43 protocol and Link strategy).
//
// # Quick Start
//
//	client := magicalauth.NewClient(magicalauth.Config{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	})
//
//	resp, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
//	    Nonce:   "random-nonce",
//	    UseCase: magicalauth.UseCaseVerifyPhoneNumber,
//	})
package magicalauth

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the production API endpoint.
	DefaultBaseURL = "https://api.glideidentity.app"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// Version is the SDK version, aligned with API version.
	Version = "3.0.0"
)

// HTTPClient is the interface for making HTTP requests.
// Implement this interface to provide a custom HTTP client (BYOC).
// Defaults to http.DefaultClient if not provided.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Logger is the interface for logging.
// Implement this interface to provide a custom logger (BYOL).
// Defaults to noopLogger if not provided.
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// noopLogger is the default logger that discards all log messages.
type noopLogger struct{}

func (noopLogger) Debug(string, ...any) {} // intentionally empty — discards debug messages
func (noopLogger) Info(string, ...any)  {} // intentionally empty — discards info messages
func (noopLogger) Warn(string, ...any)  {} // intentionally empty — discards warn messages
func (noopLogger) Error(string, ...any) {} // intentionally empty — discards error messages

// Config holds all configuration for the MagicalAuth client.
type Config struct {
	// ClientID is the OAuth2 client ID (required).
	ClientID string

	// ClientSecret is the OAuth2 client secret (required).
	ClientSecret string

	// BaseURL is the API base URL. Defaults to DefaultBaseURL.
	BaseURL string

	// HTTPClient is the HTTP client to use. Defaults to http.DefaultClient with Timeout.
	HTTPClient HTTPClient

	// Logger is the logger to use. Defaults to a no-op logger.
	Logger Logger

	// Timeout is the HTTP request timeout. Defaults to DefaultTimeout.
	// Only used when HTTPClient is not provided.
	Timeout time.Duration
}

// validate checks required fields and sets defaults.
func (c *Config) validate() error {
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	return nil
}

// applyDefaults sets default values for optional fields.
func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: c.Timeout}
	}
	if c.Logger == nil {
		c.Logger = noopLogger{}
	}
}
