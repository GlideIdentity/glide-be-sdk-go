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

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("Failed to marshal request body",
				Field{"error", err.Error()},
			)
			return nil, NewError(ErrCodeValidationError, "Failed to marshal request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
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

	// Track timing
	start := time.Now()

	// Log request with formatter if available (only if pretty format is enabled)
	if dl, ok := c.logger.(*defaultLogger); ok && dl.formatter != nil && dl.format == LogFormatPretty {
		// First show the full formatted request like Node SDK
		operation := getOperationFromURL(url)
		fmt.Printf("\n========== %s REQUEST ==========\n", operation)

		// Build request object for pretty printing
		reqObj := map[string]interface{}{
			"url":    url,
			"method": method,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
		}

		if body != nil {
			reqObj["body"] = body
		}

		// Pretty print the JSON
		if jsonBytes, err := json.MarshalIndent(reqObj, "", "  "); err == nil {
			fmt.Println(string(jsonBytes))
		}
		fmt.Println("================================================\n")

		// Then show the box summary
		details := make(map[string]interface{})
		if body != nil {
			// Add body size
			if bodyBytes, err := json.Marshal(body); err == nil {
				details["body_size"] = len(bodyBytes)
			}
			// Add specific details from body
			if bodyMap, ok := body.(map[string]interface{}); ok {
				if useCase, exists := bodyMap["use_case"]; exists {
					details["use_case"] = useCase
				}
				if plmn, exists := bodyMap["plmn"]; exists {
					details["plmn"] = plmn
				}
			}
		}
		dl.formatter.FormatRequest(method, url, details)
	}

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

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			Field{"error", err.Error()},
		)
		return nil, NewError(ErrCodeInternalServerError, "Failed to read response body")
	}

	// Log response with formatter if available (only if pretty format is enabled)
	if dl, ok := c.logger.(*defaultLogger); ok && dl.formatter != nil && dl.format == LogFormatPretty {
		// Extract operation name from URL for response logging
		operation := getOperationFromURL(url)

		// First show the full formatted response like Node SDK
		fmt.Printf("\n========== %s RESPONSE ==========\n", operation)

		// Build response object for pretty printing
		respObj := map[string]interface{}{
			"status": resp.StatusCode,
		}

		// Parse and add body if available
		if len(respBody) > 0 {
			var bodyData interface{}
			if err := json.Unmarshal(respBody, &bodyData); err == nil {
				respObj["body"] = bodyData
			}
		}

		// Pretty print the JSON
		if jsonBytes, err := json.MarshalIndent(respObj, "", "  "); err == nil {
			fmt.Println(string(jsonBytes))
		}
		fmt.Println("=================================================\n")

		// Then show the box summary
		details := make(map[string]interface{})

		// Add response-specific details if successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(respBody) > 0 {
			var respData map[string]interface{}
			if err := json.Unmarshal(respBody, &respData); err == nil {
				// Add specific fields based on response
				if phoneNumber, exists := respData["phone_number"]; exists {
					details["phone_number"] = phoneNumber
				}
				if verified, exists := respData["verified"]; exists {
					details["verified"] = verified
				}
				if strategy, exists := respData["authentication_strategy"]; exists {
					details["strategy"] = strategy
				}
				if session, exists := respData["session"]; exists {
					if sessionMap, ok := session.(map[string]interface{}); ok {
						if sessionKey, exists := sessionMap["session_key"]; exists {
							details["session_key"] = sessionKey
						}
					}
				}
			}
		}
		dl.formatter.FormatResponse(operation, resp.StatusCode, details)
		fmt.Println() // Add spacing after box
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		// Only log error details if not using pretty format
		if dl, ok := c.logger.(*defaultLogger); !ok || dl.format != LogFormatPretty {
			c.logger.Error("API error response",
				Field{"statusCode", resp.StatusCode},
				Field{"responseSize", len(respBody)},
			)
		}
		// Only log error body if not using pretty format
		if dl, ok := c.logger.(*defaultLogger); !ok || dl.format != LogFormatPretty {
			c.logger.Debug("Error response body", Field{"body", string(respBody)})
		}
		return nil, c.parseErrorResponse(resp.StatusCode, respBody)
	}

	// Only log success if not using pretty format
	if dl, ok := c.logger.(*defaultLogger); !ok || dl.format != LogFormatPretty {
		c.logger.Info("Request completed successfully",
			Field{"statusCode", resp.StatusCode},
			Field{"elapsed", elapsed.String()},
		)
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
		// Unauthorized is not a public error code, use generic internal server error
		return NewErrorWithStatus(ErrCodeInternalServerError, "Authentication failed", status)
	case 403:
		// Forbidden is not a public error code, use generic internal server error
		return NewErrorWithStatus(ErrCodeInternalServerError, "Access denied", status)
	case 404:
		// For sessions, use SESSION_NOT_FOUND, for other resources use generic error
		return NewErrorWithStatus(ErrCodeSessionNotFound, "Resource not found", status)
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

// getOperationFromURL extracts operation name from URL for logging
func getOperationFromURL(url string) string {
	if strings.Contains(url, "prepare") {
		return "MagicAuth PREPARE"
	} else if strings.Contains(url, "verify-phone-number") {
		return "MagicAuth VERIFY PHONE"
	} else if strings.Contains(url, "get-phone-number") {
		return "MagicAuth GET PHONE"
	} else if strings.Contains(url, "sim-swap") {
		if strings.Contains(url, "check") {
			return "SimSwap CHECK"
		}
		return "SimSwap RETRIEVE DATE"
	} else if strings.Contains(url, "kyc-match") {
		return "KYC MATCH"
	}
	return "API Request"
}
