package magicalauth

import (
	"fmt"
	"strings"
)

// validateEligibilityRequest validates an EligibilityRequest before sending to the API.
// At least one of phone_number or plmn is required, and client_info.user_agent is
// required so the server can pick the right authentication strategy for the device.
func (c *Client) validateEligibilityRequest(req *EligibilityRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.PhoneNumber == "" && req.PLMN == nil {
		return fmt.Errorf("%w: phone_number or plmn is required", ErrValidation)
	}
	if req.ClientInfo == nil || strings.TrimSpace(req.ClientInfo.UserAgent) == "" {
		return fmt.Errorf("%w: client_info.user_agent is required for eligibility", ErrValidation)
	}
	return nil
}

// validatePrepareRequest validates a PrepareRequest before sending to the API.
// Returns ErrMissingUseCase if use_case is empty and no parent_session_id is set
// (child sessions in the desktop-to-mobile flow inherit use_case from the parent).
func (c *Client) validatePrepareRequest(req *PrepareRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	hasParentSession := req.Options != nil && req.Options.ParentSessionID != ""
	if req.UseCase == "" && !hasParentSession {
		return ErrMissingUseCase
	}
	if req.UseCase == UseCaseVerifyPhoneNumber && req.PhoneNumber == "" {
		return ErrMissingPhoneNum
	}
	return nil
}

// validateVerifyRequest validates a VerifyPhoneNumberRequest.
// Requires session.session_key and credential.
func (c *Client) validateVerifyRequest(req *VerifyPhoneNumberRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.Session.SessionKey == "" {
		return ErrMissingSessionKey
	}
	if req.Credential == "" {
		return ErrMissingCredential
	}
	return nil
}

// validateGetPhoneRequest validates a GetPhoneNumberRequest.
// Requires session.session_key and credential.
func (c *Client) validateGetPhoneRequest(req *GetPhoneNumberRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.Session.SessionKey == "" {
		return ErrMissingSessionKey
	}
	if req.Credential == "" {
		return ErrMissingCredential
	}
	return nil
}

// validateReportRequest validates a ReportInvocationRequest.
// Requires session_id.
func (c *Client) validateReportRequest(req *ReportInvocationRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.SessionID == "" {
		return ErrMissingSessionID
	}
	return nil
}

// validateCompleteRequest validates a CompleteRequest for device-bound sessions.
// Returns specific error codes per field:
//   - session_key missing → ErrValidation (400)
//   - fe_code missing → ErrMissingBindingCookie (403)
//   - agg_code missing → ErrValidation (400)
//   - format invalid → ErrDeviceBindingFailed (403)
func (c *Client) validateCompleteRequest(req *CompleteRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if req.SessionKey == "" {
		return &APIError{
			Err:     ErrValidation,
			Code:    CodeValidationError,
			Message: "session_key is required",
			Status:  400,
		}
	}
	if req.FeCode == "" {
		return &APIError{
			Err:     ErrMissingBindingCookie,
			Code:    CodeMissingBindingCookie,
			Message: "fe_code is required (from _glide_bind cookie)",
			Status:  403,
		}
	}
	if req.AggCode == "" {
		return &APIError{
			Err:     ErrValidation,
			Code:    CodeValidationError,
			Message: "agg_code is required",
			Status:  400,
		}
	}
	if !IsValidBindingCode(req.FeCode) || !IsValidBindingCode(req.AggCode) {
		return &APIError{
			Err:     ErrDeviceBindingFailed,
			Code:    CodeDeviceBindingFailed,
			Message: "invalid device binding inputs",
			Status:  403,
		}
	}
	return nil
}
