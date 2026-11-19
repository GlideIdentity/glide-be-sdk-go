package magicalauth

import (
	"encoding/json"
	"strings"
)

const (
	recordUseCase       = "VerifyPhoneNumber"
	recordVersion       = "1.0"
	recordMaxLifetime   = 60 // seconds from iat to exp
	signOffRoleCarrier  = "carrier"
	carrierSignerPrefix = "carrier:"
)

// checkClaimShape enforces the v1 claim contract in its fixed order. iat and
// exp are already bounded by the time checks, so iat+n cannot overflow.
func checkClaimShape(obj recordObject, iat, exp int64) ([]VerificationRecordSignOff, []string, error) {
	if _, ok := jsonBool(obj["verified"]); !ok {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "verified must be a boolean")
	}
	for _, name := range []string{"jti", "nonce", "sub"} {
		if s, ok := jsonString(obj[name]); !ok || s == "" {
			return nil, nil, recordError(VerificationRecordCodeMalformed, name+" must be a non-empty string")
		}
	}
	if sub, _ := jsonString(obj["sub"]); !strings.HasPrefix(sub, "tel:+") {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "sub must be tel:+E164")
	}
	for _, name := range []string{"origin", "biometric_verification"} {
		if _, ok := jsonString(obj[name]); !ok {
			return nil, nil, recordError(VerificationRecordCodeMalformed, name+" must be a string")
		}
	}
	if nbf, ok := jsonInt(obj["nbf"]); !ok || nbf != iat {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "nbf must equal iat")
	}
	if s, ok := jsonString(obj["use_case"]); !ok || s != recordUseCase {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "use_case must be "+recordUseCase)
	}
	if exp < iat+1 || exp > iat+recordMaxLifetime {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "exp must be 1 to 60 seconds after iat")
	}
	signOffs, ok := readSignOffs(obj["sign_offs"])
	if !ok {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "sign_offs must be an array of well-formed entries")
	}
	factors, ok := readFactors(obj)
	if !ok {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "factors must be an array of strings")
	}
	if len(factors) == 1 && factors[0] == "passkey" {
		if hasCarrierSignOff(signOffs) {
			return nil, nil, recordError(VerificationRecordCodeMalformed, "a passkey-only record must have no carrier sign-off")
		}
		return signOffs, factors, nil
	}
	if !oneCarrierSignOff(signOffs) {
		return nil, nil, recordError(VerificationRecordCodeMalformed, "sign_offs must have exactly one carrier:<id> carrier entry")
	}
	return signOffs, factors, nil
}

func readFactors(obj recordObject) ([]string, bool) {
	raw, present := obj["factors"]
	if !present {
		return []string{"carrier"}, true
	}
	var entries []json.RawMessage
	if !isJSONArray(raw) || json.Unmarshal(raw, &entries) != nil {
		return nil, false
	}
	factors := make([]string, 0, len(entries))
	for _, e := range entries {
		s, ok := jsonString(e)
		if !ok {
			return nil, false
		}
		factors = append(factors, s)
	}
	return factors, true
}

func hasCarrierSignOff(signOffs []VerificationRecordSignOff) bool {
	for _, s := range signOffs {
		if s.Role == signOffRoleCarrier {
			return true
		}
	}
	return false
}

// readSignOffs requires a JSON array of well-formed sign-off objects.
func readSignOffs(raw json.RawMessage) ([]VerificationRecordSignOff, bool) {
	var entries []json.RawMessage
	if !isJSONArray(raw) || json.Unmarshal(raw, &entries) != nil {
		return nil, false
	}
	signOffs := make([]VerificationRecordSignOff, 0, len(entries))
	for _, entry := range entries {
		s, ok := readSignOff(entry)
		if !ok {
			return nil, false
		}
		signOffs = append(signOffs, s)
	}
	return signOffs, true
}

func readSignOff(entry json.RawMessage) (VerificationRecordSignOff, bool) {
	obj, ok := parseRecordObject(entry)
	if !ok {
		return VerificationRecordSignOff{}, false
	}
	r := newMemberReader(obj)
	s := VerificationRecordSignOff{
		Role: r.str("role"), Signer: r.str("signer"), Status: r.str("status"),
		Reason: r.str("reason"), Format: r.str("format"), Token: r.str("token"),
	}
	if !r.ok || s.Role == "" || s.Signer == "" {
		return s, false
	}
	switch s.Status {
	case VerificationRecordStatusUnavailable:
		return s, s.Reason != ""
	case "present":
		return s, s.Format != "" && s.Token != ""
	}
	return s, false
}

func oneCarrierSignOff(signOffs []VerificationRecordSignOff) bool {
	carriers := 0
	for _, s := range signOffs {
		if s.Role != signOffRoleCarrier {
			continue
		}
		carriers++
		if !strings.HasPrefix(s.Signer, carrierSignerPrefix) || len(s.Signer) == len(carrierSignerPrefix) {
			return false
		}
	}
	return carriers == 1
}

func carrierIDMatchesSignOff(id string, signOffs []VerificationRecordSignOff) bool {
	for _, s := range signOffs {
		if s.Role == signOffRoleCarrier {
			return s.Signer == carrierSignerPrefix+id
		}
	}
	return false
}
