package glide

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the client
type Option func(*Config)

// WithAPIKey sets the API key for authentication
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithBaseURL sets a custom base URL for the API
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithRetry sets the retry configuration
func WithRetry(count int, delay time.Duration) Option {
	return func(c *Config) {
		c.RetryCount = count
		c.RetryDelay = delay
	}
}

// WithRateLimit enables rate limiting with the specified rate
func WithRateLimit(rate int, period time.Duration) Option {
	return func(c *Config) {
		c.RateLimitEnabled = true
		c.RateLimitRate = rate
		c.RateLimitPeriod = period
	}
}

// WithNoRateLimit explicitly disables rate limiting
func WithNoRateLimit() Option {
	return func(c *Config) {
		c.RateLimitEnabled = false
	}
}

// WithDebug enables debug logging with debug level
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
		if debug {
			c.LogLevel = LogLevelDebug
		} else {
			c.LogLevel = LogLevelSilent
		}
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(level LogLevel) Option {
	return func(c *Config) {
		c.LogLevel = level
		if level > LogLevelSilent {
			c.Debug = true
		}
	}
}

// WithLogger sets a custom logger implementation
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
		// If a custom logger is provided, assume debug is enabled
		c.Debug = true
	}
}
