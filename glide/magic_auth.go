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

	// Flatten PLMN into mcc and mnc at top level
	if req.PLMN != nil {
		apiReq["mcc"] = req.PLMN.MCC
		apiReq["mnc"] = req.PLMN.MNC
	}

	if req.ConsentData != nil {
		apiReq["consent_data"] = req.ConsentData
	}

	// Debug logging
	if s.client.logger != nil {
		reqBytes, _ := json.Marshal(apiReq)
		s.client.logger.Debug("Prepare request", Field{Key: "body", Value: string(reqBytes)})
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/prep", apiReq)
	if err != nil {
		return nil, err
	}

	// Debug logging for response
	if s.client.logger != nil {
		s.client.logger.Debug("Prepare response", Field{Key: "body", Value: string(respData)})
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
	if req.SessionInfo == nil || req.SessionInfo.SessionKey == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Session info with session key is required")
	}

	if req.Credential == nil {
		return nil, NewError(ErrCodeInvalidParameters, "Credential is required")
	}

	// Build API request
	// Extract the actual SD-JWT string from the vp_token object
	var ts43dcString string
	if vpTokenWrapper, ok := req.Credential["vp_token"]; ok {
		switch vt := vpTokenWrapper.(type) {
		case string:
			// If vp_token is already a string, use it directly
			ts43dcString = vt
		case map[string]interface{}:
			// vp_token is an object like {"glide": "actual_jwt_string"}
			// Extract the first value (the actual JWT)
			for _, value := range vt {
				if str, ok := value.(string); ok {
					ts43dcString = str
					break
				}
			}
		default:
			// Fallback: JSON encode the vp_token
			if encoded, err := json.Marshal(vt); err == nil {
				ts43dcString = string(encoded)
			}
		}
	}

	// If we still don't have a string, JSON encode the entire credential
	if ts43dcString == "" {
		if encoded, err := json.Marshal(req.Credential); err == nil {
			ts43dcString = string(encoded)
		}
	}

	apiReq := map[string]interface{}{
		"session": req.SessionInfo,
		"ts43_dc": ts43dcString, // Backend expects ts43_dc as a string
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
	if req.SessionInfo == nil || req.SessionInfo.SessionKey == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Session info with session key is required")
	}

	if req.Credential == nil {
		return nil, NewError(ErrCodeInvalidParameters, "Credential is required")
	}

	// Build API request
	// Extract the actual SD-JWT string from the vp_token object
	var ts43dcString string
	if vpTokenWrapper, ok := req.Credential["vp_token"]; ok {
		switch vt := vpTokenWrapper.(type) {
		case string:
			// If vp_token is already a string, use it directly
			ts43dcString = vt
		case map[string]interface{}:
			// vp_token is an object like {"glide": "actual_jwt_string"}
			// Extract the first value (the actual JWT)
			for _, value := range vt {
				if str, ok := value.(string); ok {
					ts43dcString = str
					break
				}
			}
		default:
			// Fallback: JSON encode the vp_token
			if encoded, err := json.Marshal(vt); err == nil {
				ts43dcString = string(encoded)
			}
		}
	}

	// If we still don't have a string, JSON encode the entire credential
	if ts43dcString == "" {
		if encoded, err := json.Marshal(req.Credential); err == nil {
			ts43dcString = string(encoded)
		}
	}

	apiReq := map[string]interface{}{
		"session": req.SessionInfo,
		"ts43_dc": ts43dcString, // Backend expects ts43_dc as a string
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

// ProcessCredential is deprecated. Use VerifyPhoneNumber or GetPhoneNumber instead.
// Deprecated: This method will be removed in v2.0.0
func (s *magicAuthService) ProcessCredential(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	// For backward compatibility, default to GetPhoneNumber behavior
	// Parse session string as just the session key
	getReq := &GetPhoneNumberRequest{
		SessionInfo: &SessionInfo{
			SessionKey: req.Session,
		},
		Credential: req.Response,
	}

	getResp, err := s.GetPhoneNumber(ctx, getReq)
	if err != nil {
		return nil, err
	}

	// Map to old response type
	return &ProcessResponse{
		PhoneNumber: getResp.PhoneNumber,
		Verified:    false, // GetPhoneNumber doesn't have verified field
	}, nil
}

// generateNonce generates a random base64url-encoded nonce
func generateNonce(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

// validatePrepareRequest validates the prepare request
func (s *magicAuthService) validatePrepareRequest(req *PrepareRequest) error {
	// Validate use case
	if req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
		return NewError(ErrCodeInvalidParameters, "Invalid use case")
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
