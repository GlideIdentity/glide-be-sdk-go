package glide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// doRequest performs an HTTP request with retry logic.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying request",
				Field{"attempt", attempt},
				Field{"delay", c.config.RetryDelay * time.Duration(attempt)},
			)
			select {
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Request cancelled")
			}
		}

		respData, err := c.performRequest(ctx, method, path, body)
		if err == nil {
			return respData, nil
		}

		if glideErr, ok := err.(*MagicalAuthError); ok {
			// Only retry on server errors or rate limits
			if glideErr.Status < 500 && glideErr.Code != ErrCodeRateLimitExceeded && glideErr.Code != ErrCodeServiceUnavailable {
				return nil, err
			}
		}

		lastErr = err
	}

	return nil, lastErr
}

// performRequest executes a single HTTP request.
func (c *Client) performRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Validate configuration before making request
	if err := c.validateConfig(); err != nil {
		return nil, err
	}

	reqURL := c.config.BaseURL + path

	// Marshal body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
	}
	c.logger.Debug("Request", Field{"method", method}, Field{"path", path})

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "glide-be-sdk-go/"+GetVersion())

	// Add OAuth2 authentication header
	accessToken, err := c.oauth.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute request
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Error("HTTP request failed", Field{"error", err.Error()}, Field{"elapsed", elapsed.String()})
		return nil, NewMagicalAuthError(ErrCodeServiceUnavailable, "Failed to execute request")
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to read response body")
	}

	c.logger.Debug("Response", Field{"status", resp.StatusCode}, Field{"elapsed", elapsed.String()})

	// Handle errors
	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	var apiErr struct {
		Code      string                 `json:"code"`
		Message   string                 `json:"message"`
		RequestID string                 `json:"request_id,omitempty"`
		Timestamp string                 `json:"timestamp,omitempty"`
		TraceID   string                 `json:"trace_id,omitempty"`
		SpanID    string                 `json:"span_id,omitempty"`
		Service   string                 `json:"service,omitempty"`
		Details   map[string]interface{} `json:"details,omitempty"`
	}

	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Code != "" {
		return &MagicalAuthError{
			Code:      apiErr.Code,
			Message:   apiErr.Message,
			Status:    statusCode,
			RequestID: apiErr.RequestID,
			Timestamp: apiErr.Timestamp,
			TraceID:   apiErr.TraceID,
			SpanID:    apiErr.SpanID,
			Service:   apiErr.Service,
			Details:   apiErr.Details,
		}
	}

	return c.genericErrorForStatus(statusCode)
}

// genericErrorForStatus creates a generic error based on HTTP status.
func (c *Client) genericErrorForStatus(status int) error {
	switch status {
	case 400:
		return NewMagicalAuthErrorWithStatus(ErrCodeMissingRequiredField, "Invalid request", status)
	case 401, 403:
		return NewMagicalAuthErrorWithStatus(ErrCodeConfigurationError, "Authentication failed", status)
	case 404:
		return NewMagicalAuthErrorWithStatus(ErrCodeInvalidSession, "Resource not found", status)
	case 422:
		return NewMagicalAuthErrorWithStatus(ErrCodeCarrierNotEligible, "Request could not be processed", status)
	case 429:
		return NewMagicalAuthErrorWithStatus(ErrCodeRateLimitExceeded, "Too many requests", status)
	case 503:
		return NewMagicalAuthErrorWithStatus(ErrCodeServiceUnavailable, "Service temporarily unavailable", status)
	default:
		if status >= 500 {
			return NewMagicalAuthErrorWithStatus(ErrCodeInternalServerError, "Server error occurred", status)
		}
		return NewMagicalAuthErrorWithStatus(ErrCodeInternalServerError, fmt.Sprintf("Unexpected status: %d", status), status)
	}
}
