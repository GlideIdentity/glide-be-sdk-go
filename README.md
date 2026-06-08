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
go get github.com/glide/magic-auth-be-sdks/sdks/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/glide/magic-auth-be-sdks/sdks/go/magicauth"
)

func main() {
    // Create client
    client, err := magicauth.NewClient(magicauth.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Prepare authentication
    prepareResp, err := client.Prepare(ctx, &magicauth.PrepareRequest{
        PhoneNumber: "+14155551234",
        Nonce:       "random-base64url-nonce",
        UseCase:     magicauth.UseCaseVerifyPhoneNumber,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Strategy: %s\n", prepareResp.AuthenticationStrategy)
    fmt.Printf("Session Key: %s\n", prepareResp.Session.SessionKey)

    // After obtaining credential from the device...
    verifyResp, err := client.VerifyPhoneNumber(ctx, &magicauth.VerifyPhoneNumberRequest{
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
client, err := magicauth.NewClient(magicauth.Config{
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
resp, err := client.Prepare(ctx, &magicauth.PrepareRequest{
    PhoneNumber: "+14155551234",          // Required for VerifyPhoneNumber
    Nonce:       "base64url-nonce",       // Required
    UseCase:     magicauth.UseCaseVerifyPhoneNumber,
    PLMN:        &magicauth.PLMN{MCC: "310", MNC: "260"},  // Optional
    ClientInfo:  &magicauth.ClientInfo{UserAgent: "..."},  // Optional
})
```

### VerifyPhoneNumber

Verifies a phone number using digital credentials.

```go
resp, err := client.VerifyPhoneNumber(ctx, &magicauth.VerifyPhoneNumberRequest{
    Session:    sessionInfo,
    Credential: "sd-jwt-credential",
})

if resp.Verified {
    fmt.Println("Phone number verified!")
}
```

### GetPhoneNumber

Retrieves the device's phone number.

```go
resp, err := client.GetPhoneNumber(ctx, &magicauth.GetPhoneNumberRequest{
    Session:    sessionInfo,
    Credential: "sd-jwt-credential",
})

fmt.Printf("Phone: %s\n", resp.PhoneNumber)
```

### CheckStatus / CheckStatusPublic

Polls for verification status.

```go
// Authenticated endpoint
resp, err := client.CheckStatus(ctx, sessionKey)

// Public endpoint (no auth required)
resp, err := client.CheckStatusPublic(ctx, sessionKey)

switch resp.Status {
case magicauth.StatusPending:
    // Still waiting
case magicauth.StatusCompleted:
    // Done!
case magicauth.StatusFailed:
    // Check resp.Error
}
```

### ReportInvocation

Reports user interaction for analytics.

```go
resp, err := client.ReportInvocation(ctx, &magicauth.ReportInvocationRequest{
    SessionID: sessionID,
})
```

## Error Handling

Use Go's `errors.Is()` and `errors.As()` for error handling:

```go
resp, err := client.Prepare(ctx, req)
if err != nil {
    // Check for specific error types
    if errors.Is(err, magicauth.ErrCarrierNotEligible) {
        // Carrier not supported
    } else if errors.Is(err, magicauth.ErrSessionNotFound) {
        // Session expired
    } else if errors.Is(err, magicauth.ErrRateLimit) {
        // Rate limited - check RetryAfter
    }

    // Get detailed error information
    var apiErr *magicauth.APIError
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
| `ErrUnsupportedPlatform` | Platform not supported |
| `ErrPhoneNumberMismatch` | Phone number doesn't match |
| `ErrInvalidCredential` | Invalid credential format |
| `ErrRateLimit` | Rate limit exceeded |
| `ErrInternalServer` | Server error |

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

client, _ := magicauth.NewClient(magicauth.Config{
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

client, _ := magicauth.NewClient(magicauth.Config{
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

SDK Version: 2.0.1 (aligned with API version)

## License

See LICENSE file in the repository root.
