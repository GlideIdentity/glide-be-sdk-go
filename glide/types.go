package glide

import (
	"time"
)

// PrepareRequest initiates the authentication flow
type PrepareRequest struct {
	// PhoneNumber in E.164 format (optional if using PLMN)
	PhoneNumber string `json:"phone_number,omitempty"`

	// PLMN (Public Land Mobile Network) identifier
	PLMN *PLMN `json:"plmn,omitempty"`

	// UseCase for the authentication
	UseCase UseCase `json:"use_case"`

	// ConsentData for user consent (optional)
	ConsentData *ConsentData `json:"consent_data,omitempty"`

	// ClientInfo contains client information like user agent
	ClientInfo *ClientInfo `json:"client_info,omitempty"`
}

// PLMN represents the carrier network identifier
type PLMN struct {
	MCC string `json:"mcc"` // Mobile Country Code
	MNC string `json:"mnc"` // Mobile Network Code
}

// ConsentData contains user consent information
type ConsentData struct {
	ConsentText string `json:"consent_text"`
	PolicyLink  string `json:"policy_link"`
	PolicyText  string `json:"policy_text,omitempty"`
}

// ClientInfo contains client information for strategy selection
type ClientInfo struct {
	UserAgent string `json:"user_agent,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

// PrepareResponse contains the authentication preparation result
// SessionInfo contains session information for authentication flow
type SessionInfo struct {
	SessionKey string `json:"session_key"`
	Nonce      string `json:"nonce"`
	EncKey     string `json:"enc_key"`
}

type PrepareResponse struct {
	// AuthenticationStrategy indicates the authentication method (ts43 or link)
	AuthenticationStrategy AuthenticationStrategy `json:"authentication_strategy"`

	// Session information for this authentication flow
	Session SessionInfo `json:"session"`

	// Data contains strategy-specific information
	Data map[string]interface{} `json:"data"`

	// TTL session time-to-live in seconds (optional)
	TTL int `json:"ttl,omitempty"`

	// UseCase that was prepared (needed to know which endpoint to call in process)
	// This is a client-side field, not from the server
	UseCase UseCase `json:"-"` // Omit from JSON marshaling
}

// VerifyPhoneNumberRequest requests phone number verification
type VerifyPhoneNumberRequest struct {
	// SessionInfo from the prepare response (includes session_key, nonce, enc_key)
	SessionInfo *SessionInfo `json:"session"`

	// Credential from the Digital Credentials API (vp_token)
	Credential map[string]interface{} `json:"response"`
}

// VerifyPhoneNumberResponse contains the verification result
type VerifyPhoneNumberResponse struct {
	// PhoneNumber that was verified
	PhoneNumber string `json:"phone_number"`

	// Verified indicates if the phone number was successfully verified
	Verified bool `json:"verified"`
}

// GetPhoneNumberRequest requests phone number retrieval
type GetPhoneNumberRequest struct {
	// SessionInfo from the prepare response (includes session_key, nonce, enc_key)
	SessionInfo *SessionInfo `json:"session"`

	// Credential from the Digital Credentials API (vp_token)
	Credential map[string]interface{} `json:"response"`
}

// GetPhoneNumberResponse contains the retrieved phone number
type GetPhoneNumberResponse struct {
	// PhoneNumber retrieved from the carrier
	PhoneNumber string `json:"phone_number"`
}

// SimSwapCheckRequest checks for recent SIM swaps
type SimSwapCheckRequest struct {
	PhoneNumber string `json:"phone_number"`
	MaxAge      int    `json:"max_age_hours,omitempty"` // Default 24 hours
}

// SimSwapCheckResponse contains the SIM swap check result
type SimSwapCheckResponse struct {
	Swapped   bool       `json:"swapped"`
	SwappedAt *time.Time `json:"swapped_at,omitempty"`
	CheckedAt time.Time  `json:"checked_at"`
}

// SimSwapDateRequest retrieves the last SIM swap date
type SimSwapDateRequest struct {
	PhoneNumber string `json:"phone_number"`
}

// SimSwapDateResponse contains the last SIM swap date
type SimSwapDateResponse struct {
	LastSwapDate *time.Time `json:"last_swap_date,omitempty"`
	CheckedAt    time.Time  `json:"checked_at"`
}

// NumberVerifyRequest verifies phone number ownership
type NumberVerifyRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code,omitempty"` // For code-based verification
}

// NumberVerifyResponse contains the verification result
type NumberVerifyResponse struct {
	Verified  bool      `json:"verified"`
	CheckedAt time.Time `json:"checked_at"`
}

// KYCMatchRequest contains user information to verify
type KYCMatchRequest struct {
	PhoneNumber string   `json:"phone_number"`
	Name        string   `json:"name,omitempty"`
	GivenName   string   `json:"given_name,omitempty"`
	FamilyName  string   `json:"family_name,omitempty"`
	BirthDate   string   `json:"birth_date,omitempty"` // Format: YYYY-MM-DD
	Email       string   `json:"email,omitempty"`
	Address     *Address `json:"address,omitempty"`
	IDDocument  string   `json:"id_document,omitempty"`
}

// Address represents a physical address
type Address struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// KYCMatchResponse contains the KYC verification result
type KYCMatchResponse struct {
	MatchResults map[string]MatchResult `json:"match_results"`
	OverallMatch bool                   `json:"overall_match"`
	CheckedAt    time.Time              `json:"checked_at"`
}

// MatchResult represents the result of a single field match
type MatchResult struct {
	Matched    bool   `json:"matched"`
	Confidence string `json:"confidence,omitempty"` // high, medium, low
}
