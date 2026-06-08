package magicalauth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// API paths
const (
	pathEligibility  = "/magic-auth/v2/auth/eligibility"
	pathPrepare      = "/magic-auth/v2/auth/prepare"
	pathVerifyPhone  = "/magic-auth/v2/auth/verify-phone-number"
	pathGetPhone     = "/magic-auth/v2/auth/get-phone-number"
	pathComplete     = "/magic-auth/v2/auth/complete"
	pathReportInvoke = "/magic-auth/v2/auth/report-invocation"
	pathStatus       = "/magic-auth/v2/auth/status/"
	pathStatusPublic = "/public/status/"
)

// Client is the Magic Auth API client.
type Client struct {
	baseURL      string
	httpClient   HTTPClient
	logger       Logger
	tokenManager *TokenManager
}

// NewClient creates a new Magic Auth client with the given configuration.
// Returns an error if required configuration is missing.
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()

	return &Client{
		baseURL:      cfg.BaseURL,
		httpClient:   cfg.HTTPClient,
		logger:       cfg.Logger,
		tokenManager: newTokenManager(&cfg),
	}, nil
}

// CheckEligibility checks authentication eligibility for a phone number or PLMN and returns
// available authentication strategies with platform requirements.
func (c *Client) CheckEligibility(ctx context.Context, req *EligibilityRequest, opts ...*RequestOptions) (*EligibilityResponse, error) {
	if err := c.validateEligibilityRequest(req); err != nil {
		return nil, err
	}

	var resp EligibilityResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathEligibility, req, &resp, reqOpts); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Prepare prepares authentication by checking carrier eligibility
// and generating protocol-specific requests.
//
// For link protocol sessions, Prepare automatically generates a device binding
// fe_code and includes its SHA-256 hash in the request. The fe_code is returned in
// PrepareResult.FeCode so the caller can set it as an HttpOnly cookie via
// BuildSetBindingCookieHeader.
//
// Returns strategy-specific data (TS43, Link, or Desktop) based on carrier and platform.
func (c *Client) Prepare(ctx context.Context, req *PrepareRequest, opts ...*RequestOptions) (*PrepareResult, error) {
	if err := c.validatePrepareRequest(req); err != nil {
		return nil, err
	}

	if req.Nonce == "" {
		nonce, err := generateNonce(32)
		if err != nil {
			return nil, fmt.Errorf("generate nonce: %w", err)
		}
		req.Nonce = nonce
	}

	// Generate fe_code + fe_hash preemptively (aggregator ignores it for non-link)
	feCode, err := GenerateFeCode()
	if err != nil {
		return nil, fmt.Errorf("generate device binding code: %w", err)
	}
	feHash, err := ComputeFeHash(feCode)
	if err != nil {
		return nil, fmt.Errorf("compute device binding hash: %w", err)
	}
	req.FeHash = feHash

	c.logger.Info("preparing auth", "use_case", req.UseCase)

	var resp PrepareResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathPrepare, req, &resp, reqOpts); err != nil {
		return nil, err
	}

	result := &PrepareResult{PrepareResponse: resp}

	// Only expose fe_code for link strategy (device binding is link-only)
	if resp.AuthenticationStrategy == StrategyLink {
		// If the server returned an fe_code in the link data (because we didn't
		// send fe_hash, or for backward compat), prefer the server's code.
		var linkData LinkStrategyData
		if err := json.Unmarshal(resp.Data, &linkData); err == nil && linkData.FeCode != "" {
			result.FeCode = linkData.FeCode
		} else {
			result.FeCode = feCode
		}
	}

	c.logger.Info("auth prepared", "strategy", resp.AuthenticationStrategy, "session_key", resp.Session.SessionKey)
	return result, nil
}

// VerifyPhoneNumber verifies a phone number using digital credentials.
// Returns verified: true if phone matches device, false otherwise.
//
// For desktop QR sessions, the server may return HTTP 202 if the mobile device
// hasn't completed authentication yet. In that case a *DesktopWaitError is returned.
// The caller should inspect its Retry and SessionExpiresInSeconds fields to decide
// whether to re-issue the same request.
func (c *Client) VerifyPhoneNumber(ctx context.Context, req *VerifyPhoneNumberRequest, opts ...*RequestOptions) (*VerifyPhoneNumberResponse, error) {
	if err := c.validateVerifyRequest(req); err != nil {
		return nil, err
	}

	c.logger.Info("verifying phone number", "session_key", req.Session.SessionKey)

	var resp VerifyPhoneNumberResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathVerifyPhone, req, &resp, reqOpts); err != nil {
		return nil, err
	}

	c.logger.Info("phone verification complete", "verified", resp.Verified)
	return &resp, nil
}

// GetPhoneNumber retrieves the phone number associated with the device.
//
// For desktop QR sessions, the server may return HTTP 202 if the mobile device
// hasn't completed authentication yet. In that case a *DesktopWaitError is returned.
// The caller should inspect its Retry and SessionExpiresInSeconds fields to decide
// whether to re-issue the same request.
func (c *Client) GetPhoneNumber(ctx context.Context, req *GetPhoneNumberRequest, opts ...*RequestOptions) (*GetPhoneNumberResponse, error) {
	if err := c.validateGetPhoneRequest(req); err != nil {
		return nil, err
	}

	c.logger.Info("getting phone number", "session_key", req.Session.SessionKey)

	var resp GetPhoneNumberResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathGetPhone, req, &resp, reqOpts); err != nil {
		return nil, err
	}

	c.logger.Info("phone number retrieved", "phone_number", maskPhone(resp.PhoneNumber))
	return &resp, nil
}

// Complete validates device binding for a link protocol session.
// This must be called after the carrier OAuth redirect delivers agg_code to
// the developer's completion page.
//
// The caller should:
//  1. Extract agg_code and session_key from the completion page POST body
//  2. Extract fe_code from the _glide_bind cookie via ParseBindingCookie
//  3. Call Complete with all three values
//  4. Clear the cookie via BuildClearBindingCookieHeader (success or failure)
func (c *Client) Complete(ctx context.Context, req *CompleteRequest, opts ...*RequestOptions) (*CompleteResponse, error) {
	if err := c.validateCompleteRequest(req); err != nil {
		return nil, err
	}

	c.logger.Info("completing device-bound session", "session_key", req.SessionKey)

	var resp CompleteResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathComplete, req, &resp, reqOpts); err != nil {
		return nil, err
	}

	c.logger.Info("device-bound session completed", "session_key", req.SessionKey)
	return &resp, nil
}

// ReportInvocation reports when a user clicks the invoke/authenticate button.
// Used for Authentication Success Rate (ASR) tracking.
func (c *Client) ReportInvocation(ctx context.Context, req *ReportInvocationRequest, opts ...*RequestOptions) (*ReportInvocationResponse, error) {
	if err := c.validateReportRequest(req); err != nil {
		return nil, err
	}

	c.logger.Debug("reporting invocation", "session_id", req.SessionID)

	var resp ReportInvocationResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodPost, pathReportInvoke, req, &resp, reqOpts); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CheckStatus polls for the status of an ongoing verification request.
// This endpoint requires authentication.
func (c *Client) CheckStatus(ctx context.Context, sessionKey string, opts ...*RequestOptions) (*StatusResponse, error) {
	if sessionKey == "" {
		return nil, ErrMissingSessionKey
	}

	c.logger.Debug("checking status", "session_key", sessionKey)

	var resp StatusResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestOpts(ctx, http.MethodGet, pathStatus+url.PathEscape(sessionKey), nil, &resp, reqOpts); err != nil {
		return nil, err
	}

	c.logger.Debug("status retrieved", "status", resp.Status)
	return &resp, nil
}

// CheckStatusPublic polls for the status of an ongoing verification request.
// This endpoint is publicly accessible without authentication.
func (c *Client) CheckStatusPublic(ctx context.Context, sessionKey string, opts ...*RequestOptions) (*StatusResponse, error) {
	if sessionKey == "" {
		return nil, ErrMissingSessionKey
	}

	c.logger.Debug("checking public status", "session_key", sessionKey)

	var resp StatusResponse
	var reqOpts *RequestOptions
	if len(opts) > 0 {
		reqOpts = opts[0]
	}
	if err := c.doRequestNoAuthOpts(ctx, http.MethodGet, pathStatusPublic+url.PathEscape(sessionKey), nil, &resp, reqOpts); err != nil {
		return nil, err
	}

	c.logger.Debug("public status retrieved", "status", resp.Status)
	return &resp, nil
}

// maskPhone masks a phone number for logging (e.g., +1415***1234).
func maskPhone(phone string) string {
	if len(phone) <= 7 {
		return "***"
	}
	return phone[:4] + "***" + phone[len(phone)-4:]
}

const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateNonce(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphanumeric[b[i]%byte(len(alphanumeric))]
	}
	return string(b), nil
}
