package glide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// rateLimiter handles optional rate limiting
type rateLimiter struct {
	limiter *rate.Limiter
	enabled bool
}

// httpTransport wraps the HTTP client with retry and rate limiting
type httpTransport struct {
	client      *http.Client
	rateLimiter *rateLimiter
	config      *Config
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Apply rate limiting if enabled
	if c.config.RateLimitEnabled && c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, NewError(ErrCodeRateLimitExceeded, "Client-side rate limit exceeded")
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		// Add retry delay (except for first attempt)
		if attempt > 0 {
			select {
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, NewError(ErrCodeUnexpectedError, "Request cancelled")
			}
		}

		// Perform the request
		respData, err := c.performRequest(ctx, method, path, body)
		if err == nil {
			return respData, nil
		}

		// Check if error is retryable
		if glideErr, ok := err.(*Error); ok {
			if !glideErr.IsRetryable() {
				return nil, err
			}
		}

		lastErr = err
	}

	return nil, lastErr
}

// performRequest executes a single HTTP request
func (c *Client) performRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Build URL
	url := c.config.BaseURL + path

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, NewError(ErrCodeInvalidParameters, "Failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, NewError(ErrCodeUnexpectedError, "Failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "glide-go-sdk/1.0.0")

	// Add authentication
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewError(ErrCodeServiceUnavailable, "Failed to execute request")
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError(ErrCodeUnexpectedError, "Failed to read response body")
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// parseErrorResponse parses an error response from the API
func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	var apiErr struct {
		Code      string                 `json:"code"`
		Message   string                 `json:"message"`
		RequestID string                 `json:"request_id,omitempty"`
		Details   map[string]interface{} `json:"details,omitempty"`
	}

	// Try to parse JSON error
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Code != "" {
		// Create error from API response
		glideErr := &Error{
			Code:      apiErr.Code,
			Message:   apiErr.Message,
			Status:    statusCode,
			RequestID: apiErr.RequestID,
			Details:   apiErr.Details,
		}

		// Sanitize the error before returning
		return sanitizeError(glideErr)
	}

	// Fallback to generic error based on status code
	return c.genericErrorForStatus(statusCode)
}

// genericErrorForStatus creates a generic error based on HTTP status
func (c *Client) genericErrorForStatus(status int) error {
	switch status {
	case 400:
		return NewErrorWithStatus(ErrCodeBadRequest, "Invalid request", status)
	case 401:
		return NewErrorWithStatus(ErrCodeUnauthorized, "Authentication required", status)
	case 403:
		return NewErrorWithStatus(ErrCodeForbidden, "Access denied", status)
	case 404:
		return NewErrorWithStatus(ErrCodeNotFound, "Resource not found", status)
	case 422:
		return NewErrorWithStatus(ErrCodeUnprocessableEntity, "Request could not be processed", status)
	case 429:
		return NewErrorWithStatus(ErrCodeRateLimitExceeded, "Too many requests", status)
	case 503:
		return NewErrorWithStatus(ErrCodeServiceUnavailable, "Service temporarily unavailable", status)
	default:
		if status >= 500 {
			return NewErrorWithStatus(ErrCodeInternalServerError, "Server error occurred", status)
		}
		return NewErrorWithStatus(ErrCodeUnexpectedError, fmt.Sprintf("Unexpected status: %d", status), status)
	}
}

// initRateLimiter initializes the rate limiter if configured
func (c *Client) initRateLimiter() {
	if c.config.RateLimitEnabled {
		limit := rate.Every(c.config.RateLimitPeriod / time.Duration(c.config.RateLimitRate))
		c.rateLimiter = rate.NewLimiter(limit, c.config.RateLimitRate)
	}
}
