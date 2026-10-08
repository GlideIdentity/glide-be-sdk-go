package magicalauth

import "encoding/json"

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
	// PrivacyMode is the browsing privacy mode from the end-user's browser (optional).
	PrivacyMode string `json:"privacy_mode,omitempty"`
}

// ConsentData contains user consent information.
type ConsentData struct {
	ConsentText string `json:"consent_text"`
	PolicyLink  string `json:"policy_link"`
	PolicyText  string `json:"policy_text,omitempty"`
}

// PrepareOptions contains advanced/optional configuration for prepare requests.
type PrepareOptions struct {
	ParentSessionID    string   `json:"parent_session_id,omitempty"`
	Theme              Theme    `json:"theme,omitempty"`
	LinkType           LinkType `json:"link_type,omitempty"`
	SimSwapMaxAgeHours *int     `json:"sim_swap_max_age_hours,omitempty"`

	// On Desktop (QR), set on the parent prepare; the phone inherits it.
	// The outcome comes back as BiometricVerification on verify / get-phone.
	Biometric *bool `json:"biometric,omitempty"`

	// NetworkOrigin carries the server-observed network origin of the requesting
	// device. Stamp it on BOTH prepares to enable the same-network check
	// (NetworkCheck): on the desktop/parent prepare it is persisted on the
	// session, and on the companion/child prepare (carrying ParentSessionID) it
	// is compared against the parent's stored IP. Derive the IP server-side --
	// never trust a client-supplied value.
	NetworkOrigin *NetworkOrigin `json:"network_origin,omitempty"`

	// MobileCompanionPassthrough is opaque JSON forwarded to the branded mobile
	// companion app and rendered there. Object only, <=16KB.
	MobileCompanionPassthrough json.RawMessage `json:"mobile_companion_passthrough,omitempty"`

	// CompletionRedirectURL is where to send the browser once authentication
	// completes. It selects one of the completion_redirect_urls registered on
	// your application; matching is exact per RFC 6749 3.1.2 and an unregistered
	// value is rejected.
	//
	// Stamp it server-side from your own deployment's configured origin -- never
	// from the request Host header or any other client-controllable input. If
	// your application registers more than one completion URL, always send it:
	// when omitted the aggregator falls back to the first registered URL, so the
	// browser can land on another deployment's completion page.
	CompletionRedirectURL string `json:"completion_redirect_url,omitempty"`
}

// NetworkOrigin is the server-observed network origin of a prepare request.
//
// The IP MUST be derived by your backend from the connection it actually sees
// (e.g. the left-most untrusted hop of a trusted X-Forwarded-For chain) -- never
// from a value the end-user's client supplies in the request body or a header it
// controls. The aggregator trusts this value as-is and uses it only to compute
// the same-network NetworkCheck.
type NetworkOrigin struct {
	// IP is the server-observed IP address (IPv4 or IPv6) of the requesting device.
	IP string `json:"ip,omitempty"`
}

// NetworkCheck is the egress-closeness comparison between the desktop (parent)
// and companion (child) server-observed IPs. Returned on the companion/child
// prepare response and on the parent's verify-phone-number and get-phone-number,
// only when the desktop prepare stamped NetworkOrigin.
//
// It is a SIGNAL for how close the two devices' public egress IPs are as seen by
// the server -- NOT proof of the same physical LAN (an identical egress IP can
// front an entire NAT/CGNAT or a whole city) -- and must never be used as a hard
// gate. The caller owns any threshold on PrefixBits.
type NetworkCheck struct {
	// Match is "same_ip", "same_subnet", "different" or "unknown". Branch on it;
	// on "unknown" Family and PrefixBits are absent.
	Match string `json:"match"`
	// Family is the IP family the comparison was performed in: "ipv4" or "ipv6".
	Family string `json:"family,omitempty"`
	// PrefixBits is the number of matching leading bits between the two egress IPs
	// (0..32 for ipv4, 0..128 for ipv6). Higher = closer egress; the maximum
	// (32 / 128) means an identical egress IP. A pointer because 0 is a real
	// reading (nothing in common), distinct from nil on "unknown".
	PrefixBits *int `json:"prefix_bits,omitempty"`
	// DesktopIP and CompanionIP are the egress addresses compared, for your own risk engine.
	DesktopIP   string `json:"desktop_ip,omitempty"`
	CompanionIP string `json:"companion_ip,omitempty"`
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
	Swapped     *bool                `json:"swapped,omitempty"`
	MaxAgeHours int                  `json:"max_age_hours,omitempty"`
	Reason      SimSwapFailureReason `json:"reason,omitempty"`
}

// DeviceInfo contains information about the device that scanned the QR code.
type DeviceInfo struct {
	Type    string `json:"type"`              // Required: Device type/model
	OS      string `json:"os"`                // Required: Operating system and version
	Browser string `json:"browser,omitempty"` // Optional: Browser name and version
}

// EligibilityRequest represents a request to check authentication eligibility.
// ClientInfo is REQUIRED (with a non-blank UserAgent) and enforced by
// validateEligibilityRequest before the request is sent — the pointer type lets
// the SDK distinguish "omitted" from an empty struct and return a clear
// ValidationError rather than an avoidable server 400.
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
	// CarrierName is the normalized carrier name for the resolved network.
	// Present only for callers the gateway has explicitly enabled to receive
	// carrier information (same per-app gate as /prepare); omitted otherwise and
	// whenever Eligible is false.
	CarrierName string `json:"carrier_name,omitempty"`
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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

	// NetworkCheck is the egress-proximity comparison between the desktop
	// (parent) and companion (child) devices. Present only on the companion/child
	// prepare response (the request carrying Options.ParentSessionID), and only
	// when the desktop/parent prepare opted in by stamping Options.NetworkOrigin.
	// Nil for flows that do not opt in.
	NetworkCheck *NetworkCheck `json:"network_check,omitempty"`
	// MobileCompanionPassthrough echoes the opaque passthrough JSON stored on
	// the parent session, returned on the child (companion) prepare. Nil for
	// non-companion flows.
	MobileCompanionPassthrough json.RawMessage `json:"mobile_companion_passthrough,omitempty"`

	// CarrierName is the normalized carrier name for the resolved network.
	// Present only for callers the gateway has explicitly enabled to receive
	// carrier information (the same per-app gate as /eligibility). When gated
	// off, the field is omitted on the wire (omitempty), so it unmarshals to the
	// empty string here — preserving the privacy-by-default contract.
	CarrierName string `json:"carrier_name,omitempty"`

	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it. The session this creates is Session.SessionKey.
	RequestID string `json:"request_id,omitempty"`
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

// BiometricVerification is the device biometric outcome, shared by the verify
// and get-phone responses and the verification record's biometric_verification
// claim. A value outside the constants below is kept as sent.
type BiometricVerification string

const (
	BiometricVerificationVerified     BiometricVerification = "VERIFIED"
	BiometricVerificationFailed       BiometricVerification = "FAILED"
	BiometricVerificationNotEnrolled  BiometricVerification = "NOT_ENROLLED"
	BiometricVerificationUnsupported  BiometricVerification = "UNSUPPORTED"
	BiometricVerificationNotRequested BiometricVerification = "NOT_REQUESTED"
)

// VerifyPhoneNumberResponse is the response from phone number verification.
type VerifyPhoneNumberResponse struct {
	PhoneNumber string `json:"phone_number"`
	Verified    bool   `json:"verified"`
	// Nonce is the session's prepare nonce, equal to the record's nonce claim.
	// Empty from servers older than the release that added it.
	Nonce string `json:"nonce,omitempty"`
	Aud   string `json:"aud,omitempty"`
	// BiometricVerification equals the record's biometric_verification claim.
	BiometricVerification BiometricVerification `json:"biometric_verification"`
	SimSwap               *SimSwapInfo          `json:"sim_swap,omitempty"`
	DeviceSwap            *DeviceSwapInfo       `json:"device_swap,omitempty"`
	NetworkCheck          *NetworkCheck         `json:"network_check,omitempty"`
	PhoneNumberDetails    *PhoneNumberDetails   `json:"phone_number_details,omitempty"`
	// VerificationRecord is the Glide-signed compact JWS of this result. Check it
	// with Client.VerifyVerificationRecord before trusting it.
	VerificationRecord string `json:"verification_record,omitempty"`
	// VerificationRecordStatus is VerificationRecordStatusSigned or
	// VerificationRecordStatusUnavailable; empty from servers without records.
	VerificationRecordStatus string `json:"verification_record_status,omitempty"`
	// SessionKey is the session this response is about (the key prepare
	// returned), so a result can be stored or logged without the request.
	// Empty from servers older than the release that added it.
	SessionKey string `json:"session_key,omitempty"`
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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
	PhoneNumber string `json:"phone_number"`
	// Nonce is the session's prepare nonce. Empty from servers older than the
	// release that added it.
	Nonce                 string                `json:"nonce,omitempty"`
	Aud                   string                `json:"aud,omitempty"`
	BiometricVerification BiometricVerification `json:"biometric_verification"`
	SimSwap               *SimSwapInfo          `json:"sim_swap,omitempty"`
	DeviceSwap            *DeviceSwapInfo       `json:"device_swap,omitempty"`
	NetworkCheck          *NetworkCheck         `json:"network_check,omitempty"`
	PhoneNumberDetails    *PhoneNumberDetails   `json:"phone_number_details,omitempty"`
	// SessionKey is the session this response is about (the key prepare
	// returned), so a result can be stored or logged without the request.
	// Empty from servers older than the release that added it.
	SessionKey string `json:"session_key,omitempty"`
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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
	// SessionKey is the session this response is about (the key prepare
	// returned), so a result can be stored or logged without the request.
	// Empty from servers older than the release that added it.
	SessionKey string `json:"session_key,omitempty"`
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
}

// DeviceSwapInfo contains device swap (IMEI change) check results.
// Parallel to SimSwapInfo; only present when the carrier supports device swap detection.
type DeviceSwapInfo struct {
	Checked     bool      `json:"checked"`
	RiskLevel   RiskLevel `json:"risk_level,omitempty"`
	AgeBand     string    `json:"age_band,omitempty"`
	CarrierName string    `json:"carrier_name,omitempty"`
	CheckedAt   string    `json:"checked_at,omitempty"`
	Swapped     *bool     `json:"swapped,omitempty"`
	MaxAgeHours int       `json:"max_age_hours,omitempty"`
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
	// SessionKey is the session this response is about (the key prepare
	// returned), so a result can be stored or logged without the request.
	// Empty from servers older than the release that added it.
	SessionKey string `json:"session_key,omitempty"`
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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
	// RequestID is the id of this request, the same value as the X-Request-ID
	// response header. Quote it to Glide support. Empty from servers older than
	// the release that added it.
	RequestID string `json:"request_id,omitempty"`
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
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Status     int            `json:"status,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	SessionKey string         `json:"session_key,omitempty"`
	Timestamp  string         `json:"timestamp,omitempty"`
	TraceID    string         `json:"trace_id,omitempty"`
	SpanID     string         `json:"span_id,omitempty"`
	Service    string         `json:"service,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

// oauthError is the OAuth2 error response.
type oauthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
