package magicalauth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	tokenPath       = "/oauth2-cc/token" // #nosec G101 -- URL path, not a token
	tokenBufferSecs = 60                 // Refresh token 60 seconds before expiry
)

// TokenManager handles OAuth2 token caching and refresh.
// It is thread-safe and automatically refreshes tokens before expiry.
type TokenManager struct {
	clientID     string
	clientSecret string
	tokenURL     string
	httpClient   HTTPClient
	logger       Logger

	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

// newTokenManager creates a new TokenManager.
func newTokenManager(cfg *Config) *TokenManager {
	return &TokenManager{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		tokenURL:     cfg.BaseURL + tokenPath,
		httpClient:   cfg.HTTPClient,
		logger:       cfg.Logger,
	}
}

// GetToken returns a valid access token, fetching a new one if necessary.
// It uses double-checked locking to minimize lock contention.
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	// Fast path: check if we have a valid token
	tm.mu.RLock()
	if tm.token != "" && time.Now().Add(tokenBufferSecs*time.Second).Before(tm.expiresAt) {
		token := tm.token
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// Slow path: need to fetch a new token
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have fetched)
	if tm.token != "" && time.Now().Add(tokenBufferSecs*time.Second).Before(tm.expiresAt) {
		return tm.token, nil
	}

	return tm.fetchToken(ctx)
}

// fetchToken fetches a new access token from the OAuth2 endpoint.
// Must be called with tm.mu held for writing.
func (tm *TokenManager) fetchToken(ctx context.Context) (string, error) {
	tm.logger.Debug("fetching new access token", "url", tm.tokenURL)

	// Build form-encoded request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	body := strings.NewReader(data.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tm.tokenURL, body)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	// Set Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(tm.clientID + ":" + tm.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", tm.parseTokenError(respBody, resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	// Cache the token
	tm.token = tokenResp.AccessToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	tm.logger.Debug("access token obtained", "expires_in", tokenResp.ExpiresIn)

	return tm.token, nil
}

// parseTokenError parses an OAuth2 error response.
func (tm *TokenManager) parseTokenError(body []byte, status int) error {
	var oauthErr oauthError
	if err := json.Unmarshal(body, &oauthErr); err != nil {
		// If we can't parse, return generic error
		return &APIError{
			Err:     ErrUnauthorized,
			Code:    CodeUnauthorized,
			Message: "authentication failed",
			Status:  status,
		}
	}
	return newOAuthError(&oauthErr, status)
}

// ClearToken clears the cached token (useful for testing or forced refresh).
func (tm *TokenManager) ClearToken() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.token = ""
	tm.expiresAt = time.Time{}
}
