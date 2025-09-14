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

// PrepareResponse contains the authentication preparation result
type PrepareResponse struct {
	// Strategy indicates the authentication method (ts43 or link)
	Strategy string `json:"strategy"`

	// Session identifier for this authentication flow
	Session string `json:"session"`

	// Data contains strategy-specific information
	Data map[string]interface{} `json:"data"`

	// TTL session time-to-live in seconds (optional)
	TTL int `json:"ttl,omitempty"`
}

// ProcessRequest processes the authentication credential
type ProcessRequest struct {
	// Session from the prepare response
	Session string `json:"session"`

	// Response from the client-side authentication
	Response map[string]interface{} `json:"response"`

	// PhoneNumber for VerifyPhoneNumber use case
	PhoneNumber string `json:"phone_number,omitempty"`
}

// ProcessResponse contains the authentication result
type ProcessResponse struct {
	// PhoneNumber that was verified or retrieved
	PhoneNumber string `json:"phone_number"`

	// Verified indicates if the number was verified (for VerifyPhoneNumber)
	Verified bool `json:"verified,omitempty"`

	// Metadata contains additional information
	Metadata *ProcessMetadata `json:"metadata,omitempty"`
}

// ProcessMetadata contains additional process information
type ProcessMetadata struct {
	VerifiedAt time.Time `json:"verified_at"`
	SessionID  string    `json:"session_id,omitempty"`
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
