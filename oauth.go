package glide

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// oauthManager handles OAuth2 client credentials flow with token caching.
type oauthManager struct {
	baseURL            string
	clientID           string
	clientSecret       string
	httpClient         *http.Client
	tokenRefreshBuffer int
	logger             Logger

	// Token cache
	cachedToken     string
	cachedExpiresAt time.Time
	tokenMutex      sync.Mutex
}

// tokenResponse represents the OAuth2 token response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// newOAuthManager creates a new OAuth manager.
func newOAuthManager(baseURL, clientID, clientSecret string, httpClient *http.Client, tokenRefreshBuffer int, logger Logger) *oauthManager {
	if tokenRefreshBuffer <= 0 {
		tokenRefreshBuffer = 60
	}
	return &oauthManager{
		baseURL:            baseURL,
		clientID:           clientID,
		clientSecret:       clientSecret,
		httpClient:         httpClient,
		tokenRefreshBuffer: tokenRefreshBuffer,
		logger:             logger,
	}
}

// GetAccessToken retrieves a valid OAuth2 access token, refreshing if necessary.
func (o *oauthManager) GetAccessToken(ctx context.Context) (string, error) {
	o.tokenMutex.Lock()
	defer o.tokenMutex.Unlock()

	// Check if we have a valid cached token
	if o.cachedToken != "" && time.Now().Before(o.cachedExpiresAt.Add(-time.Duration(o.tokenRefreshBuffer)*time.Second)) {
		return o.cachedToken, nil
	}

	// Fetch new token
	o.logger.Debug("Fetching new OAuth2 access token")

	tokenURL := o.baseURL + "/oauth2-cc/token"
	credentials := base64.StdEncoding.EncodeToString([]byte(o.clientID + ":" + o.clientSecret))

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return "", NewMagicalAuthError(ErrCodeInternalServerError, "failed to create token request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", NewMagicalAuthError(ErrCodeServiceUnavailable, "token request failed: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", NewMagicalAuthError(ErrCodeConfigurationError, "Invalid client credentials")
	}

	if resp.StatusCode >= 400 {
		return "", NewMagicalAuthError(ErrCodeServiceUnavailable, fmt.Sprintf("Token fetch failed with status %d", resp.StatusCode))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewMagicalAuthError(ErrCodeInternalServerError, "failed to read token response")
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", NewMagicalAuthError(ErrCodeInternalServerError, "failed to parse token response")
	}

	if tokenResp.AccessToken == "" {
		return "", NewMagicalAuthError(ErrCodeInternalServerError, "token response missing access_token")
	}

	// Cache the token (default to 1 hour if expires_in not provided)
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	o.cachedToken = tokenResp.AccessToken
	o.cachedExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	o.logger.Debug("OAuth2 token fetched successfully",
		Field{"expires_in", expiresIn},
	)

	return o.cachedToken, nil
}

// ClearToken clears the cached token (useful for testing or forced refresh).
func (o *oauthManager) ClearToken() {
	o.tokenMutex.Lock()
	defer o.tokenMutex.Unlock()
	o.cachedToken = ""
	o.cachedExpiresAt = time.Time{}
}
