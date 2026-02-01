package glide

import (
	"context"

	"github.com/GlideIdentity/glide-be-sdk-go/core/v2"
)

// Re-export core types.
type (
	UseCase                   = core.UseCase
	AuthenticationStrategy    = core.AuthenticationStrategy
	PLMN                      = core.PLMN
	ConsentData               = core.ConsentData
	ClientInfo                = core.ClientInfo
	PrepareOptions            = core.PrepareOptions
	SessionInfo               = core.SessionInfo
	PrepareRequest            = core.PrepareRequest
	PrepareResponse           = core.PrepareResponse
	ProcessRequest            = core.ProcessRequest
	VerifyPhoneNumberRequest  = core.VerifyPhoneNumberRequest
	VerifyPhoneNumberResponse = core.VerifyPhoneNumberResponse
	GetPhoneNumberRequest     = core.GetPhoneNumberRequest
	GetPhoneNumberResponse    = core.GetPhoneNumberResponse
	ReportInvocationRequest   = core.ReportInvocationRequest
	ReportInvocationResponse  = core.ReportInvocationResponse

	// Strategy-specific data types from PrepareResponse.data
	TS43Data           = core.TS43Data
	TS43InnerData      = core.TS43InnerData
	TS43DCQLQuery      = core.TS43DCQLQuery
	TS43Credential     = core.TS43Credential
	TS43CredentialMeta = core.TS43CredentialMeta
	LinkData           = core.LinkData
	DesktopData        = core.DesktopData
	DesktopInnerData   = core.DesktopInnerData
	Challenge          = core.Challenge
)

// Re-export core constants.
const (
	UseCaseGetPhoneNumber    = core.UseCaseGetPhoneNumber
	UseCaseVerifyPhoneNumber = core.UseCaseVerifyPhoneNumber

	AuthenticationStrategyTS43    = core.AuthenticationStrategyTS43
	AuthenticationStrategyLink    = core.AuthenticationStrategyLink
	AuthenticationStrategyDesktop = core.AuthenticationStrategyDesktop
)

// MagicalAuthService handles SIM-based phone authentication.
type MagicalAuthService interface {
	Prepare(ctx context.Context, req *PrepareRequest) (*PrepareResponse, error)
	VerifyPhoneNumber(ctx context.Context, req *VerifyPhoneNumberRequest) (*VerifyPhoneNumberResponse, error)
	GetPhoneNumber(ctx context.Context, req *GetPhoneNumberRequest) (*GetPhoneNumberResponse, error)
	ReportInvocation(ctx context.Context, req *ReportInvocationRequest) (*ReportInvocationResponse, error)
}

// ============================================================
// Helper Functions
// ============================================================

// GetStatusURL extracts the status URL from strategy-specific data.
// Works for both Link and Desktop strategies.
// Returns empty string for TS43 strategy (doesn't use polling).
func GetStatusURL(resp *PrepareResponse) string {
	if resp == nil || resp.Data == nil {
		return ""
	}

	switch resp.AuthenticationStrategy {
	case AuthenticationStrategyLink:
		if statusURL, ok := resp.Data["status_url"].(string); ok {
			return statusURL
		}
	case AuthenticationStrategyDesktop:
		if innerData, ok := resp.Data["data"].(map[string]interface{}); ok {
			if statusURL, ok := innerData["status_url"].(string); ok {
				return statusURL
			}
		}
	}

	return ""
}

// GetDesktopStatusURL returns the status URL from DesktopData.
func GetDesktopStatusURL(d *DesktopData) string {
	if d != nil && d.Data != nil {
		return d.Data.StatusURL
	}
	return ""
}
