# Changelog

All notable changes to the Magic Auth Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.2.1] - 2026-11-18

### Added

- **Call authentication**: `StrategyCall` (`"call"`) authentication strategy and `ProtocolCall` protocol; `PrepareOptions.Call` (`*CallOptions` with `CalledNumber`); prepare returns `CallStrategyData` (`URL`, `StatusURL`) in `Data`. `StatusResponse` carries `Stage` and `LiveCall`; `VerifyPhoneNumberResponse.LiveCall` reports the live-call check (`LiveCallInfo`: `Checked`, `Status`, optional `CarrierName`, `CheckedAt`, `Reason`). Signed verification records accept optional `live_call` claims (`VerificationRecordLiveCall`; `Status` is required when the claim is present). Companion sessions of a call verification carry `SessionMetadataParentStrategy` (`"parent_strategy"`) in session metadata.
- Opt-in completion-page beacons via `GetCompletionPageHTML(endpoint, CompletionPageOptions{Beacon: true})`. Default is off and matches the pre-beacon page. When on: `page_start` / `invalid_link` / `post_outcome` / `close` / `tab_survived` to `window.location.pathname`, never `agg_code`; `/complete` JSON is parsed only on non-2xx; `close` sends `close_outcome` (`unknown` | `failed`) and `trigger` (`pagehide` | `visibilitychange` | `timeout`), never `close_succeeded`; `tab_survived` fires at most once after `close` if the tab becomes visible again; the close hint reappears on that visible return.
- Added `biometric` (`available` | `unavailable`) on the prepare response and on eligibility requirements.
- **Passkey step-up**: `PrepareOptions.Passkey` / `StepUp`, `PrepareResponse.Passkey` / `StepUp`, `VerifyPhoneNumberRequest.PasskeyAssertion`, `VerifyPhoneNumberResponse.Passkey` / `StepUp`, `ActionHash`, and `Client.VerifyStepUp`, which fails with `VerificationRecordCodeStepUpRejected` (`STEP_UP_REJECTED`) unless the record is verified, carries the `passkey` and `carrier` factors, and matches the action and phone number.
- **`VerificationRecordClaims.Factors` / `ActionHash`**: `Factors` defaults to `["carrier"]` when absent; a record with `factors: ["passkey"]` must have no carrier sign-off.
- **`ErrCarrierTemporarilyUnavailable` / `CodeCarrierTemporarilyUnavailable`**: 422 `CARRIER_TEMPORARILY_UNAVAILABLE`, retryable after `RetryAfter`. It previously matched `ErrInternalServer` via `errors.Is`.

### Fixed

- Opt-in page beacons fall back to `fetch` keepalive when `sendBeacon` is missing, returns `false`, or throws.
- `feCode` is returned when a passkey prepare falls back to Link; set the binding cookie for `fallback.session`.
- `StrategyPasskey` (`"passkey"`) authentication strategy.
- `RetryAfter` falls back to `details.retry_after_seconds` when the `Retry-After` header is absent.
- A 429 with no known body code is `ErrRateLimit` instead of `ErrInternalServer`.
- Token endpoint: a 5xx or `temporarily_unavailable` maps to `ErrServiceUnavailable` and a 429 to `ErrRateLimit`, instead of `ErrBadRequest` / `ErrUnauthorized`.
- Token 429 sets `RetryAfter` from `Retry-After`; negative `details.retry_after_seconds` is ignored.

### Changed

- Documented that an undated no-swap answer grades LOW at a window of 720 hours or more and UNKNOWN below that; dated answers keep their band grading.

## [3.2.0] - 2026-10-08

### Added

- **`BiometricVerification` on verify and get-phone responses**: `VerifyPhoneNumberResponse` and `GetPhoneNumberResponse` carry `BiometricVerification`, the outcome of the device biometric step, of the new type `BiometricVerification` (`BiometricVerificationVerified`, `BiometricVerificationFailed`, `BiometricVerificationNotEnrolled`, `BiometricVerificationUnsupported`, `BiometricVerificationNotRequested`). On verify it equals the signed record's `biometric_verification` claim. A value outside these five is kept as sent. Treat anything other than `BiometricVerificationVerified` as not verified.
- **TS43 biometric option**: `PrepareOptions.Biometric` (`*bool`; nil is omitted, an explicit `false` is sent) asks the device to run a biometric step during the TS43 credential ceremony. The outcome comes back as `BiometricVerification`. Ignored on Link; on Desktop (QR), set it on the desktop prepare and the phone inherits it. It never blocks authentication: a device that cannot run it still authenticates.
- **`Nonce` on verify and get-phone responses**: `VerifyPhoneNumberResponse` and `GetPhoneNumberResponse` carry `Nonce`, the nonce of the session prepare created. On verify it equals the signed record's `nonce` claim.
- **`RequestID` and `SessionKey` on results**: every response type carries `RequestID`, the same value as the `X-Request-ID` header, to quote to Glide support: `EligibilityResponse`, `PrepareResponse` (and so `PrepareResult`), `VerifyPhoneNumberResponse`, `GetPhoneNumberResponse`, `ReportInvocationResponse`, `CompleteResponse`, `StatusResponse` and `DesktopWaitResponse` (the `*DesktopWaitError` body). `VerifyPhoneNumberResponse`, `GetPhoneNumberResponse`, `ReportInvocationResponse` and `CompleteResponse` also carry `SessionKey`, the session the result is about, so a result can be stored or logged without the request (prepare's is `Session.SessionKey`). `*APIError` carries `SessionKey` once the server has read the session.
- **`ErrCarrierAppUnavailable` / `CodeCarrierAppUnavailable`**: Desktop (QR) verify answers 422 `CARRIER_APP_UNAVAILABLE` when the phone that scanned the code has no working carrier app. The session cannot continue; start a new QR.
- **`ErrCarrierRejected` / `CodeCarrierRejected`**: 422 `CARRIER_REJECTED` when the carrier rejected a device-bound (Link or Desktop) authentication. `APIError.Details` carries `carrier_error`, `carrier_error_description` and, when present, `stage` and `error_code`.
- **Signed verification record**: `VerifyPhoneNumberResponse` carries `VerificationRecord`, a Glide-signed ES256 JWS of the verify result, and `VerificationRecordStatus` (`VerificationRecordStatusSigned` / `VerificationRecordStatusUnavailable`). `Client.VerifyVerificationRecord(ctx, record, *VerificationRecordOptions)` checks it against Glide's record keys at `<baseURL>/.well-known/us/jwks.json`, then issuer, audience, validity window, claim shape and `ExpectedNonce`, and returns `*VerifiedVerificationRecord` (claims plus `SignOffResults`).
  - Failures are `*VerificationRecordError` with a `VerificationRecordCode*` constant. An empty record fails closed with `RECORD_UNAVAILABLE`, and a record over 16 KiB or one that does not match the claim contract fails `MALFORMED`.
  - `ExpectedAudience` is the record audience configured in the developer portal (empty checks your client ID). `ExpectedOrigin` is a list; the record's `origin` must equal one entry exactly, else `ORIGIN_MISMATCH`. Link records carry `UNKNOWN`, so leave `UNKNOWN` out of the list to reject them.
  - `AllowExpired` (bounded by `MaxAuditAge`, 90 days by default) supports audits; `RevokedKIDs` and `Issuer` are optional.
  - Claims: `ver` `"1.0"`, `jti`, `sub`, `nonce`, `iat`, `nbf`, `exp` (1 to 60 seconds after `iat`), `verified`, `use_case`, `origin`, `biometric_verification`, `sign_offs` (exactly one `carrier` entry), and optional `strategy`, `phone_number_details`, `carrier` and `sim_swap` / `device_swap` with an `age_band` `{from, to}` in hours. Standard library only.
- **`PrepareOptions.CompletionRedirectURL`**: names where the browser is sent once authentication completes, selecting one of the completion redirect URLs registered on your application. Matching is exact, and an unregistered value is rejected. One OAuth credential can then serve several deployments on different origins. Set it server-side from your deployment's configured origin, never from the request `Host` header or other client-controlled input. If your application registers more than one completion URL, always send it; without it the first registered URL is used.
- **`PrepareMultiple`**: prepare several phone numbers in one call, with one result per number.
- **Configurable SIM swap window**: `PrepareOptions.SimSwapMaxAgeHours` (`*int`, 1–2400) sets how far back the SIM and device swap check looks, overriding the carrier's default. `SimSwapInfo` and `DeviceSwapInfo` carry `Swapped` (`*bool`; nil when the carrier gave no determinable age, which is not the same as `false`) and `MaxAgeHours` (the window the answer describes). A carrier that finds no swap inside the window reports `RiskLevelUnknown`.
- **Optional `ClientInfo.PrivacyMode`**: sent on prepare as `privacy_mode`. When set to `private`, prepare answers 422 `PRIVATE_BROWSING_NOT_SUPPORTED`.
- **Network check**: `PrepareOptions.NetworkOrigin` (the requesting device's IP as your backend observed it), set on both the desktop prepare and the companion prepare (the request carrying `Options.ParentSessionID`), compares the two devices' egress IPs. The result, `NetworkCheck`, is returned on the companion prepare and on `VerifyPhoneNumberResponse` and `GetPhoneNumberResponse`. `Match` is `same_ip`, `same_subnet`, `different` or `unknown`; branch on it. When the two public egress IPs were compared, `Family` is `ipv4` or `ipv6` and `PrefixBits` is the number of matching leading bits (0–32 for IPv4, 0–128 for IPv6; the maximum means the same egress IP); `PrefixBits` is a `*int`, so a real 0 (nothing in common) is distinct from nil; on `unknown`, `Family` is empty, `PrefixBits` is nil, and both are omitted when the result is serialized. `DesktopIP` and `CompanionIP` are the addresses compared. It is a signal, not a gate; you choose the threshold. `NetworkCheck` is nil when the desktop prepare did not send `NetworkOrigin`. Derive the IP server-side; never trust a client-supplied value. New types: `NetworkOrigin`, `NetworkCheck`.

### Changed

- **`CARRIER_APP_UNAVAILABLE` is its own error**: it matches `ErrCarrierAppUnavailable` with `errors.Is`, where 3.0.0 matched `ErrInternalServer`.

### Fixed

- Fixed an issue where the Link completion page could report a timeout after a successful verification.
- Fixed `APIError.Details` missing some of the error's detail fields.

## [3.0.0] - 2026-06-08

First public release, as `github.com/GlideIdentity/glide-be-sdk-go/v3`. `ReportInvocationResponse` has a breaking shape change: code reading `resp.Status` or using the `ReportStatus` / `IsSuccess` / `Success` symbols no longer compiles; read `resp.Success` (bool).

### Added

- **`CheckEligibility`**: carrier and strategy discovery for a phone number or PLMN, without creating a session. `PrepareRequest` gains optional `AuthenticationStrategy` and `SessionTimeout`.

### Changed

- **Breaking**: `ReportInvocationResponse` matches the API's `{"success": true}` response.
  - Removed: `ReportStatus` type, `ReportStatusSuccess`, `ReportStatusError` constants, `IsSuccess()` method, deprecated `Success()` method.
  - New shape: `ReportInvocationResponse{ Success bool json:"success" }`. Callers access `resp.Success` directly.
- **Breaking**: `DesktopQRData` is a single universal QR. The companion app detects the device at runtime and picks TS43 (Android) or Link (iOS), so there are no per-platform QR images or deep links.
  - Removed fields: `AndroidQRImg`, `IOSQRImg`, `AndroidURL`, `IOSURL` (and their `android_qr_image` / `ios_qr_image` / `android_url` / `ios_url` JSON tags).
  - Added fields: `QRImage` (`json:"qr_image"`, base64 PNG data URL of the universal QR) and `URL` (`json:"url"`, the universal mobile-auth URL encoded in the QR). Both are required.
  - Callers should render a single inline QR using `desktopData.Data.QRImage` instead of toggling between platforms.

## [2.3.0] - 2026-02-26

### Added

- **Device binding (link protocol)**: Anti-phishing protection that cryptographically binds Prepare, Complete, and VerifyPhoneNumber/GetPhoneNumber to the same browser
  - `GenerateFeCode()` -- generates 32 random bytes as 64-char lowercase hex
  - `ComputeFeHash(feCode)` -- SHA-256 hash with lowercase normalization
  - `IsValidBindingCode(s)` -- validates 64-char hex (case-insensitive)
  - `HexPattern64` -- compiled regex (exported for reuse)
- **Session-scoped binding cookies** (`_glide_bind_{sessionKey}`) to prevent parallel session collisions
  - `GetBindingCookieName(sessionKey)` -- returns full cookie name with injection validation
  - `BuildSetBindingCookieHeader(feCode, sessionKey, opts)` -- HttpOnly, Secure, SameSite=Lax, Max-Age=300
  - `BuildClearBindingCookieHeader(sessionKey, opts)` -- includes both Max-Age=0 and Expires
  - `ParseBindingCookie(cookieHeader, sessionKey)` -- extracts and validates fe_code
  - `ClearStaleBindingCookies(cookieHeader, opts)` -- expires all `_glide_bind_*` cookies
- **Completion page HTML**: `GetCompletionPageHTML(completeEndpoint)` -- self-contained HTML page with CSP, localStorage signal, auto-close, friendly error UI
- **`Complete(ctx, req, opts...)`** -- new client method for `POST /magic-auth/v2/auth/complete`
- **`PrepareResult`** struct embedding `PrepareResponse` with optional `FeCode` (set for link strategy only)
- **`CompleteRequest`** / **`CompleteResponse`** types
- **`DesktopWaitResponse`** type for HTTP 202 retry contract
- **`DesktopWaitError`** -- returned on HTTP 202, carries `Response` with retry info
- **`DeviceSwapInfo`** type -- IMEI change detection (parallel to `SimSwapInfo`)
- **`DeviceSwap`** field on `VerifyPhoneNumberResponse` and `GetPhoneNumberResponse`
- **`FeCode`** field on `VerifyPhoneNumberRequest`, `GetPhoneNumberRequest`, and `LinkStrategyData`
- **`FeHash`** field on `PrepareRequest`
- **`StatusPendingCompletion`** added to `SessionStatus`
- **`BindingCookieOptions`** type for cookie configuration
- **New sentinel errors**: `ErrForbidden`, `ErrBrowserMismatch`, `ErrSessionNotEligible`, `ErrSessionExpired`, `ErrDeviceBindingFailed`, `ErrMissingBindingCookie`
- **New error codes**: `CodeForbidden`, `CodeBrowserMismatch`, `CodeSessionNotEligible`, `CodeQRAlreadyScanned`, `CodeSessionExpired`, `CodeAudienceValidationFailed`, `CodeDeviceBindingFailed`, `CodeMissingBindingCookie`
- **`RequestOptions`** struct with `CorrelationID` for `X-Correlation-ID` header on all API methods
- **`SensitiveBindingKeys`** list for log sanitization of binding secrets

### Changed

- `Prepare()` now returns `*PrepareResult` (superset of `PrepareResponse`) and auto-generates `FeCode`/`FeHash`
- `VerifyPhoneNumber()` and `GetPhoneNumber()` now return `DesktopWaitError` on HTTP 202
- `ConsentData.PolicyText` now has `omitempty` tag
- All API methods accept variadic `*RequestOptions` parameter

## [2.2.0] - 2026-02-01

### Breaking Changes

- **`StatusResponse` schema updated**: The following fields have been removed:
  - `UseCase` - No longer returned from status endpoints
  - `Verified` - No longer returned from status endpoints
  - `PhoneNumber` - No longer returned from status endpoints (use authenticated verify/get-phone-number endpoints)
  - `Error` - Replaced by `Extra` map for error metadata
- **`StatusResponse.Protocol`** is now optional (has `omitempty` tag)

### Added

- **`StatusResponse.Extra`** field (`map[string]interface{}`) for additional metadata (e.g., error reason for failed status)

### Changed

- Session key pattern updated in docs/comments from 64-char to 16-char hex format

### Migration Guide

If you were accessing removed fields from `StatusResponse`:

```go
// Before (these fields no longer exist)
if resp.Verified { ... }
useCase := resp.UseCase
phoneNumber := resp.PhoneNumber

// After - use dedicated endpoints for sensitive data
verifyResp, err := client.VerifyPhoneNumber(ctx, &magicauth.VerifyPhoneNumberRequest{...})
if verifyResp.Verified { ... }

// For error info, check the Extra map
if resp.Status == magicauth.StatusFailed {
    if reason, ok := resp.Extra["reason"].(string); ok {
        log.Printf("Failed: %s", reason)
    }
}
```

## [2.1.0] - 2026-01-29

### Breaking Changes

- **`ClientInfo.UserAgent` is now required**: The `omitempty` tag has been removed. You must provide the end-user's browser user agent string.
- **`PrepareRequest.ClientInfo` is now required**: Changed from `*ClientInfo` (pointer, optional) to `ClientInfo` (value, required). All `Prepare` calls must include `ClientInfo` with a valid `UserAgent`.

### Changed

- Added comprehensive documentation to `ClientInfo` explaining that `UserAgent` must be the end-user's browser `navigator.userAgent`, not the server's user agent.

### Migration Guide

If you were previously calling `Prepare` without `ClientInfo`:

```go
// Before (will no longer compile)
req := magicauth.PrepareRequest{
    Nonce:   "...",
    UseCase: magicauth.UseCaseVerifyPhoneNumber,
}

// After (required)
req := magicauth.PrepareRequest{
    Nonce:   "...",
    UseCase: magicauth.UseCaseVerifyPhoneNumber,
    ClientInfo: magicauth.ClientInfo{
        UserAgent: userAgentFromFrontend, // Must come from end-user's browser
    },
}
```

## [2.0.2] - 2026-01-29

### Added

- **Strategy-specific typed data**: Added concrete types for `PrepareResponse.Data` parsing
  - `TS43StrategyData` - T-Mobile TS43 protocol data with DCQL query
  - `LinkStrategyData` - Verizon OAuth/Link redirect data
  - `DesktopStrategyData` - Desktop QR code cross-device authentication data
- New supporting types: `DCQLCredential`, `DCQLCredentialMeta`, `DCQLQuery`, `TS43RequestData`, `DesktopQRData`
- `ReportStatus` type for report invocation status enum (`ReportStatusSuccess`, `ReportStatusError`)

### Changed

- **Breaking**: `DeviceInfo.Type` and `DeviceInfo.OS` are now required fields (removed `omitempty`)
- **Breaking**: `StatusResponse.Protocol`, `StatusResponse.UseCase`, and `StatusResponse.Verified` are now required fields (removed `omitempty`)
- **Breaking**: `ReportInvocationResponse.Status` is now typed as `ReportStatus` (was `string`)

### Fixed

- Aligned all model types with OpenAPI spec v2.0.0 for full spec compliance

## [2.0.1] - 2026-01-29

### Added

- Initial release of the Magic Auth Go SDK
- All 7 API operations implemented:
  - `Prepare` - Prepare authentication and get strategy-specific data
  - `VerifyPhoneNumber` - Verify phone number using digital credentials
  - `GetPhoneNumber` - Retrieve device phone number
  - `ReportInvocation` - Report user interaction for analytics
  - `CheckStatus` - Poll verification status (authenticated)
  - `CheckStatusPublic` - Poll verification status (public)
  - Internal token management (automatic)
- Zero external dependencies (stdlib only)
- BYOC (Bring Your Own Client) pattern with `HTTPClient` interface
- BYOL (Bring Your Own Logger) pattern with `Logger` interface
- Thread-safe token caching with `sync.RWMutex`
- Token auto-refresh with 60-second buffer before expiry
- Full `context.Context` support for cancellation and timeouts
- Idiomatic Go error handling:
  - Sentinel errors for `errors.Is()` checks
  - `APIError` type with `Unwrap()` for `errors.As()` extraction
  - Detailed error information (code, message, request_id, trace_id)
- `Retry-After` header extraction for rate limit errors
- Comprehensive test suite with mock HTTP client
- Complete documentation with examples

### API Compatibility

- Aligned with Magic Auth Aggregator API v2.0.1
- Supports all authentication strategies: TS43, Link, Desktop
- Full support for SIM swap information in responses
