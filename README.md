# Glide Go SDK

Official Go SDK for Glide Identity's phone authentication services.

**Documentation**: https://docs.glideidentity.com

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────────┐
│    Your     │      │    Your     │      │  Glide Services │
│  Frontend   │ ──── │   Backend   │ ──── │                 │
│             │      │             │      │                 │
│ glide-fe-sdk│      │ glide-be-sdk│      │  magic-auth API │
└─────────────┘      └─────────────┘      └─────────────────┘
```

This SDK runs in your backend application, handling secure communication with Glide's authentication services. It includes HTTP client, logging, retry logic, and error handling.

## Installation

```bash
go get github.com/GlideIdentity/glide-be-sdk-go
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"

    glide "github.com/GlideIdentity/glide-be-sdk-go"
)

func main() {
    client := glide.New(
        glide.WithClientCredentials(
            os.Getenv("GLIDE_CLIENT_ID"),
            os.Getenv("GLIDE_CLIENT_SECRET"),
        ),
    )

    resp, err := client.MagicalAuth.Prepare(context.Background(), &glide.PrepareRequest{
        UseCase:    glide.UseCaseGetPhoneNumber,
        PLMN:       &glide.PLMN{MCC: "310", MNC: "260"},
        ClientInfo: &glide.ClientInfo{Platform: "web"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Strategy: %s\n", resp.AuthenticationStrategy)
    fmt.Printf("Session: %s\n", resp.Session.SessionKey)
}
```

## Production Usage

In production, your backend receives requests from a frontend application using a Glide frontend SDK:

```go
package main

import (
    "encoding/json"
    "net/http"
    "os"
    "time"

    glide "github.com/GlideIdentity/glide-be-sdk-go"
)

var client = glide.New(
    glide.WithClientCredentials(
        os.Getenv("GLIDE_CLIENT_ID"),
        os.Getenv("GLIDE_CLIENT_SECRET"),
    ),
    glide.WithTimeout(30*time.Second),
)

func main() {
    http.HandleFunc("/api/phone-auth/prepare", handlePrepare)
    http.HandleFunc("/api/phone-auth/process", handleProcess)
    http.ListenAndServe(":8080", nil)
}

func handlePrepare(w http.ResponseWriter, r *http.Request) {
    var req glide.PrepareRequest
    json.NewDecoder(r.Body).Decode(&req)

    resp, err := client.MagicalAuth.Prepare(r.Context(), &req)
    if err != nil {
        handleError(w, err)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
    var req glide.ProcessRequest
    json.NewDecoder(r.Body).Decode(&req)

    switch req.UseCase {
    case glide.UseCaseGetPhoneNumber:
        resp, err := client.MagicalAuth.GetPhoneNumber(r.Context(), &glide.GetPhoneNumberRequest{
            Session:    req.Session,
            Credential: req.Credential,
        })
        if err != nil {
            handleError(w, err)
            return
        }
        json.NewEncoder(w).Encode(resp)

    case glide.UseCaseVerifyPhoneNumber:
        resp, err := client.MagicalAuth.VerifyPhoneNumber(r.Context(), &glide.VerifyPhoneNumberRequest{
            Session:    req.Session,
            Credential: req.Credential,
        })
        if err != nil {
            handleError(w, err)
            return
        }
        json.NewEncoder(w).Encode(resp)
    }
}

func handleError(w http.ResponseWriter, err error) {
    if glideErr, ok := err.(*glide.MagicalAuthError); ok {
        w.WriteHeader(glideErr.Status)
        json.NewEncoder(w).Encode(glideErr)
        return
    }
    http.Error(w, err.Error(), http.StatusInternalServerError)
}
```

## Report Invocation

Reports that the user has started the authentication flow:

```go
resp, err := client.MagicalAuth.ReportInvocation(ctx, &glide.ReportInvocationRequest{
    SessionID: session.SessionKey,
})
```

This call is optional and can be made asynchronously (`go client.MagicalAuth.ReportInvocation(...)`) without blocking the authentication flow.

## Configuration

```go
client := glide.New(
    glide.WithClientCredentials(clientID, clientSecret),
    glide.WithBaseURL("https://api.glideidentity.app"),
    glide.WithTimeout(30*time.Second),
    glide.WithRetry(3, time.Second),
    glide.WithLogLevel(glide.LogLevelDebug),
)
```

| Option | Env Variable | Description |
|--------|--------------|-------------|
| `WithClientCredentials(id, secret)` | `GLIDE_CLIENT_ID`, `GLIDE_CLIENT_SECRET` | OAuth2 credentials (required) |
| `WithBaseURL(url)` | `GLIDE_API_BASE_URL` | API base URL |
| `WithTimeout(d)` | — | Request timeout |
| `WithRetry(n, d)` | — | Retry count and delay |
| `WithHTTPClient(c)` | — | Custom HTTP client |
| `WithLogLevel(l)` | `GLIDE_LOG_LEVEL` | Log level (debug, info, warn, error) |
| `WithTokenRefreshBuffer(s)` | — | Seconds before token expiry to refresh (default: 60) |

## Error Handling

```go
resp, err := client.MagicalAuth.Prepare(ctx, req)
if err != nil {
    if glideErr, ok := err.(*glide.MagicalAuthError); ok {
        switch glideErr.Code {
        case glide.ErrCodeConfigurationError:
            // Missing credentials or invalid configuration
        case glide.ErrCodeCarrierNotEligible:
            // Carrier not supported
        case glide.ErrCodeInvalidSession:
            // Session expired
        case glide.ErrCodeRateLimitExceeded:
            // Too many requests
        }
    }
}
```

## Core Package

For minimal dependencies and full control over HTTP requests, use the core package:

```bash
go get github.com/GlideIdentity/glide-be-sdk-go/core
```

See [core/README.md](./core/README.md) for details.

## License

Use of this SDK is permitted solely to enable your use of Glide Identity Inc.'s services. All other use of this SDK is prohibited. All other rights are reserved.
