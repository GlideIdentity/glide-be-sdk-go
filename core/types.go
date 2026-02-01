package core

// ============================================================
// Common Types
// ============================================================

// PLMN represents the Public Land Mobile Network identifiers.
type PLMN struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

// ConsentData contains user consent information.
type ConsentData struct {
	ConsentText string `json:"consent_text"`
	PolicyLink  string `json:"policy_link"`
	PolicyText  string `json:"policy_text,omitempty"`
}

// ClientInfo contains client information for strategy selection.
type ClientInfo struct {
	UserAgent string `json:"user_agent,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

// SessionInfo contains session information for authentication flow.
type SessionInfo struct {
	SessionKey string            `json:"session_key"`
	Nonce      string            `json:"nonce,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Challenge contains visual verification data for desktop QR authentication.
type Challenge struct {
	Pattern   string `json:"pattern"`
	Color     string `json:"color"`
	SessionID string `json:"session_id"`
}

// ============================================================
// Strategy-specific Data Types
// ============================================================

// TS43CredentialMeta contains metadata for TS43 Digital Credentials.
type TS43CredentialMeta struct {
	VCTValues                  []string `json:"vct_values"`
	CredentialAuthorizationJWT string   `json:"credential_authorization_jwt"`
}

// TS43Credential contains a single credential request for TS43.
type TS43Credential struct {
	ID     string             `json:"id"`
	Format string             `json:"format"`
	Meta   TS43CredentialMeta `json:"meta"`
	Claims []string           `json:"claims,omitempty"`
}

// TS43DCQLQuery contains the DCQL query for TS43.
type TS43DCQLQuery struct {
	Credentials []TS43Credential `json:"credentials"`
}

// TS43InnerData contains the protocol-specific data for TS43.
type TS43InnerData struct {
	Nonce        string        `json:"nonce"`
	ResponseMode string        `json:"response_mode"`
	ResponseType string        `json:"response_type"`
	DCQLQuery    TS43DCQLQuery `json:"dcql_query"`
}

// TS43Data contains TS43 strategy response data.
// Used for Android Digital Credentials API (Chrome 128+).
type TS43Data struct {
	Protocol string        `json:"protocol"`
	Data     TS43InnerData `json:"data"`
}

// LinkData contains Link strategy response data.
// Used for iOS App Clips and OAuth redirects.
type LinkData struct {
	URL       string            `json:"url"`
	ReturnURL string            `json:"return_url,omitempty"`
	StatusURL string            `json:"status_url,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

// DesktopInnerData contains QR codes and URLs for desktop authentication.
type DesktopInnerData struct {
	SessionID      string     `json:"session_id,omitempty"`
	IOSQRImage     string     `json:"ios_qr_image,omitempty"`
	AndroidQRImage string     `json:"android_qr_image,omitempty"`
	IOSURL         string     `json:"ios_url,omitempty"`
	AndroidURL     string     `json:"android_url,omitempty"`
	StatusURL      string     `json:"status_url,omitempty"`
	Challenge      *Challenge `json:"challenge,omitempty"`
}

// DesktopData contains Desktop strategy response data.
// Used for QR code-based authentication from desktop browsers.
type DesktopData struct {
	Protocol string            `json:"protocol,omitempty"`
	Data     *DesktopInnerData `json:"data,omitempty"`
}

// ============================================================
// Request Types
// ============================================================

// PrepareRequest initiates the authentication flow.
// Tip: Call ahead of user interaction to minimize wait time during authentication.
type PrepareRequest struct {
	PhoneNumber string          `json:"phone_number,omitempty"`
	PLMN        *PLMN           `json:"plmn,omitempty"`
	UseCase     UseCase         `json:"use_case,omitempty"`
	ConsentData *ConsentData    `json:"consent_data,omitempty"`
	ClientInfo  *ClientInfo     `json:"client_info,omitempty"`
	Nonce       string          `json:"nonce,omitempty"`
	Options     *PrepareOptions `json:"options,omitempty"`
}

// PrepareOptions contains optional configuration for special features.
type PrepareOptions struct {
	ParentSessionID string `json:"parent_session_id,omitempty"`
	Theme           string `json:"theme,omitempty"`
}

// VerifyPhoneNumberRequest requests phone number verification.
type VerifyPhoneNumberRequest struct {
	Session    SessionInfo `json:"session"`
	Credential string      `json:"credential"`
}

// GetPhoneNumberRequest requests phone number retrieval.
type GetPhoneNumberRequest struct {
	Session    SessionInfo `json:"session"`
	Credential string      `json:"credential"`
}

// ProcessRequest is the unified FE-to-BE SDK request for credential processing.
// The FE SDK sends this to the /process endpoint with use_case to determine routing.
type ProcessRequest struct {
	Credential string      `json:"credential"`
	Session    SessionInfo `json:"session"`
	UseCase    UseCase     `json:"use_case"`
}

// ReportInvocationRequest reports that an authentication flow was started.
type ReportInvocationRequest struct {
	SessionID string `json:"session_id"`
}

// ============================================================
// Response Types
// ============================================================

// PrepareResponse contains the authentication preparation result.
type PrepareResponse struct {
	AuthenticationStrategy AuthenticationStrategy `json:"authentication_strategy"`
	Session                SessionInfo            `json:"session"`
	Data                   map[string]interface{} `json:"data"`
	Challenge              *Challenge             `json:"challenge,omitempty"`
}

// VerifyPhoneNumberResponse contains the verification result.
type VerifyPhoneNumberResponse struct {
	PhoneNumber string `json:"phone_number"`
	Verified    bool   `json:"verified"`
	Aud         string `json:"aud,omitempty"`
}

// GetPhoneNumberResponse contains the retrieved phone number.
type GetPhoneNumberResponse struct {
	PhoneNumber string `json:"phone_number"`
	Aud         string `json:"aud,omitempty"`
}

// ReportInvocationResponse contains the result of reporting an invocation.
type ReportInvocationResponse struct {
	Success bool `json:"success"`
}
