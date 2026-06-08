# Changelog

All notable changes to the Magic Auth Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - Unreleased

Major version bump — `ReportInvocationResponse` has a breaking shape change. Any consumer reading `resp.Status` or using the `ReportStatus` / `IsSuccess` / `Success` symbols will fail to compile. Consumers should read `resp.Success` (bool) directly.

Note on Go module path: a strict Go v2 release requires the module path to include `/v2` (e.g. `github.com/GlideIdentity/glide-be-sdk-go/v3/v2`). This PR updates the CHANGELOG + SDK surface; the module-path rename is a separate follow-up and will be coordinated with the first public release so existing internal consumers can stay on the current path until they're ready.

### Changed

- **Breaking**: `ReportInvocationResponse` aligned with the actual backend contract. The backend returns `{"success": true}` (and always has). The previous `{"status": "success"|"error", "message"?}` shape was speculative and never returned by the API.
  - Removed: `ReportStatus` type, `ReportStatusSuccess`, `ReportStatusError` constants, `IsSuccess()` method, deprecated `Success()` method.
  - New shape: `ReportInvocationResponse{ Success bool json:"success" }`. Callers access `resp.Success` directly.
  - OpenAPI spec updated to match (`success` required boolean).
- **Breaking**: `DesktopQRData` collapsed to a single universal QR (DEV-1525). The Mobile Auth companion app now detects the device at runtime and picks TS43 (Android) or Link (iOS), so the server no longer returns per-platform QR images or deep links.
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
