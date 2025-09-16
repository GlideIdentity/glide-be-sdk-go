# Glide Go SDK Examples

This directory contains example code demonstrating how to use the Glide Go SDK.

## Examples

### 1. Phone Authentication (`phone-auth/`)
Demonstrates the core MagicAuth service for SIM-based phone authentication:
- Verifying a phone number
- Getting a phone number using PLMN
- Error handling patterns
- Understanding the prepare/process flow

```bash
cd phone-auth
go run main.go
```

### 2. Complete Example (`complete/`)
Comprehensive example showing all available Glide services:
- **MagicAuth**: Phone authentication
- **SimSwap**: Detect recent SIM card changes
- **NumberVerify**: Verify phone ownership
- **KYC**: Identity verification matching

```bash
cd complete
go run main.go
```

## Key Concepts

### Authentication Flow
1. **Prepare**: Server determines the authentication strategy
2. **Invoke**: Client performs authentication (browser/app)
3. **Process**: Server verifies the credential

### Strategies
- **TS43**: Digital Credentials API (Android)
- **Link**: Deep links/App clips (iOS/Android)

The backend decides the strategy based on carrier and OS - the SDK handles both transparently.

### Error Handling
All errors are typed with specific codes:
- Check error codes to determine the issue
- Use `IsRetryable()` for automatic retry logic
- No carrier information is exposed in errors

## Running the Examples

1. Get your API key from [Glide Dashboard](https://dashboard.glideidentity.app)
2. Replace `"your-api-key"` in the examples
3. Run the example:
   ```bash
   cd examples/phone-auth
   go run main.go
   ```

## Note on Client Authentication

In these examples, we simulate the client authentication step. In production:
1. Your backend calls `Prepare()`
2. Send the response to your frontend
3. Frontend uses Web SDK to perform authentication
4. Frontend sends credential back to your backend
5. Your backend calls `VerifyPhoneNumber()` or `GetPhoneNumber()`
