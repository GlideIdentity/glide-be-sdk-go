package magicalauth

import (
	"context"
	"strings"
	"time"
)

// Values of VerifyPhoneNumberResponse.VerificationRecordStatus.
const (
	VerificationRecordStatusSigned      = "signed"
	VerificationRecordStatusUnavailable = "unavailable"
)

// VerificationRecordOptions tunes VerifyVerificationRecord. A nil value uses the defaults.
type VerificationRecordOptions struct {
	// ExpectedNonce, when set, must equal the record's nonce (the prepare nonce).
	ExpectedNonce string
	// AllowExpired accepts a record past exp, for audit of stored records.
	// The record's iat must still be within MaxAuditAge.
	AllowExpired bool
	// MaxAuditAge bounds iat when AllowExpired is set. Zero or negative means 90 days.
	MaxAuditAge time.Duration
	// Issuer overrides the expected iss. Empty means the client base URL without trailing "/".
	Issuer string
	// RevokedKIDs are rejected in addition to the SDK's built-in revoked list.
	RevokedKIDs []string
	// ExpectedAudience is the record aud to require. Empty means this client's ID.
	ExpectedAudience string
	// ExpectedOrigin, when non-empty, requires the record's origin to equal one entry exactly.
	ExpectedOrigin []string
	// Now overrides the clock. Nil means time.Now.
	Now func() time.Time
}

// VerificationRecordSignOff is one entry of the record's sign_offs claim.
type VerificationRecordSignOff struct {
	Role   string `json:"role"`
	Signer string `json:"signer"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Format string `json:"format,omitempty"`
	Token  string `json:"token,omitempty"`
}

// VerificationRecordCarrier is the optional carrier claim on a verified record.
type VerificationRecordCarrier struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	MVNO string `json:"mvno,omitempty"`
}

// VerificationRecordAgeBand is a swap age range in hours; -1 means unknown.
type VerificationRecordAgeBand struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// VerificationRecordSwap is the sim_swap or device_swap claim on a verified record.
type VerificationRecordSwap struct {
	Checked     bool                       `json:"checked"`
	RiskLevel   string                     `json:"risk_level,omitempty"`
	AgeBand     *VerificationRecordAgeBand `json:"age_band,omitempty"`
	CarrierName string                     `json:"carrier_name,omitempty"`
	Reason      string                     `json:"reason,omitempty"`
	Swapped     *bool                      `json:"swapped,omitempty"`
	MaxAgeHours int                        `json:"max_age_hours,omitempty"`
}

// VerificationRecordClaims are the claims of a verified record.
type VerificationRecordClaims struct {
	Ver                   string                      `json:"ver"`
	Iss                   string                      `json:"iss"`
	Sub                   string                      `json:"sub"`
	Aud                   string                      `json:"aud"`
	Iat                   int64                       `json:"iat"`
	Nbf                   int64                       `json:"nbf"`
	Exp                   int64                       `json:"exp"`
	Jti                   string                      `json:"jti"`
	Nonce                 string                      `json:"nonce"`
	Origin                string                      `json:"origin"`
	UseCase               string                      `json:"use_case"`
	Verified              bool                        `json:"verified"`
	Strategy              string                      `json:"strategy,omitempty"`
	PhoneNumberDetails    *PhoneNumberDetails         `json:"phone_number_details,omitempty"`
	Carrier               *VerificationRecordCarrier  `json:"carrier,omitempty"`
	BiometricVerification BiometricVerification       `json:"biometric_verification"`
	SimSwap               *VerificationRecordSwap     `json:"sim_swap,omitempty"`
	DeviceSwap            *VerificationRecordSwap     `json:"device_swap,omitempty"`
	SignOffs              []VerificationRecordSignOff `json:"sign_offs"`
}

// SignOffResult reports how the SDK treated one sign-off. Result is
// "unavailable" (with the server's Reason) or "unverified": a present
// sign-off is never treated as carrier trust in this version.
type SignOffResult struct {
	Role   string `json:"role"`
	Signer string `json:"signer"`
	Result string `json:"result"`
	Reason string `json:"reason,omitempty"`
}

// VerifiedVerificationRecord is a record whose signature and claims passed every check.
type VerifiedVerificationRecord struct {
	Claims         VerificationRecordClaims
	SignOffResults []SignOffResult
}

// VerifyVerificationRecord verifies a verification record returned by
// VerifyPhoneNumber: the ES256 signature against Glide's published record keys,
// then issuer, audience (this client's ID by default), validity window, shape, nonce and origin (when ExpectedOrigin is set).
// An empty record fails with RECORD_UNAVAILABLE, so a caller that requires
// records can pass VerifyPhoneNumberResponse.VerificationRecord straight in.
// Every failure is a *VerificationRecordError.
func (c *Client) VerifyVerificationRecord(ctx context.Context, record string, opts *VerificationRecordOptions) (*VerifiedVerificationRecord, error) {
	if opts == nil {
		opts = &VerificationRecordOptions{}
	}
	now := time.Now()
	if opts.Now != nil {
		now = opts.Now()
	}

	parsed, err := parseRecord(record)
	if err != nil {
		return nil, err
	}
	if err := checkRecordKID(parsed.kid, opts.RevokedKIDs); err != nil {
		return nil, err
	}
	key, err := c.records.key(ctx, parsed.kid)
	if err != nil {
		return nil, err
	}
	if err := verifyRecordSignature(parsed, key); err != nil {
		return nil, err
	}
	claims, err := checkRecordClaims(parsed.payload, c.recordExpectations(opts, now))
	if err != nil {
		return nil, err
	}
	return &VerifiedVerificationRecord{Claims: *claims, SignOffResults: signOffResults(claims.SignOffs)}, nil
}

func (c *Client) recordExpectations(opts *VerificationRecordOptions, now time.Time) recordExpectations {
	issuer := opts.Issuer
	if issuer == "" {
		issuer = strings.TrimRight(c.baseURL, "/")
	}
	maxAuditAge := opts.MaxAuditAge
	if maxAuditAge <= 0 {
		maxAuditAge = defaultMaxAuditAge
	}
	audience := opts.ExpectedAudience
	if audience == "" {
		audience = c.clientID
	}
	return recordExpectations{
		issuer:       issuer,
		audience:     audience,
		nonce:        opts.ExpectedNonce,
		origins:      opts.ExpectedOrigin,
		allowExpired: opts.AllowExpired,
		maxAuditAge:  maxAuditAge,
		now:          now,
	}
}
