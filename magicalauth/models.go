package magicalauth

import (
	"encoding/json"
)

// UseCase represents the authentication use case.
type UseCase string

const (
	// UseCaseGetPhoneNumber retrieves the device's phone number.
	UseCaseGetPhoneNumber UseCase = "GetPhoneNumber"

	// UseCaseVerifyPhoneNumber verifies a provided phone number matches the device.
	UseCaseVerifyPhoneNumber UseCase = "VerifyPhoneNumber"
)

// AuthenticationStrategy represents the authentication strategy returned by the API.
type AuthenticationStrategy string

const (
	StrategyTS43    AuthenticationStrategy = "ts43"
	StrategyLink    AuthenticationStrategy = "link"
	StrategyDesktop AuthenticationStrategy = "desktop"
)

// ProtocolType represents the protocol used for authentication.
type ProtocolType string

const (
	ProtocolTS43    ProtocolType = "ts43"
	ProtocolLink    ProtocolType = "link"
	ProtocolDesktop ProtocolType = "desktop"
)

// SessionStatus represents the status of a verification session.
type SessionStatus string

const (
	StatusPending           SessionStatus = "pending"
	StatusScanned           SessionStatus = "scanned"
	StatusPendingCompletion SessionStatus = "pending_completion"
	StatusCompleted         SessionStatus = "completed"
	StatusFailed            SessionStatus = "failed"
)

// RiskLevel represents the SIM swap risk level.
type RiskLevel string

const (
	RiskLevelHigh    RiskLevel = "RISK_LEVEL_HIGH"
	RiskLevelMedium  RiskLevel = "RISK_LEVEL_MEDIUM"
	RiskLevelLow     RiskLevel = "RISK_LEVEL_LOW"
	RiskLevelUnknown RiskLevel = "RISK_LEVEL_UNKNOWN"
)

// SimSwapFailureReason represents why a SIM swap check failed.
type SimSwapFailureReason string

const (
	SimSwapReasonTimeout           SimSwapFailureReason = "timeout"
	SimSwapReasonCarrierNotSupport SimSwapFailureReason = "carrier_not_supported"
	SimSwapReasonDisabled          SimSwapFailureReason = "disabled"
	SimSwapReasonError             SimSwapFailureReason = "error"
)

// Theme represents the UI theme for authentication flows.
//
// Free-form string (max 32 characters). ThemeDark and ThemeLight cover the
// built-in palettes; callers can also pass any custom theme identifier
// understood by their downstream surfaces. The aggregator forwards the
// value verbatim without server-side translation.
type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

// LinkType represents the link strategy dispatch override.
// Only applies to Link authentication strategy; ignored for TS43/Desktop.
type LinkType string

const (
	LinkTypeAppClip  LinkType = "app_clip"
	LinkTypeAppLink  LinkType = "app_link"
	LinkTypeHeadless LinkType = "headless"
	LinkTypeWebLink  LinkType = "web_link"
)

// PLMN represents Mobile Country Code and Mobile Network Code.
type PLMN struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

// ClientInfo contains client information for strategy selection.
//
// IMPORTANT: The UserAgent field MUST be the actual browser's navigator.userAgent
// from the end-user's device, NOT the server's user agent. The SDK runs on your
// backend, but this value must come from your frontend (e.g., passed via your API).
type ClientInfo struct {
	// UserAgent is the end-user's browser user agent string (REQUIRED).
	// Must be captured from the frontend via navigator.userAgent.
	UserAgent string `json:"user_agent"`
	// Platform is the platform identifier from the end-user's browser (optional).
	Platform string `json:"platform,omitempty"`
}

// ConsentData contains user consent information.
type ConsentData struct {
	ConsentText string `json:"consent_text"`
	PolicyLink  string `json:"policy_link"`
	PolicyText  string `json:"policy_text,omitempty"`
}

// PrepareOptions contains advanced/optional configuration for prepare requests.
type PrepareOptions struct {
	ParentSessionID string   `json:"parent_session_id,omitempty"`
	Theme           Theme    `json:"theme,omitempty"`
	LinkType        LinkType `json:"link_type,omitempty"`
}

// SessionInfo contains session information from the API.
type SessionInfo struct {
	SessionKey   string            `json:"session_key"`
	Nonce        string            `json:"nonce,omitempty"`
	ProtocolType ProtocolType      `json:"protocol_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Challenge contains visual verification code for desktop-mobile QR authentication.
type Challenge struct {
	Pattern   string `json:"pattern,omitempty"`
	Color     string `json:"color,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// ========================
// Strategy-Specific Data Types
// ========================

// TS43StrategyData contains TS43 protocol data for T-Mobile authentication.
type TS43StrategyData struct {
	Protocol string          `json:"protocol"` // e.g., "openid4vp-v1-unsigned"
	Data     TS43RequestData `json:"data"`
}

// TS43RequestData contains OpenID4VP request data for navigator.credentials.get.
type TS43RequestData struct {
	Nonce        string    `json:"nonce"`
	ResponseMode string    `json:"response_mode"`
	ResponseType string    `json:"response_type"`
	DCQLQuery    DCQLQuery `json:"dcql_query"`
}

// DCQLQuery contains the DCQL credential query.
type DCQLQuery struct {
	Credentials []DCQLCredential `json:"credentials"`
}

// DCQLCredential contains a DCQL credential request specification.
type DCQLCredential struct {
	ID     string             `json:"id"`
	Format string             `json:"format"`
	Meta   DCQLCredentialMeta `json:"meta"`
	Claims []string           `json:"claims,omitempty"`
}

// DCQLCredentialMeta contains credential metadata.
type DCQLCredentialMeta struct {
	VctValues                  []string `json:"vct_values"`
	CredentialAuthorizationJWT string   `json:"credential_authorization_jwt"`
}

// LinkStrategyData contains Link/OAuth strategy data for Verizon authentication.
type LinkStrategyData struct {
	URL       string            `json:"url"`                  // OAuth authorization URL
	ReturnURL string            `json:"return_url,omitempty"` // URL to return to after OAuth
	StatusURL string            `json:"status_url,omitempty"` // URL to poll for status
	Params    map[string]string `json:"params,omitempty"`     // Additional OAuth parameters
	FeCode    string            `json:"fe_code,omitempty"`    // Server-generated fe_code (when fe_hash was not provided)
	LinkType  LinkType          `json:"link_type,omitempty"`  // How the client SDK should complete the link flow
}

// DesktopStrategyData contains Desktop QR code strategy data for cross-device authentication.
type DesktopStrategyData struct {
	Protocol string        `json:"protocol"` // e.g., "qr-auth-v1"
	Data     DesktopQRData `json:"data"`
}

// DesktopQRData contains QR code and session binding information.
//
// A single universal QR/URL is returned: the Mobile Auth companion app detects
// the device at runtime and picks TS43 (Android) or Link (iOS).
type DesktopQRData struct {
	SessionID string     `json:"session_id"`
	QRImage   string     `json:"qr_image"`   // Base64-encoded PNG of the universal QR code (data URL).
	URL       string     `json:"url"`        // Universal mobile auth URL encoded in the QR.
	StatusURL string     `json:"status_url"` // URL to poll for authentication status.
	Challenge *Challenge `json:"challenge,omitempty"`
}

// SimSwapInfo contains SIM swap check results.
type SimSwapInfo struct {
	Checked     bool                 `json:"checked"`
	RiskLevel   RiskLevel            `json:"risk_level,omitempty"`
	AgeBand     string               `json:"age_band,omitempty"`
	CarrierName string               `json:"carrier_name,omitempty"`
	CheckedAt   string               `json:"checked_at,omitempty"`
	Reason      SimSwapFailureReason `json:"reason,omitempty"`
}

// DeviceInfo contains information about the device that scanned the QR code.
type DeviceInfo struct {
	Type    string `json:"type"`              // Required: Device type/model
	OS      string `json:"os"`                // Required: Operating system and version
	Browser string `json:"browser,omitempty"` // Optional: Browser name and version
}

// EligibilityRequest represents a request to check authentication eligibility.
type EligibilityRequest struct {
	PhoneNumber string      `json:"phone_number,omitempty"`
	PLMN        *PLMN       `json:"plmn,omitempty"`
	ClientInfo  *ClientInfo `json:"client_info,omitempty"`
}

// Requirements describes the device that completes an option (always the phone).
// Typed fields cover the documented contract; Extra captures any future/unknown
// fields so the SDK never drops data and stays forward-compatible.
type Requirements struct {
	Target            *string `json:"target,omitempty"`              // "self" | "companion"
	AndroidMinVersion *int    `json:"android_min_version,omitempty"` // e.g. 14
	AndroidMinSDK     *int    `json:"android_min_sdk,omitempty"`     // e.g. 34
	IOSMinVersion     *int    `json:"ios_min_version,omitempty"`     // e.g. 14

	// Extra holds any fields not modeled above (forward-compat).
	Extra map[string]any `json:"-"`
}

func (r *Requirements) UnmarshalJSON(b []byte) error {
	type alias Requirements
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*r = Requirements(a)
	var all map[string]any
	if err := json.Unmarshal(b, &all); err != nil {
		return err
	}
	for _, k := range []string{"target", "android_min_version", "android_min_sdk", "ios_min_version"} {
		delete(all, k)
	}
	if len(all) > 0 {
		r.Extra = all
	}
	return nil
}

// AvailableOption represents an available authentication strategy option.
type AvailableOption struct {
	AuthenticationStrategy string       `json:"authentication_strategy"`
	Name                   string       `json:"name"`
	Description            string       `json:"description"`
	AvailablePlatforms     []string     `json:"available_platforms"`
	Requirements           Requirements `json:"requirements"`
}

// EligibilityResponse represents the response from the eligibility check.
type EligibilityResponse struct {
	Eligible         bool              `json:"eligible"`
	AvailableOptions []AvailableOption `json:"available_options"`
}

// PrepareRequest is the request for preparing authentication.
type PrepareRequest struct {
	PhoneNumber string       `json:"phone_number,omitempty"`
	PLMN        *PLMN        `json:"plmn,omitempty"`
	Nonce       string       `json:"nonce"`
	UseCase     UseCase      `json:"use_case"`
	ConsentData *ConsentData `json:"consent_data,omitempty"`
	// ClientInfo is REQUIRED. Must contain the end-user's browser user agent.
	ClientInfo ClientInfo      `json:"client_info"`
	Options    *PrepareOptions `json:"options,omitempty"`

	// AuthenticationStrategy optionally overrides the server's strategy selection.
	AuthenticationStrategy string `json:"authentication_strategy,omitempty"`
	// SessionTimeout optionally sets the session timeout in seconds (60-900).
	SessionTimeout *int `json:"session_timeout,omitempty"`

	// FeHash is set internally by the SDK (SHA256 hex of the generated fe_code).
	// Callers should NOT set this field directly.
	FeHash string `json:"fe_hash,omitempty"`

	// PublicKey is a base64-encoded x963 P-256 public key (65 bytes unencoded)
	// for native device binding (SE/TEE). Only used for native link protocol sessions.
	PublicKey string `json:"public_key,omitempty"`
}

// PrepareResponse is the response from preparing authentication.
type PrepareResponse struct {
	AuthenticationStrategy AuthenticationStrategy `json:"authentication_strategy"`
	Session                SessionInfo            `json:"session"`
	Data                   json.RawMessage        `json:"data"`
	Challenge              *Challenge             `json:"challenge,omitempty"`

	// SessionExpiresInSeconds is the effective TTL applied to this session, in
	// seconds. Equals the requested session_timeout when provided, otherwise the
	// server-configured default. Pointer type to distinguish "field absent"
	// (older server) from "TTL is 0" -- matches the Node SDK
	// (`number | undefined`) and the Java SDK (`Integer`).
	SessionExpiresInSeconds *int `json:"session_expires_in_seconds,omitempty"`
}

// VerifyPhoneNumberRequest is the request for verifying a phone number.
type VerifyPhoneNumberRequest struct {
	Session    SessionInfo `json:"session"`
	Credential string      `json:"credential"`
	FeCode     string      `json:"fe_code,omitempty"` // Required for device-bound link protocol sessions (web)
	// Signature is a base64-encoded DER ECDSA signature for native device binding.
	// The native app signs "{session_key}:{credential}" with its SE/TEE private key.
	Signature string `json:"signature,omitempty"`
}

// VerifyPhoneNumberResponse is the response from phone number verification.
type VerifyPhoneNumberResponse struct {
	PhoneNumber        string              `json:"phone_number"`
	Verified           bool                `json:"verified"`
	Aud                string              `json:"aud,omitempty"`
	SimSwap            *SimSwapInfo        `json:"sim_swap,omitempty"`
	DeviceSwap         *DeviceSwapInfo     `json:"device_swap,omitempty"`
	PhoneNumberDetails *PhoneNumberDetails `json:"phone_number_details,omitempty"`
}

// GetPhoneNumberRequest is the request for retrieving a phone number.
type GetPhoneNumberRequest struct {
	Session    SessionInfo `json:"session"`
	Credential string      `json:"credential"`
	FeCode     string      `json:"fe_code,omitempty"` // Required for device-bound link protocol sessions (web)
	// Signature is a base64-encoded DER ECDSA signature for native device binding.
	// The native app signs "{session_key}:{credential}" with its SE/TEE private key.
	Signature string `json:"signature,omitempty"`
}

// GetPhoneNumberResponse is the response from phone number retrieval.
type GetPhoneNumberResponse struct {
	PhoneNumber        string              `json:"phone_number"`
	Aud                string              `json:"aud,omitempty"`
	SimSwap            *SimSwapInfo        `json:"sim_swap,omitempty"`
	DeviceSwap         *DeviceSwapInfo     `json:"device_swap,omitempty"`
	PhoneNumberDetails *PhoneNumberDetails `json:"phone_number_details,omitempty"`
}

// ReportInvocationRequest is the request for reporting an invocation event.
type ReportInvocationRequest struct {
	SessionID string `json:"session_id"`
}

// ReportInvocationResponse is the response from reporting an invocation.
//
// The backend returns `{"success": true}` on 2xx. Invocation recording is
// idempotent — reporting the same session_id twice also returns
// `{"success": true}`.
type ReportInvocationResponse struct {
	Success bool `json:"success"`
}

// StatusResponse is the response from checking session status.
// Note: This endpoint does NOT expose phone_number or verification results directly.
// Sensitive data is only returned via the protected verify/get-phone-number endpoints.
type StatusResponse struct {
	SessionKey     string                 `json:"session_key"`
	Status         SessionStatus          `json:"status"`
	Protocol       ProtocolType           `json:"protocol,omitempty"` // Optional: ts43, link, or desktop
	CreatedAt      string                 `json:"created_at"`
	LastUpdated    string                 `json:"last_updated"`
	DeviceInfo     *DeviceInfo            `json:"device_info,omitempty"`
	GraceExpiresAt string                 `json:"grace_expires_at,omitempty"`
	ScannedAt      string                 `json:"scanned_at,omitempty"`
	Extra          map[string]interface{} `json:"extra,omitempty"` // Additional metadata (e.g., error reason)
}

// DeviceSwapInfo contains device swap (IMEI change) check results.
// Parallel to SimSwapInfo; only present when the carrier supports device swap detection.
type DeviceSwapInfo struct {
	Checked     bool      `json:"checked"`
	RiskLevel   RiskLevel `json:"risk_level,omitempty"`
	AgeBand     string    `json:"age_band,omitempty"`
	CarrierName string    `json:"carrier_name,omitempty"`
	CheckedAt   string    `json:"checked_at,omitempty"`
}

// PhoneNumberDetails contains parsed components of an E.164 phone number.
type PhoneNumberDetails struct {
	CountryCode    int    `json:"country_code"`
	NationalNumber string `json:"national_number"`
	RegionCode     string `json:"region_code"`
}

// PrepareResult extends PrepareResponse with SDK-generated device binding info.
// Returned by Client.Prepare instead of the raw PrepareResponse.
type PrepareResult struct {
	PrepareResponse

	// FeCode is the frontend binding code (64-char hex).
	// Only populated when the strategy is "link".
	// The caller MUST set this as an HttpOnly cookie using BuildSetBindingCookieHeader.
	FeCode string `json:"-"`
}

// CompleteRequest is the request for completing device-bound authentication.
type CompleteRequest struct {
	SessionKey string `json:"session_key"`
	FeCode     string `json:"fe_code"`
	AggCode    string `json:"agg_code"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// CompleteResponse is the response from completing device-bound authentication.
type CompleteResponse struct {
	Status string `json:"status"`
}

// DesktopWaitResponse is returned with HTTP 202 when a desktop QR session's
// wait timer expires before the mobile device completes authentication.
type DesktopWaitResponse struct {
	Status                  string `json:"status"`
	Message                 string `json:"message"`
	SessionKey              string `json:"session_key"`
	SessionStatus           string `json:"session_status"`
	Retry                   bool   `json:"retry"`
	SessionExpiresInSeconds int    `json:"session_expires_in_seconds"`
}

// BindingCookieOptions configures the device binding cookie.
type BindingCookieOptions struct {
	Domain string // Cookie domain (optional)
	Path   string // Cookie path (default: "/")
	Secure bool   // Secure flag (default: true; set false for local HTTP dev)
}

// tokenResponse is the internal response from the OAuth2 token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// errorResponse is the internal error response from the API.
type errorResponse struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Status    int                 `json:"status,omitempty"`
	RequestID string              `json:"request_id,omitempty"`
	Timestamp string              `json:"timestamp,omitempty"`
	TraceID   string              `json:"trace_id,omitempty"`
	SpanID    string              `json:"span_id,omitempty"`
	Service   string              `json:"service,omitempty"`
	Details   *errorResponseDeets `json:"details,omitempty"`
}

type errorResponseDeets struct {
	Fields map[string]string `json:"fields,omitempty"`
}

// oauthError is the OAuth2 error response.
type oauthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
