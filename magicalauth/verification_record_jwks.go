package magicalauth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultJWKSMaxAge   = 300 * time.Second
	jwksRefreshInterval = 60 * time.Second
	maxJWKSBytes        = 1 << 20
	// maxSafeInteger bounds max-age exactly as the Node and Java SDKs do.
	maxSafeInteger = 1<<53 - 1
)

const recordJWKSPath = "/.well-known/us/jwks.json"

// recordKeyCache holds the record signing keys from the configured base URL's
// JWKS document. URLs are fixed at construction, never taken from a token.
// It runs on its own clock: a caller's VerificationRecordOptions.Now must not
// be able to pin or age the shared cache.
type recordKeyCache struct {
	baseURL    string
	httpClient HTTPClient
	now        func() time.Time

	mu           sync.Mutex
	keys         map[string]recordJWK
	expires      time.Time
	lastRefresh  time.Time
	backoffUntil time.Time
}

func newRecordKeyCache(baseURL string, httpClient HTTPClient) *recordKeyCache {
	return &recordKeyCache{baseURL: baseURL, httpClient: httpClient, now: time.Now}
}

// key resolves kid, refetching when the cache has expired, or when kid is
// missing and the last refresh is at least jwksRefreshInterval old. After a
// failed refresh the previous keys are kept and further refreshes are refused
// for jwksRefreshInterval.
func (c *recordKeyCache) key(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	_, cached := c.keys[kid]
	expired := c.keys == nil || !now.Before(c.expires)
	if expired || (!cached && now.Sub(c.lastRefresh) >= jwksRefreshInterval) {
		if now.Before(c.backoffUntil) {
			return nil, recordError(VerificationRecordCodeUnknownKey, "record key fetch is backing off after a failure")
		}
		if err := c.refresh(ctx, now); err != nil {
			// The caller giving up is not a JWKS outage; don't penalise the next caller.
			if ctx.Err() == nil {
				c.backoffUntil = now.Add(jwksRefreshInterval)
			}
			return nil, recordError(VerificationRecordCodeUnknownKey, "fetch record keys: "+err.Error())
		}
	}
	jwk, ok := c.keys[kid]
	if !ok {
		return nil, recordError(VerificationRecordCodeUnknownKey, "kid "+kid+" is not published")
	}
	pub, err := jwk.publicKey()
	if err != nil {
		return nil, recordError(VerificationRecordCodeUnknownKey, "kid "+kid+": "+err.Error())
	}
	return pub, nil
}

func (c *recordKeyCache) refresh(ctx context.Context, now time.Time) error {
	docKeys, maxAge, err := c.fetch(ctx, c.baseURL+recordJWKSPath)
	if err != nil {
		return err
	}
	keys := make(map[string]recordJWK, len(docKeys))
	for _, k := range docKeys {
		if _, dup := keys[k.Kid]; !dup {
			keys[k.Kid] = k
		}
	}
	c.keys = keys
	c.expires = now.Add(maxAge)
	c.lastRefresh = now
	return nil
}

func (c *recordKeyCache) fetch(ctx context.Context, url string) ([]recordJWK, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "magic-auth-go/"+Version)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxJWKSBytes))
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxJWKSBytes+1))
	if err != nil {
		return nil, 0, fmt.Errorf("%s: read JWKS: %w", url, err)
	}
	if len(body) > maxJWKSBytes {
		return nil, 0, fmt.Errorf("%s: JWKS is larger than 1 MiB", url)
	}
	keys, err := parseJWKS(body)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", url, err)
	}
	return keys, cacheMaxAge(resp.Header.Get("Cache-Control")), nil
}

// cacheMaxAge reads the max-age directive; absent or invalid means defaultJWKSMaxAge.
func cacheMaxAge(cacheControl string) time.Duration {
	for _, directive := range strings.Split(cacheControl, ",") {
		name, value, found := strings.Cut(strings.TrimSpace(directive), "=")
		if !found || !strings.EqualFold(name, "max-age") {
			continue
		}
		seconds, err := strconv.ParseInt(value, 10, 64)
		if err != nil || seconds < 0 || seconds > maxSafeInteger {
			continue
		}
		if seconds > int64(math.MaxInt64/time.Second) {
			return time.Duration(math.MaxInt64)
		}
		return time.Duration(seconds) * time.Second
	}
	return defaultJWKSMaxAge
}
