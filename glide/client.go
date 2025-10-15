package glide

import (
	"context"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

// Client is the main Glide SDK client
type Client struct {
	// Services
	MagicAuth    MagicAuthService
	SimSwap      SimSwapService
	NumberVerify NumberVerifyService
	KYC          KYCService

	// Internal
	config      *Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      Logger
}

// Config holds the client configuration
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration

	// Optional rate limiting
	RateLimitEnabled bool
	RateLimitRate    int
	RateLimitPeriod  time.Duration

	// HTTP client (optional)
	HTTPClient *http.Client

	// Debug logging
	Debug     bool      // Enable debug logging
	LogLevel  LogLevel  // Log level (default: LogLevelSilent)
	LogFormat LogFormat // Log output format (default: LogFormatPretty)
	Logger    Logger    // Custom logger implementation (optional)
}

// New creates a new Glide client with the given options
func New(opts ...Option) *Client {
	cfg := &Config{
		BaseURL:    "https://api.glideidentity.app",
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: time.Second,
		LogLevel:   LogLevelSilent,  // Default to no logging
		LogFormat:  LogFormatPretty, // Default to pretty format
	}

	// Check environment variables for debug mode
	if envDebug := os.Getenv("GLIDE_DEBUG"); envDebug != "" {
		if envDebug == "true" || envDebug == "1" {
			cfg.Debug = true
			cfg.LogLevel = LogLevelDebug
		}
	}

	// Check for log level environment variable
	if envLogLevel := os.Getenv("GLIDE_LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = ParseLogLevel(envLogLevel)
		if cfg.LogLevel > LogLevelSilent {
			cfg.Debug = true
		}
	}

	// Check for log format environment variable
	if envLogFormat := os.Getenv("GLIDE_LOG_FORMAT"); envLogFormat != "" {
		cfg.LogFormat = ParseLogFormat(envLogFormat)
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	// Create HTTP client if not provided
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	client := &Client{
		config:     cfg,
		httpClient: cfg.HTTPClient,
	}

	// Set up logger
	if cfg.Logger != nil {
		// Use custom logger if provided
		client.logger = cfg.Logger
	} else if cfg.Debug || cfg.LogLevel > LogLevelSilent {
		// Use default logger with specified level and format
		client.logger = NewDefaultLoggerWithFormat(cfg.LogLevel, cfg.LogFormat)
	} else {
		// Use noop logger when logging is disabled
		client.logger = NewNoopLogger()
	}

	// Log initialization (skip if using pretty format to avoid clutter)
	if dl, ok := client.logger.(*defaultLogger); !ok || dl.format != LogFormatPretty {
		client.logger.Info("Glide SDK initialized",
			Field{"version", "1.0.0"},
			Field{"baseURL", cfg.BaseURL},
			Field{"logLevel", cfg.LogLevel.String()},
		)
	}

	// Initialize rate limiter if configured
	if cfg.RateLimitEnabled {
		limit := rate.Every(cfg.RateLimitPeriod / time.Duration(cfg.RateLimitRate))
		client.rateLimiter = rate.NewLimiter(limit, cfg.RateLimitRate)
		client.logger.Debug("Rate limiting enabled",
			Field{"rate", cfg.RateLimitRate},
			Field{"period", cfg.RateLimitPeriod.String()},
		)
	}

	// Initialize services
	client.MagicAuth = newMagicAuthService(client)
	client.SimSwap = newSimSwapService(client)
	client.NumberVerify = newNumberVerifyService(client)
	client.KYC = newKYCService(client)

	return client
}

// Context returns a context with the client's timeout
func (c *Client) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.config.Timeout)
}
