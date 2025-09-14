package glide

import (
	"context"
	"encoding/json"
)

// kycService implements the KYCService interface
type kycService struct {
	client *Client
}

// newKYCService creates a new KYC service
func newKYCService(client *Client) KYCService {
	return &kycService{
		client: client,
	}
}

// Match verifies user identity information
func (s *kycService) Match(ctx context.Context, req *KYCMatchRequest) (*KYCMatchResponse, error) {
	// Validate request
	if req.PhoneNumber == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number is required")
	}

	if !isValidE164(req.PhoneNumber) {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number must be in E.164 format")
	}

	// At least one field besides phone number should be provided for matching
	if req.Name == "" && req.GivenName == "" && req.FamilyName == "" &&
		req.BirthDate == "" && req.Email == "" && req.Address == nil && req.IDDocument == "" {
		return nil, NewError(ErrCodeInvalidParameters, "At least one field to match is required")
	}

	// Validate birth date format if provided
	if req.BirthDate != "" && !isValidDateFormat(req.BirthDate) {
		return nil, NewError(ErrCodeInvalidParameters, "Birth date must be in YYYY-MM-DD format")
	}

	// Build API request - only include non-empty fields
	apiReq := map[string]interface{}{
		"phone_number": req.PhoneNumber,
	}

	if req.Name != "" {
		apiReq["name"] = req.Name
	}
	if req.GivenName != "" {
		apiReq["given_name"] = req.GivenName
	}
	if req.FamilyName != "" {
		apiReq["family_name"] = req.FamilyName
	}
	if req.BirthDate != "" {
		apiReq["birth_date"] = req.BirthDate
	}
	if req.Email != "" {
		apiReq["email"] = req.Email
	}
	if req.Address != nil {
		apiReq["address"] = req.Address
	}
	if req.IDDocument != "" {
		apiReq["id_document"] = req.IDDocument
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/kyc/match", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp KYCMatchResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// isValidDateFormat checks if a date string is in YYYY-MM-DD format
func isValidDateFormat(date string) bool {
	// Simple validation for YYYY-MM-DD format
	if len(date) != 10 {
		return false
	}

	if date[4] != '-' || date[7] != '-' {
		return false
	}

	// Check year, month, day are digits
	for i, c := range date {
		if i == 4 || i == 7 {
			continue // Skip hyphens
		}
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
