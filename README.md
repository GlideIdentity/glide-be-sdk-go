# Glide Go SDK

Official Go SDK for Glide Identity's authentication and verification services.

## Features

- 🔐 **Magic Auth**: Carrier-based phone authentication using Digital Credentials API (TS43)
- 📱 **Phone Number Verification**: Verify phone numbers without SMS/OTP
- 🔍 **Phone Number Retrieval**: Get the device's phone number securely
- 🌐 **Carrier Detection**: Automatic carrier identification and eligibility checking
- 🔒 **Privacy-First**: No phone numbers exposed in API responses unless verified

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
    
    "github.com/glideidentity/glide-go-sdk"
)

func main() {
    // Initialize client
    client := glide.New(
        glide.WithAPIKey("your-api-key"),
    )
    
    ctx := context.Background()
    
    // Step 1: Prepare authentication
    prepareResp, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
        PhoneNumber: "+1234567890",
        UseCase:     glide.UseCaseVerifyPhoneNumber,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Step 2: Client gets credential via Digital Credentials API
    // (This happens in the browser)
    
    // Step 3: Verify the phone number
    verifyResp, err := client.MagicAuth.VerifyPhoneNumber(ctx, &glide.VerifyPhoneNumberRequest{
        Session:    prepareResp.Session,
        Credential: credentialResponse, // From browser
    })
    if err != nil {
        log.Fatal(err)
    }
    
    if verifyResp.Verified {
        log.Printf("Phone verified: %s", verifyResp.PhoneNumber)
    }
}
```

## API Reference

### Magic Auth Service

The Magic Auth service provides carrier-based phone authentication without SMS.

#### Prepare Authentication

```go
prepareResp, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
    // Option 1: Use phone number
    PhoneNumber: "+1234567890",
    
    // Option 2: Use PLMN (Mobile Country Code + Mobile Network Code)
    PLMN: &glide.PLMN{
        MCC: "310",
        MNC: "260",
    },
    
    UseCase: glide.UseCaseGetPhoneNumber, // or UseCaseVerifyPhoneNumber
    
    // Optional: Include client info for better strategy selection
    ClientInfo: &glide.ClientInfo{
        UserAgent: req.Header.Get("User-Agent"),
        Platform:  "web",
    },
})
```

#### Get Phone Number

```go
result, err := client.MagicAuth.GetPhoneNumber(ctx, &glide.GetPhoneNumberRequest{
    Session:    prepareResp.Session,
    Credential: credentialFromBrowser,
})

if err == nil {
    fmt.Printf("Phone number: %s\n", result.PhoneNumber)
}
```

#### Verify Phone Number

```go
result, err := client.MagicAuth.VerifyPhoneNumber(ctx, &glide.VerifyPhoneNumberRequest{
    Session:    prepareResp.Session,
    Credential: credentialFromBrowser,
})

if err == nil && result.Verified {
    fmt.Printf("Phone %s verified!\n", result.PhoneNumber)
}
```

## Error Handling

The SDK provides typed errors with detailed information:

```go
if err != nil {
    if glideErr, ok := err.(*glide.Error); ok {
        log.Printf("Error Code: %s", glideErr.Code)
        log.Printf("Status: %d", glideErr.Status)
        log.Printf("Message: %s", glideErr.Message)
        log.Printf("Request ID: %s", glideErr.RequestID)
        
        switch glideErr.Code {
        case glide.ErrCodeCarrierNotEligible:
            // Carrier doesn't support TS43, use fallback
        case glide.ErrCodeSessionNotFound:
            // Session expired, restart flow
        case glide.ErrCodeValidationError:
            // Invalid input parameters
        case glide.ErrCodeInvalidCredentialFormat:
            // Credential validation failed
        default:
            // Handle other errors
        }
    }
}
```

### Common Error Codes

- `CARRIER_NOT_ELIGIBLE` - Carrier doesn't support the service
- `SESSION_NOT_FOUND` - Session expired or invalid
- `VALIDATION_ERROR` - Invalid input parameters  
- `INVALID_CREDENTIAL_FORMAT` - Invalid or tampered credential
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `SERVICE_UNAVAILABLE` - Temporary service outage
- `UNSUPPORTED_PLATFORM` - Browser/platform not supported

## Configuration

```go
client := glide.New(
    glide.WithAPIKey("your-api-key"),
    glide.WithBaseURL("https://api.glideidentity.app"),
    glide.WithTimeout(30 * time.Second),
    glide.WithRetry(3, time.Second),
    glide.WithLogLevel(glide.LogLevelDebug),
    glide.WithLogFormat(glide.LogFormatPretty), // Options: LogFormatPretty, LogFormatJSON, LogFormatSimple
)
```

### Environment Variables

The SDK also supports configuration via environment variables:

- `GLIDE_API_KEY` - API key for authentication
- `GLIDE_BASE_URL` - API base URL
- `GLIDE_LOG_LEVEL` - Log level (debug, info, warn, error)
- `GLIDE_LOG_FORMAT` - Log output format (pretty, json, simple)

## Security Best Practices

- **API Keys**: Never expose API keys in client-side code
- **Credential Validation**: Always validate credentials on your backend
- **HTTPS Only**: The SDK enforces secure connections
- **Session Expiry**: Sessions are automatically invalidated after use
- **Phone Number Privacy**: Phone numbers are never logged in full

## License

This SDK is proprietary software. All rights reserved.
