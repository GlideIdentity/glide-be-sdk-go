package glide

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

// magicAuthService implements the MagicAuthService interface
type magicAuthService struct {
	client *Client
}

// newMagicAuthService creates a new MagicAuth service
func newMagicAuthService(client *Client) MagicAuthService {
	return &magicAuthService{
		client: client,
	}
}

// Prepare initiates the authentication flow
func (s *magicAuthService) Prepare(ctx context.Context, req *PrepareRequest) (*PrepareResponse, error) {
	// Validate request
	if err := s.validatePrepareRequest(req); err != nil {
		return nil, err
	}

	// Generate nonce (random string for request identification)
	nonce := generateNonce(32)

	// Build API request
	apiReq := map[string]interface{}{
		"nonce":    nonce,
		"id":       "glide", // Aggregator ID
		"use_case": string(req.UseCase),
	}

	if req.PhoneNumber != "" {
		apiReq["phone_number"] = req.PhoneNumber
	}

	// Add PLMN as nested object to match Node.js SDK structure
	if req.PLMN != nil {
		apiReq["plmn"] = map[string]string{
			"mcc": req.PLMN.MCC,
			"mnc": req.PLMN.MNC,
		}
	}

	if req.ConsentData != nil {
		apiReq["consent_data"] = req.ConsentData
	}

	// Add client info if provided
	if req.ClientInfo != nil {
		apiReq["client_info"] = req.ClientInfo
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/prepare", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp PrepareResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		s.client.logger.Error("Failed to parse response", Field{Key: "error", Value: err.Error()})
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	// Store the use case so we know which endpoint to call later
	resp.UseCase = req.UseCase

	return &resp, nil
}

// VerifyPhoneNumber verifies a phone number using the credential from Digital Credentials API
func (s *magicAuthService) VerifyPhoneNumber(ctx context.Context, req *VerifyPhoneNumberRequest) (*VerifyPhoneNumberResponse, error) {
	// Validate request
	if req.Session == nil {
		return nil, NewError(ErrCodeMissingParameters, "Session is required")
	}
	if req.Credential == nil {
		return nil, NewError(ErrCodeMissingParameters, "Credential is required")
	}

	// Build API request - pass through what the client sent
	// Just like the Node SDK, we pass the session and credential directly
	apiReq := map[string]interface{}{
		"session":    req.Session,
		"credential": s.extractCredentialString(req.Credential),
	}

	// Call the verify endpoint
	endpoint := "/magic-auth/v2/auth/verify-phone-number"

	respData, err := s.client.doRequest(ctx, "POST", endpoint, apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp VerifyPhoneNumberResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// GetPhoneNumber retrieves the phone number using the credential from Digital Credentials API
func (s *magicAuthService) GetPhoneNumber(ctx context.Context, req *GetPhoneNumberRequest) (*GetPhoneNumberResponse, error) {
	// Validate request
	if req.Session == nil {
		return nil, NewError(ErrCodeMissingParameters, "Session is required")
	}
	if req.Credential == nil {
		return nil, NewError(ErrCodeMissingParameters, "Credential is required")
	}

	// Build API request - pass through what the client sent
	// Just like the Node SDK, we pass the session and credential directly
	apiReq := map[string]interface{}{
		"session":    req.Session,
		"credential": s.extractCredentialString(req.Credential),
	}

	// Call the get phone number endpoint
	endpoint := "/magic-auth/v2/auth/get-phone-number"

	respData, err := s.client.doRequest(ctx, "POST", endpoint, apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp GetPhoneNumberResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// generateNonce generates a random base64url-encoded nonce
func generateNonce(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

// extractCredentialString extracts the credential string from various formats
// The client SDK sends the credential as a JWT string directly
func (s *magicAuthService) extractCredentialString(credential interface{}) string {
	// If it's already a string, use it directly
	if str, ok := credential.(string); ok {
		return str
	}

	// If it's raw JSON containing a string, unmarshal it
	if jsonBytes, ok := credential.(json.RawMessage); ok {
		var credStr string
		if err := json.Unmarshal(jsonBytes, &credStr); err == nil {
			return credStr
		}
	}

	// If it's a map with vp_token field (legacy format)
	if credMap, ok := credential.(map[string]interface{}); ok {
		if vpToken, exists := credMap["vp_token"]; exists {
			if vpStr, ok := vpToken.(string); ok {
				return vpStr
			}
		}
	}

	// Fallback: JSON encode the credential
	if encoded, err := json.Marshal(credential); err == nil {
		return string(encoded)
	}

	return ""
}

// validatePrepareRequest validates the prepare request
func (s *magicAuthService) validatePrepareRequest(req *PrepareRequest) error {
	// Validate use case
	if req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
		return NewError(ErrCodeValidationError, "Invalid use case")
	}

	// Validate use case requirements (handles the business logic)
	if err := ValidateUseCaseRequirements(req.UseCase, req.PhoneNumber, req.PLMN); err != nil {
		return err
	}

	// Validate phone number format if provided
	if req.PhoneNumber != "" {
		if err := ValidatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}

	// Validate PLMN format if provided
	if req.PLMN != nil {
		if err := ValidatePLMN(req.PLMN); err != nil {
			return err
		}
	}

	// Validate consent data if provided
	if req.ConsentData != nil {
		if err := ValidateConsentData(req.ConsentData); err != nil {
			return err
		}
	}

	return nil
}
