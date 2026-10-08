package magicalauth

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

type recordJWK struct {
	Kty string
	Crv string
	Use string
	Alg string
	Kid string
	X   string
	Y   string
}

// parseJWKS reads a JWKS with the same exact-case member rules as records.
// Entries that aren't objects or carry wrong-typed members are skipped: they
// can never resolve, and one bad entry must not hide the others.
func parseJWKS(body []byte) ([]recordJWK, error) {
	doc, ok := parseRecordObject(body)
	if !ok {
		return nil, errors.New("JWKS is not a JSON object")
	}
	var entries []json.RawMessage
	if !isJSONArray(doc["keys"]) || json.Unmarshal(doc["keys"], &entries) != nil {
		return nil, errors.New("JWKS keys is not an array")
	}
	keys := make([]recordJWK, 0, len(entries))
	for _, entry := range entries {
		obj, ok := parseRecordObject(entry)
		if !ok {
			continue
		}
		r := newMemberReader(obj)
		k := recordJWK{Kty: r.str("kty"), Crv: r.str("crv"), Use: r.str("use"), Alg: r.str("alg"), Kid: r.str("kid"), X: r.str("x"), Y: r.str("y")}
		if r.ok {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (k recordJWK) publicKey() (*ecdsa.PublicKey, error) {
	if k.Kty != "EC" || k.Crv != "P-256" || k.Use != "sig" || k.Alg != recordAlg {
		return nil, fmt.Errorf("JWK is not an EC P-256 ES256 signing key")
	}
	x, errX := decodeBase64URL(k.X)
	y, errY := decodeBase64URL(k.Y)
	if errX != nil || errY != nil || len(x) != 32 || len(y) != 32 {
		return nil, fmt.Errorf("JWK coordinates are not 32-byte base64url")
	}
	// ecdh validates that the point is on the curve.
	if _, err := ecdh.P256().NewPublicKey(append(append([]byte{4}, x...), y...)); err != nil {
		return nil, fmt.Errorf("JWK point is not on P-256")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}, nil
}
