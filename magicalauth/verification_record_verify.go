package magicalauth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"slices"
	"strings"
)

const (
	recordAlg      = "ES256"
	recordTyp      = "glide-verification+jwt"
	maxRecordBytes = 16 << 10
)

// recordKIDs are the only kids a record may be signed with. Other keys in the
// same JWKS (invocation signing, encryption) must never verify a record.
var recordKIDs = []string{
	"glide-record-signing-key-ecdsa-es256-us-2026",
	"glide-record-signing-key-ecdsa-es256-eu-2026",
}

// revokedRecordKIDs is the built-in revoked list, extended per call by
// VerificationRecordOptions.RevokedKIDs.
var revokedRecordKIDs = []string{}

type parsedRecord struct {
	kid          string
	signingInput []byte
	payload      []byte
	signature    []byte
}

// parseRecord splits and decodes the compact JWS and validates the protected
// header, before any key lookup or crypto.
func parseRecord(record string) (*parsedRecord, error) {
	if record == "" {
		return nil, recordError(VerificationRecordCodeRecordUnavailable, "no verification record")
	}
	if len(record) > maxRecordBytes {
		return nil, recordError(VerificationRecordCodeMalformed, "record is larger than 16 KiB")
	}
	segments := strings.Split(record, ".")
	if len(segments) != 3 {
		return nil, recordError(VerificationRecordCodeMalformed, "record is not a compact JWS")
	}
	decoded := make([][]byte, 3)
	for i, segment := range segments {
		b, err := decodeBase64URL(segment)
		if err != nil {
			return nil, recordError(VerificationRecordCodeMalformed, "record segment is not base64url")
		}
		decoded[i] = b
	}
	var header map[string]json.RawMessage
	if err := json.Unmarshal(decoded[0], &header); err != nil || header == nil {
		return nil, recordError(VerificationRecordCodeMalformed, "header is not a JSON object")
	}
	kid, err := checkRecordHeader(header)
	if err != nil {
		return nil, err
	}
	return &parsedRecord{
		kid:          kid,
		signingInput: []byte(segments[0] + "." + segments[1]),
		payload:      decoded[1],
		signature:    decoded[2],
	}, nil
}

func checkRecordHeader(header map[string]json.RawMessage) (string, error) {
	alg, algOK := jsonString(header["alg"])
	typ, typOK := jsonString(header["typ"])
	kid, kidOK := jsonString(header["kid"])
	if len(header) != 3 || !algOK || !typOK || !kidOK {
		return "", recordError(VerificationRecordCodeDisallowedHeader, "header must contain exactly alg, typ and kid")
	}
	if alg != recordAlg || typ != recordTyp {
		return "", recordError(VerificationRecordCodeDisallowedHeader, "header alg/typ not allowed")
	}
	return kid, nil
}

func checkRecordKID(kid string, revoked []string) error {
	if slices.Contains(revokedRecordKIDs, kid) || slices.Contains(revoked, kid) {
		return recordError(VerificationRecordCodeRevokedKey, "kid "+kid+" is revoked")
	}
	if !slices.Contains(recordKIDs, kid) {
		return recordError(VerificationRecordCodeUnknownKey, "kid "+kid+" is not a record signing key")
	}
	return nil
}

func verifyRecordSignature(record *parsedRecord, key *ecdsa.PublicKey) error {
	if len(record.signature) != 64 {
		return recordError(VerificationRecordCodeBadSignature, "signature is not 64 bytes")
	}
	r := new(big.Int).SetBytes(record.signature[:32])
	s := new(big.Int).SetBytes(record.signature[32:])
	digest := sha256.Sum256(record.signingInput)
	if !ecdsa.Verify(key, digest[:], r, s) {
		return recordError(VerificationRecordCodeBadSignature, "signature does not verify")
	}
	return nil
}

var strictBase64URL = base64.RawURLEncoding.Strict()

// decodeBase64URL rejects padding, non-zero trailing bits and any byte outside
// [A-Za-z0-9_-]; the alphabet check is needed because Go's decoder skips CR/LF.
func decodeBase64URL(segment string) ([]byte, error) {
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return nil, errors.New("invalid base64url character")
		}
	}
	return strictBase64URL.DecodeString(segment)
}

// jsonString returns the value only when raw is a JSON string (not null, number, ...).
func jsonString(raw json.RawMessage) (string, bool) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '"' {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}
