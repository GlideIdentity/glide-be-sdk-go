package glide

import (
	"context"
	"net/http"
	"os"
	"time"
)

// Client is the Glide SDK client.
type Client struct {
	MagicalAuth MagicalAuthService

	config     *Config
	httpClient *http.Client
	logger     Logger
	oauth      *oauthManager
}

// Config holds client configuration.
type Config struct {
	// OAuth2 Client Credentials authentication
	ClientID     string
	ClientSecret string

	// Common settings
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	HTTPClient *http.Client
	Logger     Logger

	// OAuth2 token refresh buffer (seconds before expiry to refresh)
	TokenRefreshBuffer int
}

// Option configures the client.
type Option func(*Config)

// WithClientCredentials sets OAuth2 client credentials.
func WithClientCredentials(clientID, clientSecret string) Option {
	return func(c *Config) {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Config) { c.BaseURL = url }
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) { c.Timeout = timeout }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.HTTPClient = client }
}

// WithRetry sets retry count and delay.
func WithRetry(count int, delay time.Duration) Option {
	return func(c *Config) {
		c.RetryCount = count
		c.RetryDelay = delay
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger Logger) Option {
	return func(c *Config) { c.Logger = logger }
}

// WithLogLevel sets the log level.
func WithLogLevel(level LogLevel) Option {
	return func(c *Config) {
		c.Logger = NewDefaultLogger(level)
	}
}

// WithDebug enables debug logging.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		if debug {
			c.Logger = NewDefaultLogger(LogLevelDebug)
		}
	}
}

// WithTokenRefreshBuffer sets how many seconds before token expiry to refresh.
// Default is 60 seconds.
func WithTokenRefreshBuffer(seconds int) Option {
	return func(c *Config) {
		c.TokenRefreshBuffer = seconds
	}
}

// New creates a new Glide client.
func New(opts ...Option) *Client {
	cfg := &Config{
		BaseURL:            "https://api.glideidentity.app",
		Timeout:            30 * time.Second,
		RetryCount:         3,
		RetryDelay:         time.Second,
		TokenRefreshBuffer: 60,
	}

	// Check environment variables for credentials
	if envClientID := os.Getenv("GLIDE_CLIENT_ID"); envClientID != "" {
		cfg.ClientID = envClientID
	}
	if envClientSecret := os.Getenv("GLIDE_CLIENT_SECRET"); envClientSecret != "" {
		cfg.ClientSecret = envClientSecret
	}
	if envBaseURL := os.Getenv("GLIDE_API_BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if envLevel := os.Getenv("GLIDE_LOG_LEVEL"); envLevel != "" {
		cfg.Logger = NewDefaultLogger(ParseLogLevel(envLevel))
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}

	var logger Logger
	if cfg.Logger != nil {
		logger = cfg.Logger
	} else {
		logger = NewNoopLogger()
	}

	client := &Client{
		config:     cfg,
		httpClient: cfg.HTTPClient,
		logger:     logger,
	}

	// Initialize OAuth manager if client credentials are provided
	if cfg.ClientID != "" && cfg.ClientSecret != "" {
		client.oauth = newOAuthManager(cfg.BaseURL, cfg.ClientID, cfg.ClientSecret, cfg.HTTPClient, cfg.TokenRefreshBuffer, logger)
		logger.Info("Glide SDK initialized",
			Field{"version", GetVersion()},
			Field{"baseURL", cfg.BaseURL},
		)
	} else {
		logger.Warn("Glide SDK initialized without credentials",
			Field{"version", GetVersion()},
			Field{"baseURL", cfg.BaseURL},
		)
	}

	client.MagicalAuth = newMagicalAuthService(client)

	return client
}

// Context returns a context with the client's timeout.
func (c *Client) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.config.Timeout)
}

func (c *Client) validateConfig() error {
	// Check for OAuth2 credentials
	if c.config.ClientID == "" || c.config.ClientSecret == "" {
		return NewMagicalAuthError(ErrCodeConfigurationError,
			"OAuth2 credentials required. Use WithClientCredentials(clientID, clientSecret) or set GLIDE_CLIENT_ID and GLIDE_CLIENT_SECRET environment variables")
	}

	if c.config.BaseURL == "" {
		return NewMagicalAuthError(ErrCodeConfigurationError,
			"Base URL cannot be empty. Use WithBaseURL() option to set a valid URL")
	}

	return nil
}
