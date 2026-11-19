package magicalauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"slices"
	"strings"
	"time"
)

const defaultStepUpMaxAge = 5 * time.Minute

// ActionHash is the base64url SHA-256 of the canonical action: keys sorted by
// byte value, each "key=value", joined with "|".
func ActionHash(action map[string]string) string {
	keys := make([]string, 0, len(action))
	for k := range action {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + action[k]
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// VerifyStepUp verifies a step-up verification record and requires that it is
// verified, carries both the passkey and carrier factors, is bound to action
// and is for phoneNumber (E.164). maxAge bounds the record's iat; zero or
// negative means 5 minutes. expectedAudience is the record aud to require
// (first non-empty value); absent means this client's ID. A record that fails
// these checks returns STEP_UP_REJECTED; every failure is a *VerificationRecordError.
func (c *Client) VerifyStepUp(ctx context.Context, record string, action map[string]string, phoneNumber string, maxAge time.Duration, expectedAudience ...string) (*VerifiedVerificationRecord, error) {
	if maxAge <= 0 {
		maxAge = defaultStepUpMaxAge
	}
	opts := &VerificationRecordOptions{AllowExpired: true, MaxAuditAge: maxAge}
	for _, a := range expectedAudience {
		if a != "" {
			opts.ExpectedAudience = a
			break
		}
	}
	rec, err := c.VerifyVerificationRecord(ctx, record, opts)
	if err != nil {
		return nil, err
	}
	cl := rec.Claims
	switch {
	case time.Now().Unix()-cl.Iat > int64(maxAge/time.Second):
		return nil, recordError(VerificationRecordCodeExpired, "record is older than maxAge")
	case !cl.Verified:
		return nil, recordError(VerificationRecordCodeStepUpRejected, "record is not verified")
	case !slices.Contains(cl.Factors, "passkey") || !slices.Contains(cl.Factors, "carrier"):
		return nil, recordError(VerificationRecordCodeStepUpRejected, "record lacks the passkey and carrier factors")
	case cl.ActionHash != ActionHash(action):
		return nil, recordError(VerificationRecordCodeStepUpRejected, "action_hash does not match the action")
	case cl.Sub != "tel:"+phoneNumber:
		return nil, recordError(VerificationRecordCodeStepUpRejected, "sub does not match the phone number")
	}
	return rec, nil
}
