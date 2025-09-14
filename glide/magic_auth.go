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
	// Must have either phone number or PLMN
	if req.PhoneNumber == "" && req.PLMN == nil {
		return NewError(ErrCodeInvalidParameters, "Either phone_number or PLMN is required")
	}

	// Validate use case
	if req.UseCase != UseCaseGetPhoneNumber && req.UseCase != UseCaseVerifyPhoneNumber {
		return NewError(ErrCodeInvalidParameters, "Invalid use case")
	}

	// VerifyPhoneNumber requires a phone number
	if req.UseCase == UseCaseVerifyPhoneNumber && req.PhoneNumber == "" {
		return NewError(ErrCodeInvalidParameters, "Phone number is required for VerifyPhoneNumber")
	}

	// Validate PLMN if provided
	if req.PLMN != nil {
		if req.PLMN.MCC == "" || req.PLMN.MNC == "" {
			return NewError(ErrCodeInvalidParameters, "Both MCC and MNC are required for PLMN")
		}
	}

	// Validate phone number format if provided
	if req.PhoneNumber != "" && !isValidE164(req.PhoneNumber) {
		return NewError(ErrCodeInvalidParameters, "Phone number must be in E.164 format")
	}

	return nil
}

// isValidE164 checks if a phone number is in E.164 format
func isValidE164(phoneNumber string) bool {
	// Basic E.164 validation
	if len(phoneNumber) < 2 || phoneNumber[0] != '+' {
		return false
	}

	// Check if rest are digits
	for i := 1; i < len(phoneNumber); i++ {
		if phoneNumber[i] < '0' || phoneNumber[i] > '9' {
			return false
		}
	}

	// E.164 numbers are max 15 digits (plus the +)
	if len(phoneNumber) > 16 {
		return false
	}

	return true
}
