package glide

import (
	"context"
	"encoding/json"
)

// simSwapService implements the SimSwapService interface
type simSwapService struct {
	client *Client
}

// newSimSwapService creates a new SimSwap service
func newSimSwapService(client *Client) SimSwapService {
	return &simSwapService{
		client: client,
	}
}

// Check verifies if a SIM swap occurred recently
func (s *simSwapService) Check(ctx context.Context, req *SimSwapCheckRequest) (*SimSwapCheckResponse, error) {
	// Validate request
	if req.PhoneNumber == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number is required")
	}

	if !isValidE164(req.PhoneNumber) {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number must be in E.164 format")
	}

	// Default max age to 24 hours if not specified
	maxAge := req.MaxAge
	if maxAge == 0 {
		maxAge = 24
	}

	// Build API request
	apiReq := map[string]interface{}{
		"phone_number":  req.PhoneNumber,
		"max_age_hours": maxAge,
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/sim-swap/check", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp SimSwapCheckResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}

// GetLastSwapDate retrieves the last SIM swap date
func (s *simSwapService) GetLastSwapDate(ctx context.Context, req *SimSwapDateRequest) (*SimSwapDateResponse, error) {
	// Validate request
	if req.PhoneNumber == "" {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number is required")
	}

	if !isValidE164(req.PhoneNumber) {
		return nil, NewError(ErrCodeInvalidParameters, "Phone number must be in E.164 format")
	}

	// Build API request
	apiReq := map[string]interface{}{
		"phone_number": req.PhoneNumber,
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/sim-swap/last-swap-date", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp SimSwapDateResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}
