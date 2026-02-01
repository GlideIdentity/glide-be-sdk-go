package glide

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

// magicalAuthService implements the MagicalAuthService interface.
type magicalAuthService struct {
	client *Client
}

// newMagicalAuthService creates a new MagicalAuth service.
func newMagicalAuthService(client *Client) MagicalAuthService {
	return &magicalAuthService{
		client: client,
	}
}

// Prepare initiates the authentication flow.
func (s *magicalAuthService) Prepare(ctx context.Context, req *PrepareRequest) (*PrepareResponse, error) {
	if err := s.validatePrepareRequest(req); err != nil {
		return nil, err
	}

	nonce := req.Nonce
	if nonce == "" {
		nonce = generateNonce(32)
	}

	apiReq := map[string]interface{}{
		"nonce": nonce,
	}

	if req.UseCase != "" {
		apiReq["use_case"] = string(req.UseCase)
	}

	if req.PhoneNumber != "" {
		apiReq["phone_number"] = req.PhoneNumber
	}

	if req.PLMN != nil {
		apiReq["plmn"] = map[string]string{
			"mcc": req.PLMN.MCC,
			"mnc": req.PLMN.MNC,
		}
	}

	if req.ConsentData != nil {
		apiReq["consent_data"] = req.ConsentData
	}

	if req.ClientInfo != nil {
		apiReq["client_info"] = req.ClientInfo
	}

	if req.Options != nil {
		apiReq["options"] = req.Options
	}

	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/prepare", apiReq)
	if err != nil {
		return nil, err
	}

	var resp PrepareResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		s.client.logger.Error("Failed to parse response", Field{Key: "error", Value: err.Error()})
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to parse response")
	}

	s.client.logger.Debug("Prepare completed", Field{Key: "strategy", Value: string(resp.AuthenticationStrategy)})
	return &resp, nil
}

// VerifyPhoneNumber verifies a phone number using the credential.
func (s *magicalAuthService) VerifyPhoneNumber(ctx context.Context, req *VerifyPhoneNumberRequest) (*VerifyPhoneNumberResponse, error) {
	if req.Session.SessionKey == "" {
		return nil, NewMagicalAuthError(ErrCodeMissingRequiredField, "Session is required")
	}
	if req.Credential == "" {
		return nil, NewMagicalAuthError(ErrCodeMissingRequiredField, "Credential is required")
	}

	apiReq := map[string]interface{}{
		"session":    req.Session,
		"credential": req.Credential,
	}

	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/verify-phone-number", apiReq)
	if err != nil {
		return nil, err
	}

	var resp VerifyPhoneNumberResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to parse response")
	}

	s.client.logger.Debug("VerifyPhoneNumber completed", Field{Key: "verified", Value: resp.Verified})
	return &resp, nil
}

// GetPhoneNumber retrieves the phone number using the credential.
func (s *magicalAuthService) GetPhoneNumber(ctx context.Context, req *GetPhoneNumberRequest) (*GetPhoneNumberResponse, error) {
	if req.Session.SessionKey == "" {
		return nil, NewMagicalAuthError(ErrCodeMissingRequiredField, "Session is required")
	}
	if req.Credential == "" {
		return nil, NewMagicalAuthError(ErrCodeMissingRequiredField, "Credential is required")
	}

	apiReq := map[string]interface{}{
		"session":    req.Session,
		"credential": req.Credential,
	}

	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/get-phone-number", apiReq)
	if err != nil {
		return nil, err
	}

	var resp GetPhoneNumberResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to parse response")
	}

	s.client.logger.Debug("GetPhoneNumber completed", Field{Key: "phone_number", Value: resp.PhoneNumber})
	return &resp, nil
}

// ReportInvocation reports that an authentication flow was started.
// Returns the server response. For fire-and-forget usage, callers can ignore the response:
//
//	go client.MagicalAuth.ReportInvocation(ctx, req) // fire-and-forget
//	resp, err := client.MagicalAuth.ReportInvocation(ctx, req) // with response
func (s *magicalAuthService) ReportInvocation(ctx context.Context, req *ReportInvocationRequest) (*ReportInvocationResponse, error) {
	if req.SessionID == "" {
		return nil, NewMagicalAuthError(ErrCodeMissingRequiredField, "session_id is required")
	}

	apiReq := map[string]interface{}{
		"session_id": req.SessionID,
	}

	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/report-invocation", apiReq)
	if err != nil {
		s.client.logger.Error("Failed to report invocation", Field{Key: "error", Value: err.Error()}, Field{Key: "session_id", Value: req.SessionID})
		return nil, err
	}

	var resp ReportInvocationResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		s.client.logger.Error("Failed to parse response", Field{Key: "error", Value: err.Error()})
		return nil, NewMagicalAuthError(ErrCodeInternalServerError, "Failed to parse response")
	}

	s.client.logger.Debug("ReportInvocation completed", Field{Key: "session_id", Value: req.SessionID}, Field{Key: "success", Value: resp.Success})
	return &resp, nil
}

func generateNonce(length int) string {
	byteLength := (length*3 + 3) / 4
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	encoded := base64.RawURLEncoding.EncodeToString(bytes)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// validatePrepareRequest validates the prepare request
func (s *magicalAuthService) validatePrepareRequest(req *PrepareRequest) error {
	// Check if we're inheriting from a parent session
	hasParentSession := req.Options != nil && req.Options.ParentSessionID != ""

	if !hasParentSession {
		if req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
			return NewMagicalAuthError(ErrCodeInvalidUseCase, "Invalid use case")
		}

		if err := ValidateUseCaseRequirements(req.UseCase, req.PhoneNumber, req.PLMN); err != nil {
			return err
		}
	} else {
		if req.UseCase != "" && req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
			return NewMagicalAuthError(ErrCodeInvalidUseCase, "Invalid use case")
		}
	}

	if req.PhoneNumber != "" {
		if err := ValidatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}

	if req.PLMN != nil {
		if err := ValidatePLMN(req.PLMN); err != nil {
			return err
		}
	}

	return nil
}
