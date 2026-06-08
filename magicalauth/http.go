package magicalauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// RequestOptions contains per-request options.
type RequestOptions struct {
	// CorrelationID is an optional ID forwarded as X-Correlation-ID header.
	// The server echoes it back in the response.
	CorrelationID string
}

const (
	headerCorrelationID = "X-Correlation-ID"
)

// doRequestOpts performs an HTTP request with optional per-request options.
func (c *Client) doRequestOpts(ctx context.Context, method, path string, body any, result any, opts *RequestOptions) error {
	token, err := c.tokenManager.GetToken(ctx)
	if err != nil {
		return err
	}

	return c.doRequestWithAuth(ctx, method, path, "Bearer "+token, body, result, opts)
}

// doRequestNoAuthOpts performs an HTTP request without authentication but with per-request options.
func (c *Client) doRequestNoAuthOpts(ctx context.Context, method, path string, body any, result any, opts *RequestOptions) error {
	return c.doRequestWithAuth(ctx, method, path, "", body, result, opts)
}

// doRequestWithAuth performs an HTTP request with optional auth header.
func (c *Client) doRequestWithAuth(ctx context.Context, method, path, authHeader string, body any, result any, opts *RequestOptions) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "magic-auth-go/"+Version)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	if opts != nil && opts.CorrelationID != "" {
		req.Header.Set(headerCorrelationID, opts.CorrelationID)
	}

	c.logger.Debug("request", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.logger.Debug("response", "status", resp.StatusCode, "path", path)

	if resp.StatusCode >= 400 {
		return c.parseErrorResponse(respBody, resp.StatusCode, resp.Header)
	}

	// HTTP 202: desktop wait — return typed error so the caller can decide to retry
	if resp.StatusCode == http.StatusAccepted {
		var waitResp DesktopWaitResponse
		if err := json.Unmarshal(respBody, &waitResp); err != nil {
			return fmt.Errorf("decode 202 response: %w", err)
		}
		return &DesktopWaitError{Response: &waitResp}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(body []byte, status int, headers http.Header) error {
	// Try to parse as API error response
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// Couldn't parse - return generic error based on status
		return c.genericError(status, string(body))
	}

	// Extract Retry-After header for rate limiting
	retryAfter := 0
	if ra := headers.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil {
			retryAfter = seconds
		}
	}

	return newAPIError(&errResp, status, retryAfter)
}

// genericError creates a generic error based on HTTP status code.
func (c *Client) genericError(status int, body string) *APIError {
	var sentinel error
	var code ErrorCode
	var message string

	switch status {
	case http.StatusBadRequest:
		sentinel = ErrBadRequest
		code = CodeBadRequest
		message = "bad request"
	case http.StatusUnauthorized:
		sentinel = ErrUnauthorized
		code = CodeUnauthorized
		message = "unauthorized"
	case http.StatusForbidden:
		sentinel = ErrForbidden
		code = CodeForbidden
		message = "forbidden"
	case http.StatusNotFound:
		sentinel = ErrSessionNotFound
		code = CodeSessionNotFound
		message = "not found"
	case http.StatusConflict:
		sentinel = ErrSessionNotEligible
		code = CodeSessionNotEligible
		message = "session not eligible"
	case http.StatusTooManyRequests:
		sentinel = ErrRateLimit
		code = CodeRateLimitExceeded
		message = "rate limit exceeded"
	case http.StatusUnprocessableEntity:
		sentinel = ErrValidation
		code = CodeValidationError
		message = "validation error"
	default:
		sentinel = ErrInternalServer
		code = CodeInternalServerError
		message = "internal server error"
	}

	if body != "" && len(body) < 200 {
		message = body
	}

	return &APIError{
		Err:     sentinel,
		Code:    code,
		Message: message,
		Status:  status,
	}
}
