package core

import "context"

// MagicalAuthService defines the interface for SIM-based phone authentication.
type MagicalAuthService interface {
	// Prepare initiates the authentication flow and returns strategy-specific data.
	Prepare(ctx context.Context, req *PrepareRequest) (*PrepareResponse, error)

	// VerifyPhoneNumber verifies the phone number matches the device.
	VerifyPhoneNumber(ctx context.Context, req *VerifyPhoneNumberRequest) (*VerifyPhoneNumberResponse, error)

	// GetPhoneNumber retrieves the phone number from the device.
	GetPhoneNumber(ctx context.Context, req *GetPhoneNumberRequest) (*GetPhoneNumberResponse, error)

	// ReportInvocation reports that an authentication flow was started.
	ReportInvocation(ctx context.Context, req *ReportInvocationRequest) (*ReportInvocationResponse, error)
}
