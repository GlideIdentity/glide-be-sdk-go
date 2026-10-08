# Magic Auth Go SDK

Go SDK for the Magic Auth Aggregator API - phone number authentication and verification using digital credentials.

## Features

- **Zero external dependencies** - Uses only Go standard library
- **BYOC (Bring Your Own Client)** - Implement `HTTPClient` interface for custom HTTP handling
- **BYOL (Bring Your Own Logger)** - Implement `Logger` interface for custom logging
- **Thread-safe token management** - Automatic token refresh with caching
- **Context support** - Full `context.Context` integration for cancellation/timeouts
- **Idiomatic Go error handling** - Sentinel errors with `errors.Is()` and `errors.As()`

## Installation

```bash
go get github.com/GlideIdentity/glide-be-sdk-go/v3
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/GlideIdentity/glide-be-sdk-go/v3/magicalauth"
)

func main() {
    // Create client
    client, err := magicalauth.NewClient(magicalauth.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Prepare authentication
    prepareResp, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
        PhoneNumber: "+14155551234",
        Nonce:       "random-base64url-nonce",
        UseCase:     magicalauth.UseCaseVerifyPhoneNumber,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Strategy: %s\n", prepareResp.AuthenticationStrategy)
    fmt.Printf("Session Key: %s\n", prepareResp.Session.SessionKey)

    // After obtaining credential from the device...
    verifyResp, err := client.VerifyPhoneNumber(ctx, &magicalauth.VerifyPhoneNumberRequest{
        Session:    prepareResp.Session,
        Credential: "sd-jwt-credential-from-device",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Verified: %v\n", verifyResp.Verified)
}
```

## Configuration

```go
client, err := magicalauth.NewClient(magicalauth.Config{
    // Required
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    
    // Optional (with defaults)
    BaseURL:    "https://api.glideidentity.app",  // Production API
    Timeout:    30 * time.Second,                  // HTTP timeout
    HTTPClient: nil,                               // Custom HTTP client
    Logger:     nil,                               // Custom logger
})
```

## API Methods

### Prepare

Prepares authentication and returns strategy-specific data.

```go
resp, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
    PhoneNumber: "+14155551234",          // Required for VerifyPhoneNumber
    Nonce:       "base64url-nonce",       // Required
    UseCase:     magicalauth.UseCaseVerifyPhoneNumber,
    PLMN:        &magicalauth.PLMN{MCC: "310", MNC: "260"},  // Optional
    ClientInfo:  &magicalauth.ClientInfo{UserAgent: "..."},  // Optional
})
```

#### PrepareOptions

Everything optional about a session is set through `Options`:

| Field | Type | Description |
|---|---|---|
| `ParentSessionID` | `string` | Parent session for the desktop QR flow — set it on the companion/child prepare. |
| `Theme` | `Theme` | UI theme forwarded verbatim to every touchpoint. `ThemeDark` / `ThemeLight`, or any custom identifier (max 32 chars). |
| `LinkType` | `LinkType` | Overrides link dispatch (`app_clip`, `app_link`, `headless`, `web_link`). Link strategy only. |
| `NetworkOrigin` | `*NetworkOrigin` | The **server-observed** IP of the requesting device, stamped by your backend. Stamp it on both prepares to get `NetworkCheck` back. Never trust a client-supplied value. |
| `MobileCompanionPassthrough` | `json.RawMessage` | Opaque JSON forwarded to the branded companion app. Object only, ≤16 KB. |
| `CompletionRedirectURL` | `string` | Where the browser lands once authentication completes. Selects one of your application's registered `completion_redirect_urls` (exact match per RFC 6749 §3.1.2). Stamp it server-side from your own deployment's origin — never from the request `Host` header or any client-controllable value. |
| `SimSwapMaxAgeHours` | `*int` | SIM swap query window in hours. See below. |
| `Biometric` | `*bool` | Opt in to the TS43 biometric signal. See below. |

**`SimSwapMaxAgeHours`** sets how far back a SIM/device swap check looks. Integer, **1–2400**; anything outside that range is refused with `400 VALIDATION_ERROR` before a session is created. When omitted, the server falls back to the carrier's configured `default_max_age_hours` and then to a **24 hour** default. The window that actually produced the answer comes back as `MaxAgeHours` on the swap payloads, so you can always tell which window you are reading.

**`Biometric`** asks the device to run a biometric step during the TS43 credential ceremony and report the outcome as `BiometricVerification`. It is a `*bool` so the option stays tri-state: `nil` is omitted from the request, while an explicit `false` still serialises as JSON `false`. On Desktop (QR), set on the parent prepare; the phone inherits it. Ignored on Link. Requesting it never makes biometric mandatory: a device that cannot perform it still authenticates.

```go
window := 72
biometric := true

resp, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
    PhoneNumber: "+14155551234",
    Nonce:       "base64url-nonce",
    UseCase:     magicalauth.UseCaseVerifyPhoneNumber,
    ClientInfo:  &magicalauth.ClientInfo{UserAgent: reqUserAgent},
    Options: &magicalauth.PrepareOptions{
        SimSwapMaxAgeHours: &window,
        Biometric:          &biometric,
    },
})
```

### PrepareMultiple

Prepares authentication for up to `MaxPrepareMultiple` (10) phone numbers in one call. Fans out one ordinary `Prepare` per number — each gets its own nonce, session, and (for link strategy) `FeCode`/`FeHash`. Returns one `PrepareMultipleResult` per input number, in input order.

Per-number failures are returned in the slice, not thrown. `PrepareMultiple` returns an error only for caller-input problems (`ErrNoPhoneNumbers`, `ErrTooManyPhoneNumbers`, `ErrBlankPhoneNumber`, nil request, or a shared-field validation error such as `ErrMissingUseCase`) and only before any HTTP request is issued.

`CARRIER_NOT_ELIGIBLE` and `CARRIER_TEMPORARILY_UNAVAILABLE` are both HTTP 422; only the code tells a durable refusal from a transient, retryable carrier outage — on each failed result, branch with `errors.Is(r.Err, magicalauth.ErrCarrierNotEligible)` or extract a `*magicalauth.APIError` with `errors.As(r.Err, &apiErr)` and inspect `apiErr.Code`, not status alone, or a transiently breakered number gets hidden as permanently ineligible.

Omit `FeCode` from each successful `PrepareResult` before forwarding to a browser — it is the device-binding secret and must never reach client-side JavaScript.

Pass a short `SessionTimeout` (60–900 seconds, same as `Prepare`) so unpicked sessions expire quickly.

```go
sessionTimeout := 120
results, err := client.PrepareMultiple(ctx, &magicalauth.PrepareMultipleRequest{
    PhoneNumbers:   []string{"+14155551234", "+14155559876"},
    UseCase:        magicalauth.UseCaseVerifyPhoneNumber,
    ClientInfo:     magicalauth.ClientInfo{UserAgent: reqUserAgent},
    SessionTimeout: &sessionTimeout,
})
if err != nil {
    // validation only — empty list, over cap, blank entry, or invalid shared prepare fields
    return err
}

for _, r := range results {
    if !r.OK() {
        continue
    }
    // For the link strategy, retain FeCode server-side before stripping it —
    // Complete needs it to authenticate the picked session. The usual place is
    // a session-scoped HttpOnly cookie:
    if r.Result.FeCode != "" {
        header, err := magicalauth.BuildSetBindingCookieHeader(r.Result.FeCode, r.Result.Session.SessionKey, nil)
        if err != nil {
            return err
        }
        w.Header().Add("Set-Cookie", header)
    }
    safe := *r.Result
    safe.FeCode = ""
    // build picker entry from r.PhoneNumber and safe
}
```

Go uses nullable `Result`/`Err` fields and an `OK()` helper; Node and Java express the same contract with a discriminated union and `isOk()` respectively.

### VerifyPhoneNumber

Verifies a phone number using digital credentials.

```go
resp, err := client.VerifyPhoneNumber(ctx, &magicalauth.VerifyPhoneNumberRequest{
    Session:    sessionInfo,
    Credential: "sd-jwt-credential",
})

if resp.Verified {
    fmt.Println("Phone number verified!")
}
```

### Verifying the signed verification record

`VerifyPhoneNumberResponse` also carries `VerificationRecord`, a Glide-signed ES256 JWS of the verify result, and `VerificationRecordStatus` (`VerificationRecordStatusSigned` or `VerificationRecordStatusUnavailable`). Keep the record if you need proof of the result later, and check it with `VerifyVerificationRecord`, passing the nonce you sent to prepare:

```go
rec, err := client.VerifyVerificationRecord(ctx, resp.VerificationRecord, &magicalauth.VerificationRecordOptions{
    ExpectedNonce: prepareNonce,
})
var recErr *magicalauth.VerificationRecordError
if errors.As(err, &recErr) {
    // recErr.Code is a VerificationRecordCode* constant, e.g. "BAD_SIGNATURE"
}
if err == nil && rec.Claims.Verified {
    fmt.Println("Signed result for", rec.Claims.Sub)
}
```

- An empty record fails closed with `VerificationRecordCodeRecordUnavailable` (`RECORD_UNAVAILABLE`), so the response field can be passed straight in.
- Records longer than 16 KiB are rejected as `MALFORMED` before decoding.
- The other codes are `MALFORMED`, `DISALLOWED_HEADER`, `UNKNOWN_KEY`, `BAD_SIGNATURE`, `WRONG_ISSUER`, `WRONG_AUDIENCE` (`aud` must equal the expected audience, default your client ID), `EXPIRED`, `NOT_YET_VALID`, `NONCE_MISMATCH`, `ORIGIN_MISMATCH`, `UNSUPPORTED_VERSION` and `REVOKED_KEY`. Every failure is a `*VerificationRecordError` and matches `ErrVerificationRecord` with `errors.Is`.
- Claims that break the v1 contract fail with `MALFORMED`: `ver` must be the string `"1.0"` (otherwise `UNSUPPORTED_VERSION`), `verified` a boolean, `jti`, `nonce` and `sub` (`tel:+E164`) non-empty strings, `nbf` equal to `iat`, `use_case` exactly `VerifyPhoneNumber`, `origin` (canonical device origin, `UNKNOWN` on Link) and `biometric_verification` strings, `exp` 1 to 60 s after `iat`, and `sign_offs` an array of entries with non-empty `role`, `signer` and `status` (`unavailable` with a `reason`, or `present` with `format` and `token`), exactly one of them a `carrier` entry signed `carrier:<id>`. Optional `strategy` (`TS_43`, `LINK`, `DESKTOP__TS_43`, `DESKTOP__LINK`, `DESKTOP`), `phone_number_details`, `carrier` (`{id, name, mvno}`), `sim_swap` and `device_swap` are accepted when present; a wrong JSON type, a `carrier` without a non-empty string `id`, or a `carrier.id` that does not match the carrier sign-off signer (`carrier:<id>`), is `MALFORMED`. Record `sim_swap` / `device_swap` carry `age_band` `{from, to}` in hours (`-1` unknown) and omit `checked_at` — `iat` is the authoritative check time.
- Keys come only from `<baseURL>/.well-known/us/jwks.json` (it carries both regions' record keys), cached per `Cache-Control` (300 s default); nothing in the token picks a URL. Only Glide's record-signing kids are accepted. Glide keeps a retired record key published for at least 90 days (the default `MaxAuditAge`); a record whose key has since been unpublished fails `UNKNOWN_KEY`, even with `AllowExpired`.
- Set `ExpectedAudience` to the record audience you configured in the developer portal (empty checks your client ID). To pin the device origin, set `ExpectedOrigin` (a list): the record's canonical `origin` must equal one entry exactly, else `ORIGIN_MISMATCH`; Link records carry `UNKNOWN`, so leave `UNKNOWN` out of the list to reject them.
- To audit a stored record after `exp`, set `AllowExpired`; `iat` must still be within `MaxAuditAge` (90 days by default) and revoked kids (`RevokedKIDs` plus the built-in list) are always rejected. `Issuer` overrides the expected `iss` (default: the base URL).
- In v1 `Claims.SignOffs` holds one `carrier` entry with status `unavailable`; `SignOffResults` reports it as `unavailable`, and a present sign-off would be reported `unverified`, never trusted.
- `ctx` bounds the JWKS fetch. A fetch that ends because `ctx` was cancelled fails with `UNKNOWN_KEY` without starting the 60 s refresh backoff that a real fetch failure does.

### GetPhoneNumber

Retrieves the device's phone number.

```go
resp, err := client.GetPhoneNumber(ctx, &magicalauth.GetPhoneNumberRequest{
    Session:    sessionInfo,
    Credential: "sd-jwt-credential",
})

fmt.Printf("Phone: %s\n", resp.PhoneNumber)
```

### Risk signals on the success response

`VerifyPhoneNumberResponse` and `GetPhoneNumberResponse` carry the same signals: `BiometricVerification`, plus the optional `SimSwap`, `DeviceSwap` and `PhoneNumberDetails`. Each optional one is omitted when the carrier or the session did not produce it, so branch on presence rather than assuming a zero value is meaningful.

#### SimSwapInfo

| Field | Type | Description |
|---|---|---|
| `Checked` | `bool` | Whether the check completed. When `false`, read `Reason` and ignore the rest. |
| `RiskLevel` | `RiskLevel` | `RiskLevelHigh`, `RiskLevelMedium`, `RiskLevelLow`, `RiskLevelUnknown`. |
| `AgeBand` | `string` | When the SIM last changed, e.g. `"0-4 hours"`, `"7-14 days"`, `"more than 3 years"`, `"older than query window"`, `"temporarily unavailable"`, `"not available"`. |
| `CarrierName` | `string` | The carrier that performed the check. |
| `CheckedAt` | `string` | When the check ran, RFC3339. Kept as a string — the SDK passes the wire value through without parsing. |
| `Swapped` | `*bool` | Whether a swap fell inside `MaxAgeHours`. **`nil` when the carrier gave no determinable answer** (T-Mobile codes `999` / `NA`) — that is not the same as `false`. |
| `MaxAgeHours` | `int` | The window this answer actually describes. |
| `Reason` | `SimSwapFailureReason` | Only when `Checked` is `false`: `timeout`, `carrier_not_supported`, `disabled`, `error`. |

#### DeviceSwapInfo

Device (IMEI) change, reported only by carriers that support it — currently T-Mobile. It is a fully independent signal, not a variant of `SimSwapInfo`: identical fields **except that it has no `Reason`**, because the server never populates one.

#### Reading the swap window

A `nil` `Swapped` means the check was inconclusive — the carrier gave no determinable age. `RiskLevelUnknown` is broader: it covers that inconclusive case **and** a CAMARA carrier successfully confirming no swap inside the window, which returns `Swapped` pointing at `false` and `AgeBand` `"older than query window"` but still grades `RiskLevelUnknown`, because the carrier never dated the last swap. That second case used to report `RiskLevelLow`, so code reading `Low` as "safe" now falls through to its default. Branch on `Swapped` for "was there a swap in my window"; read `RiskLevelUnknown` as "no risk grade available", not as a good or bad outcome.

A known `age_band` together with `Swapped` pointing at `false` is not a contradiction: it means the carrier dated the last swap and it fell outside the window. The flag is the band overlapped with the window, and a band straddling the boundary resolves to swapped rather than to an unknown — so with a 24-hour window a `"14-30 days"` band gives `*Swapped == false` alongside `RiskLevelMedium`, because risk comes from the band's own recency, not from the window.

```go
if s := resp.SimSwap; s != nil && s.Checked {
    switch {
    case s.Swapped == nil:
        // carrier could not determine an age for this window
    case *s.Swapped:
        log.Printf("SIM swapped within %dh (%s)", s.MaxAgeHours, s.AgeBand)
    }
}
```

#### BiometricVerification

`BiometricVerification` is on every success response.

| Value | Meaning |
|---|---|
| `VERIFIED` | The device performed biometric verification. |
| `FAILED` | The device attempted it and reported a failure. |
| `NOT_ENROLLED` | No biometric is enrolled on the device. |
| `UNSUPPORTED` | Biometric was requested but no usable result came back. |
| `NOT_REQUESTED` | Nothing to report. |

On verify it equals the signed record's `biometric_verification` claim. Values outside this list are kept as sent, so treat anything unrecognised as "no useful signal". It is a signal, not an authentication result — the phone number is verified either way.

### CheckStatus / CheckStatusPublic

Polls for verification status.

```go
// Authenticated endpoint
resp, err := client.CheckStatus(ctx, sessionKey)

// Public endpoint (no auth required)
resp, err := client.CheckStatusPublic(ctx, sessionKey)

switch resp.Status {
case magicalauth.StatusPending:
    // Still waiting
case magicalauth.StatusCompleted:
    // Done!
case magicalauth.StatusFailed:
    // Check resp.Error
}
```

### ReportInvocation

Reports user interaction for analytics.

```go
resp, err := client.ReportInvocation(ctx, &magicalauth.ReportInvocationRequest{
    SessionID: sessionID,
})
```

## Error Handling

Use Go's `errors.Is()` and `errors.As()` for error handling:

```go
resp, err := client.Prepare(ctx, req)
if err != nil {
    // Check for specific error types
    if errors.Is(err, magicalauth.ErrCarrierNotEligible) {
        // Carrier not supported
    } else if errors.Is(err, magicalauth.ErrSessionNotFound) {
        // Session expired
    } else if errors.Is(err, magicalauth.ErrRateLimit) {
        // Rate limited - check RetryAfter
    }

    // Get detailed error information
    var apiErr *magicalauth.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Code: %s\n", apiErr.Code)
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("RequestID: %s\n", apiErr.RequestID)
        fmt.Printf("RetryAfter: %d\n", apiErr.RetryAfter)
    }
}
```

### Sentinel Errors

| Error | Description |
|-------|-------------|
| `ErrBadRequest` | Invalid request parameters |
| `ErrUnauthorized` | Authentication failed |
| `ErrValidation` | Validation error |
| `ErrSessionNotFound` | Session not found or expired |
| `ErrCarrierNotEligible` | Carrier not supported |
| `ErrCarrierRejected` | Carrier rejected a device-bound (link/desktop) auth; see `APIError.Details` for `carrier_error` / `carrier_error_description` / `stage` / `error_code` |
| `ErrCarrierAppUnavailable` | Desktop (QR): the phone that scanned the code has no working carrier app (missing, off or outdated); terminal for the session, start a new QR |
| `ErrUnsupportedPlatform` | Platform not supported |
| `ErrPhoneNumberMismatch` | Phone number doesn't match |
| `ErrInvalidCredential` | Invalid credential format |
| `ErrRateLimit` | Rate limit exceeded |
| `ErrInternalServer` | Server error |

### The sentinel list is not exhaustive

Sentinels cover the codes worth branching on today; the API can return codes that have no
sentinel. `specs/openapi.yaml` currently documents `TOKEN_ACQUISITION_FAILED`,
`INVALID_PARAMETERS`, `UNPROCESSABLE_ENTITY`, `UNEXPECTED_ERROR` and
`COMPLETION_REDIRECT_URL_NOT_REGISTERED`, none of which map to a sentinel.

Always keep a fallback branch, and read `apiErr.Code` rather than assuming `errors.Is` against
the list above is total:

```go
var apiErr *magicalauth.APIError
if errors.As(err, &apiErr) {
    switch {
    case errors.Is(err, magicalauth.ErrCarrierNotEligible):
        // handled case
    default:
        // anything else, including codes newer than this SDK
        log.Printf("unhandled %s: %s", apiErr.Code, apiErr.Message)
    }
}
```

The edge can answer `401` and `429` with a `text/plain` body rather than the JSON error envelope.
This SDK handles that: when the body will not parse, it derives the error from the HTTP status
instead, so a plain-text `401` still gives `ErrUnauthorized` / `CodeUnauthorized` and a plain-text
`429` still gives `ErrRateLimit`. The body becomes the message when it is under 200 bytes.

## Custom HTTP Client (BYOC)

Implement the `HTTPClient` interface:

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Example: Add custom headers
type customClient struct {
    base http.Client
}

func (c *customClient) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set("X-Custom-Header", "value")
    return c.base.Do(req)
}

client, _ := magicalauth.NewClient(magicalauth.Config{
    ClientID:     "...",
    ClientSecret: "...",
    HTTPClient:   &customClient{},
})
```

## Custom Logger (BYOL)

Implement the `Logger` interface:

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
}

// Example: Wrap zap logger
type zapLogger struct {
    *zap.SugaredLogger
}

func (l *zapLogger) Debug(msg string, kv ...any) { l.Debugw(msg, kv...) }
func (l *zapLogger) Info(msg string, kv ...any)  { l.Infow(msg, kv...) }
func (l *zapLogger) Warn(msg string, kv ...any)  { l.Warnw(msg, kv...) }
func (l *zapLogger) Error(msg string, kv ...any) { l.Errorw(msg, kv...) }

client, _ := magicalauth.NewClient(magicalauth.Config{
    ClientID:     "...",
    ClientSecret: "...",
    Logger:       &zapLogger{sugar},
})
```

## Testing

Run tests:

```bash
cd sdks/go
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## Version

The SDK version is `magicalauth.Version`; see `CHANGELOG.md` for what each release changes.

## License

Apache License 2.0 — see the LICENSE file in the repository root.
