package glide

import (
	"context"
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

	// Build API request
	apiReq := map[string]interface{}{
		"use_case": req.UseCase,
	}

	if req.PhoneNumber != "" {
		apiReq["phone_number"] = req.PhoneNumber
	}

	if req.PLMN != nil {
		apiReq["plmn"] = req.PLMN
	}

	if req.ConsentData != nil {
		apiReq["consent_data"] = req.ConsentData
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/prepare", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp PrepareResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// ProcessCredential processes the authentication response
func (s *magicAuthService) ProcessCredential(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	// Validate request
	if req.Session == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Session is required")
	}

	if req.Response == nil {
		return nil, NewError(ErrCodeInvalidParameters, "Response is required")
	}

	// Build API request
	apiReq := map[string]interface{}{
		"session":  req.Session,
		"response": req.Response,
	}

	if req.PhoneNumber != "" {
		apiReq["phone_number"] = req.PhoneNumber
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/magic-auth/v2/auth/process", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp ProcessResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// validatePrepareRequest validates the prepare request
func (s *magicAuthService) validatePrepareRequest(req *PrepareRequest) error {
	// Validate use case
	if req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
		return NewError(ErrCodeInvalidParameters, "Invalid use case")
	}

	// Validate use case requirements first
	if err := ValidateUseCaseRequirements(req.UseCase, req.PhoneNumber); err != nil {
		return err
	}

	// Validate phone number if provided
	if err := ValidatePhoneNumber(req.PhoneNumber); err != nil {
		return err
	}

	// Validate PLMN if provided
	if err := ValidatePLMN(req.PLMN); err != nil {
		return err
	}

	// Either phone number or PLMN must be provided
	if req.PhoneNumber == "" && req.PLMN == nil {
		return NewError(ErrCodeMissingParameters, "Either phone number or PLMN (MCC/MNC) must be provided")
	}

	return nil
}
