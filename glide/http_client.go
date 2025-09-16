package glide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// httpTransport wraps the HTTP client with retry and rate limiting
type httpTransport struct {
	client      *http.Client
	rateLimiter *rate.Limiter
	config      *Config
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Apply rate limiting if enabled
	if c.config.RateLimitEnabled && c.rateLimiter != nil {
		c.logger.Debug("Applying rate limiting",
			Field{"method", method},
			Field{"path", path},
		)
		if err := c.rateLimiter.Wait(ctx); err != nil {
			c.logger.Error("Rate limit exceeded",
				Field{"error", err.Error()},
			)
			return nil, NewError(ErrCodeRateLimitExceeded, "Client-side rate limit exceeded")
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		// Add retry delay (except for first attempt)
		if attempt > 0 {
			c.logger.Debug("Retrying request",
				Field{"attempt", attempt},
				Field{"delay", c.config.RetryDelay * time.Duration(attempt)},
			)
			select {
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				c.logger.Error("Request cancelled during retry",
					Field{"attempt", attempt},
				)
				return nil, NewError(ErrCodeInternalServerError, "Request cancelled")
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
				c.logger.Error("Non-retryable error",
					Field{"error", glideErr.Error()},
					Field{"code", glideErr.Code},
				)
				return nil, err
			}
			c.logger.Warn("Retryable error occurred",
				Field{"error", glideErr.Error()},
				Field{"code", glideErr.Code},
				Field{"attempt", attempt},
			)
		}

		lastErr = err
	}

	c.logger.Error("All retry attempts exhausted",
		Field{"lastError", lastErr.Error()},
		Field{"retryCount", c.config.RetryCount},
	)
	return nil, lastErr
}

// performRequest executes a single HTTP request
func (c *Client) performRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Build URL with API key as query parameter
	url := c.config.BaseURL + path
	if c.config.APIKey != "" {
		// Add API key as query parameter (like Node SDK)
		if strings.Contains(url, "?") {
			url += "&apikey=" + url2.QueryEscape(c.config.APIKey)
		} else {
			url += "?apikey=" + url2.QueryEscape(c.config.APIKey)
		}
	}

	// Log the request
	c.logger.Debug("Preparing HTTP request",
		Field{"method", method},
		Field{"url", url},
	)

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body",
				Field{"error", err.Error()},
			)
			return nil, NewError(ErrCodeInvalidParameters, "Failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
		c.logger.Debug("Request body prepared",
			Field{"size", len(jsonBody)},
		)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		c.logger.Error("Failed to create request",
			Field{"error", err.Error()},
		)
		return nil, NewError(ErrCodeInternalServerError, "Failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "glide-go-sdk/1.0.0")

	// API key is added as query parameter above, not as header
	if c.config.APIKey != "" {
		c.logger.Debug("API key authentication added",
			Field{"apiKey", c.config.APIKey}, // Will be automatically redacted
		)
	}

	// Log request details
	start := time.Now()
	c.logger.Info("Sending HTTP request",
		Field{"method", method},
		Field{"path", path},
	)

	// Execute request
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Error("HTTP request failed",
			Field{"error", err.Error()},
			Field{"elapsed", elapsed.String()},
		)
		return nil, NewError(ErrCodeServiceUnavailable, "Failed to execute request")
	}
	defer resp.Body.Close()

	// Log response status
	c.logger.Debug("HTTP response received",
		Field{"statusCode", resp.StatusCode},
		Field{"elapsed", elapsed.String()},
	)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			Field{"error", err.Error()},
		)
		return nil, NewError(ErrCodeInternalServerError, "Failed to read response body")
	}

	c.logger.Debug("Response body received",
		Field{"size", len(respBody)},
	)

	// Check for errors
	if resp.StatusCode >= 400 {
		c.logger.Error("API error response",
			Field{"statusCode", resp.StatusCode},
			Field{"responseSize", len(respBody)},
		)
		c.logger.Debug("Error response body", Field{"body", string(respBody)})
		return nil, c.parseErrorResponse(resp.StatusCode, respBody)
	}

	c.logger.Info("Request completed successfully",
		Field{"statusCode", resp.StatusCode},
		Field{"elapsed", elapsed.String()},
	)

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
		return NewErrorWithStatus(ErrCodeInternalServerError, fmt.Sprintf("Unexpected status: %d", status), status)
	}
}

// initRateLimiter initializes the rate limiter if configured
func (c *Client) initRateLimiter() {
	if c.config.RateLimitEnabled {
		limit := rate.Every(c.config.RateLimitPeriod / time.Duration(c.config.RateLimitRate))
		c.rateLimiter = rate.NewLimiter(limit, c.config.RateLimitRate)
	}
}
