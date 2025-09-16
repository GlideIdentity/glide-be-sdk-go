# Go SDK Test Server

This is a test server for integration testing the Glide Go SDK. It uses the Go SDK itself to make calls to the Glide API and exposes endpoints that can be used for testing.

## Quick Start

```bash
# Install dependencies
go mod download

# Run the server
GLIDE_API_KEY=your-api-key go run server.go

# Or with debug mode
GLIDE_DEBUG=true GLIDE_API_KEY=your-api-key go run server.go
```

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `PORT` | Server port | `3001` |
| `GLIDE_API_KEY` | Your Glide API key | Required |
| `GLIDE_API_BASE_URL` | Glide API base URL | `https://api.glideidentity.app` |
| `GLIDE_AUTH_BASE_URL` | Glide Auth base URL | `https://oidc.gateway-x.io` |
| `ALLOWED_ORIGIN` | CORS allowed origin | `http://localhost:3000` |
| `GLIDE_DEBUG` | Enable debug logging | `false` |
| `GLIDE_LOG_LEVEL` | Log level (debug, info, warn, error) | `error` |

## Endpoints

### Prepare Authentication
```bash
POST /api/phone-auth/prepare
```

Request:
```json
{
  "use_case": "VerifyPhoneNumber",
  "phone_number": "+14155552671"
}
```

### Process Credential
```bash
POST /api/phone-auth/process
```

Request:
```json
{
  "session": "session_key_from_prepare",
  "response": {
    "vp_token": "credential_token"
  },
  "phoneNumber": "+14155552671"
}
```

## Testing

Run integration tests against this server:

```bash
# Start the server in one terminal
GLIDE_API_KEY=your-key go run server.go

# Run tests in another terminal
cd ../examples/phone-auth
go run main.go
```
