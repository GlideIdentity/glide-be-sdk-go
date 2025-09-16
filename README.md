# Glide Go SDK

Official Go SDK for Glide Identity's authentication and verification services.

## Features

- 🔐 **Magic Auth**: SIM-based phone authentication (TS43 & Link strategies)
- 🔄 **SIM Swap Detection**: Check for recent SIM card changes
- 📱 **Number Verification**: Verify phone number ownership
- 👤 **KYC Match**: Identity verification services

## Installation

```bash
go get github.com/glideidentity/glide-go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/glideidentity/glide-go-sdk/glide"
)

func main() {
    // Initialize client
    client := glide.New(
        glide.WithAPIKey("your-api-key"),
    )
    
    // Use MagicAuth service
    ctx := context.Background()
    
    // Prepare authentication
    prepareResp, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
        PhoneNumber: "+1234567890",
        UseCase:     glide.UseCaseVerifyPhoneNumber,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Process credential (after client-side authentication)
    processResp, err := client.MagicAuth.ProcessCredential(ctx, &glide.ProcessRequest{
        Session:  prepareResp.Session.SessionKey,
        Response: credentialResponse,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Phone verified: %s", processResp.PhoneNumber)
}
```

## Examples

See the [`examples/`](examples/) directory for complete working examples:
- [`phone-auth/`](examples/phone-auth/) - Phone authentication with MagicAuth
- [`complete/`](examples/complete/) - All services demonstration

## Services

### Magic Auth
SIM-based phone authentication using carrier verification.

```go
// Use case constants
glide.UseCaseGetPhoneNumber    // "GetPhoneNumber"
glide.UseCaseVerifyPhoneNumber // "VerifyPhoneNumber"

// Authentication strategy constants
glide.AuthenticationStrategyTS43 // "ts43" - Digital Credentials API
glide.AuthenticationStrategyLink // "link" - Deep links/App clips
```

### SIM Swap
Detect recent SIM card changes for fraud prevention.

### Number Verify
Verify phone number ownership through carrier lookup.

### KYC Match
Match user information for identity verification.

## Error Handling

The SDK uses typed errors for better error handling:

```go
if err != nil {
    if glideErr, ok := err.(*glide.Error); ok {
        switch glideErr.Code {
        case glide.ErrCodeCarrierNotEligible:
            // Handle carrier not eligible
        case glide.ErrCodeSessionExpired:
            // Handle session expiration
        default:
            // Handle other errors
        }
    }
}
```

## Configuration

```go
client := glide.New(
    glide.WithAPIKey("key"),
    glide.WithBaseURL("https://api.glideidentity.app"),
    glide.WithTimeout(30 * time.Second),
    glide.WithRetry(3, time.Second),
    glide.WithRateLimit(100, time.Second),  // Optional rate limiting
)
```

## License

MIT License - see LICENSE file for details.
