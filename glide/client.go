package glide

import (
	"context"
	"net/http"
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
}

// New creates a new Glide client with the given options
func New(opts ...Option) *Client {
	cfg := &Config{
		BaseURL:    "https://api.glideidentity.app",
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: time.Second,
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

	// Initialize rate limiter if configured
	if cfg.RateLimitEnabled {
		limit := rate.Every(cfg.RateLimitPeriod / time.Duration(cfg.RateLimitRate))
		client.rateLimiter = rate.NewLimiter(limit, cfg.RateLimitRate)
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
