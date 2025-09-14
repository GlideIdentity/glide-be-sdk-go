package glide

import (
	"context"
	"encoding/json"
)

// numberVerifyService implements the NumberVerifyService interface
type numberVerifyService struct {
	client *Client
}

// newNumberVerifyService creates a new NumberVerify service
func newNumberVerifyService(client *Client) NumberVerifyService {
	return &numberVerifyService{
		client: client,
	}
}

// Verify checks if a phone number belongs to the user
func (s *numberVerifyService) Verify(ctx context.Context, req *NumberVerifyRequest) (*NumberVerifyResponse, error) {
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

	// Add code if provided (for code-based verification)
	if req.Code != "" {
		apiReq["code"] = req.Code
	}

	// Make API call
	respData, err := s.client.doRequest(ctx, "POST", "/number-verify/verify", apiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp NumberVerifyResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, NewError(ErrCodeInternalServerError, "Failed to parse response")
	}

	return &resp, nil
}
