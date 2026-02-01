# Glide Go SDK - Core

Minimal types and interfaces for Glide Identity's phone authentication services.

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

This package provides types and interfaces for building custom integrations with Glide's authentication services.

## Installation

```bash
go get github.com/GlideIdentity/glide-be-sdk-go/core
```

## When to Use

Use the core package when you need:
- Minimal dependencies (zero external dependencies)
- Full control over HTTP client configuration

For most applications, use the full SDK instead:
```bash
go get github.com/GlideIdentity/glide-be-sdk-go
```

## Package Contents

| File | Description |
|------|-------------|
| `types.go` | Request and response structs |
| `services.go` | `MagicalAuthService` interface |
| `constants.go` | `UseCase` and `AuthenticationStrategy` constants |
| `version.go` | SDK version |

### Types

| Type | Description |
|------|-------------|
| `PrepareRequest` | Request for initiating authentication |
| `PrepareResponse` | Response from prepare endpoint |
| `Challenge` | Visual verification data for desktop QR flows |
| `SessionInfo` | Session information for credential processing |
| `PLMN` | Mobile network identifier (MCC/MNC) |
| `ClientInfo` | Client device/platform information |

## Quickstart

```go
package main

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "os"

    "github.com/GlideIdentity/glide-be-sdk-go/core"
)

var (
    clientID     = os.Getenv("GLIDE_CLIENT_ID")
    clientSecret = os.Getenv("GLIDE_CLIENT_SECRET")
    baseURL      = "https://api.glideidentity.app"
)

func main() {
    // Get OAuth2 access token
    accessToken := getAccessToken()

    req := &core.PrepareRequest{
        UseCase:    core.UseCaseGetPhoneNumber,
        PLMN:       &core.PLMN{MCC: "310", MNC: "260"},
        ClientInfo: &core.ClientInfo{Platform: "web"},
    }

    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequest("POST", baseURL+"/magic-auth/v2/auth/prepare", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result core.PrepareResponse
    json.NewDecoder(resp.Body).Decode(&result)

    fmt.Printf("Strategy: %s\n", result.AuthenticationStrategy)
    fmt.Printf("Session: %s\n", result.Session.SessionKey)
}

// getAccessToken fetches an OAuth2 token using client credentials.
// NOTE: Error handling simplified for brevity. In production, handle all errors appropriately.
func getAccessToken() string {
    credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
    req, _ := http.NewRequest("POST", baseURL+"/oauth2-cc/token", 
        bytes.NewBufferString("grant_type=client_credentials"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Authorization", "Basic "+credentials)

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()

    var tokenResp struct {
        AccessToken string `json:"access_token"`
    }
    json.NewDecoder(resp.Body).Decode(&tokenResp)
    return tokenResp.AccessToken
}
```

## Production Usage

Implement the `MagicalAuthService` interface with your HTTP client:

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/GlideIdentity/glide-be-sdk-go/core"
)

type MagicalAuthClient struct {
    baseURL      string
    clientID     string
    clientSecret string
    httpClient   *http.Client
    
    // OAuth2 token cache
    cachedToken     string
    cachedExpiresAt time.Time
    tokenMutex      sync.Mutex
}

func (c *MagicalAuthClient) Prepare(ctx context.Context, req *core.PrepareRequest) (*core.PrepareResponse, error) {
    // Implement HTTP call to /magic-auth/v2/auth/prepare
}

func (c *MagicalAuthClient) VerifyPhoneNumber(ctx context.Context, req *core.VerifyPhoneNumberRequest) (*core.VerifyPhoneNumberResponse, error) {
    // Implement HTTP call to /magic-auth/v2/auth/verify-phone-number
}

func (c *MagicalAuthClient) GetPhoneNumber(ctx context.Context, req *core.GetPhoneNumberRequest) (*core.GetPhoneNumberResponse, error) {
    // Implement HTTP call to /magic-auth/v2/auth/get-phone-number
}

// Verify interface compliance
var _ core.MagicalAuthService = (*MagicalAuthClient)(nil)
```

Then use in your HTTP handlers:

```go
var magicalAuth core.MagicalAuthService = &MagicalAuthClient{...}

func handlePrepare(w http.ResponseWriter, r *http.Request) {
    var req core.PrepareRequest
    json.NewDecoder(r.Body).Decode(&req)

    resp, err := magicalAuth.Prepare(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
    var req core.ProcessRequest
    json.NewDecoder(r.Body).Decode(&req)

    switch req.UseCase {
    case core.UseCaseGetPhoneNumber:
        resp, err := magicalAuth.GetPhoneNumber(r.Context(), &core.GetPhoneNumberRequest{
            Session:    req.Session,
            Credential: req.Credential,
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(resp)

    case core.UseCaseVerifyPhoneNumber:
        resp, err := magicalAuth.VerifyPhoneNumber(r.Context(), &core.VerifyPhoneNumberRequest{
            Session:    req.Session,
            Credential: req.Credential,
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(resp)
    }
}
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /magic-auth/v2/auth/prepare` | Initialize authentication flow |
| `POST /magic-auth/v2/auth/get-phone-number` | Retrieve phone number |
| `POST /magic-auth/v2/auth/verify-phone-number` | Verify phone number |

## License

Use of this SDK is permitted solely to enable your use of Glide Identity Inc.'s services. All other use of this SDK is prohibited. All other rights are reserved.
