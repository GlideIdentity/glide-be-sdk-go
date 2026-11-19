package magicalauth

import (
	"bytes"
	"encoding/json"
	"slices"
	"time"
)

const (
	defaultMaxAuditAge = 90 * 24 * time.Hour
	recordIatLeeway    = 60 // seconds an iat may sit in the future
)

type recordExpectations struct {
	issuer       string
	audience     string
	nonce        string
	origins      []string
	allowExpired bool
	maxAuditAge  time.Duration
	now          time.Time
}

// checkRecordClaims validates a signature-verified payload and returns claims
// built only from the exact-case members it checked. Type checks run on raw
// members, so an array aud reports WRONG_AUDIENCE rather than MALFORMED.
func checkRecordClaims(payload []byte, want recordExpectations) (*VerificationRecordClaims, error) {
	obj, ok := parseRecordObject(payload)
	if !ok {
		return nil, recordError(VerificationRecordCodeMalformed, "payload is not a JSON object")
	}
	if ver, ok := jsonString(obj["ver"]); !ok || ver != recordVersion {
		return nil, recordError(VerificationRecordCodeUnsupportedVersion, "ver must be \""+recordVersion+"\"")
	}
	iss, ok := jsonString(obj["iss"])
	if !ok || iss != want.issuer {
		return nil, recordError(VerificationRecordCodeWrongIssuer, "iss does not match "+want.issuer)
	}
	aud, ok := jsonString(obj["aud"])
	if !ok || aud != want.audience {
		return nil, recordError(VerificationRecordCodeWrongAudience, "aud does not match the expected audience")
	}
	iat, exp, err := checkRecordTimes(obj, want)
	if err != nil {
		return nil, err
	}
	signOffs, factors, err := checkClaimShape(obj, iat, exp)
	if err != nil {
		return nil, err
	}
	verified, _ := jsonBool(obj["verified"])
	r := newMemberReader(obj)
	claims := &VerificationRecordClaims{
		Ver: recordVersion, Iss: iss, Aud: aud, Iat: iat, Nbf: iat, Exp: exp, Verified: verified, SignOffs: signOffs,
		Factors: factors, ActionHash: r.str("action_hash"),
		Jti: r.str("jti"), Nonce: r.str("nonce"), UseCase: r.str("use_case"), Sub: r.str("sub"),
		Origin: r.str("origin"), Strategy: r.str("strategy"),
		PhoneNumberDetails: readPhoneNumberDetails(r), Carrier: readCarrier(r), BiometricVerification: BiometricVerification(r.str("biometric_verification")),
		SimSwap: readSwap(r, "sim_swap"), DeviceSwap: readSwap(r, "device_swap"),
		LiveCall: readLiveCall(r, "live_call"),
	}
	if !r.ok {
		return nil, recordError(VerificationRecordCodeMalformed, "claims have the wrong JSON types")
	}
	if claims.Carrier != nil && !carrierIDMatchesSignOff(claims.Carrier.ID, signOffs) {
		return nil, recordError(VerificationRecordCodeMalformed, "carrier.id must match the carrier sign-off signer")
	}
	if want.nonce != "" && claims.Nonce != want.nonce {
		return nil, recordError(VerificationRecordCodeNonceMismatch, "nonce does not match the expected nonce")
	}
	if len(want.origins) > 0 && !slices.Contains(want.origins, claims.Origin) {
		return nil, recordError(VerificationRecordCodeOriginMismatch, "origin is not an expected origin")
	}
	return claims, nil
}

func checkRecordTimes(obj recordObject, want recordExpectations) (iat, exp int64, err error) {
	iat, iatOK := jsonInt(obj["iat"])
	exp, expOK := jsonInt(obj["exp"])
	if !iatOK || !expOK {
		return 0, 0, recordError(VerificationRecordCodeMalformed, "iat and exp must be integers")
	}
	now := want.now.Unix()
	if iat > now+recordIatLeeway {
		return 0, 0, recordError(VerificationRecordCodeNotYetValid, "iat is in the future")
	}
	if exp > now {
		return iat, exp, nil
	}
	if !want.allowExpired {
		return 0, 0, recordError(VerificationRecordCodeExpired, "record has expired")
	}
	if iat < now-int64(want.maxAuditAge/time.Second) {
		return 0, 0, recordError(VerificationRecordCodeExpired, "record is older than the maximum audit age")
	}
	return iat, exp, nil
}

func signOffResults(signOffs []VerificationRecordSignOff) []SignOffResult {
	results := make([]SignOffResult, 0, len(signOffs))
	for _, s := range signOffs {
		if s.Status == VerificationRecordStatusUnavailable {
			results = append(results, SignOffResult{Role: s.Role, Signer: s.Signer, Result: "unavailable", Reason: s.Reason})
			continue
		}
		results = append(results, SignOffResult{Role: s.Role, Signer: s.Signer, Result: "unverified"})
	}
	return results
}

// isJSONArray is needed because json.Unmarshal accepts null into a slice.
func isJSONArray(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) > 0 && raw[0] == '['
}
