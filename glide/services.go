package glide

import (
	"context"
)

// UseCase represents the authentication use case
type UseCase string

const (
	UseCaseGetPhoneNumber    UseCase = "GetPhoneNumber"
	UseCaseVerifyPhoneNumber UseCase = "VerifyPhoneNumber"
)

// AuthenticationStrategy represents the authentication method
type AuthenticationStrategy string

const (
	AuthenticationStrategyTS43 AuthenticationStrategy = "ts43"
	AuthenticationStrategyLink AuthenticationStrategy = "link"
)

// MagicAuthService handles SIM-based phone authentication
type MagicAuthService interface {
	// Prepare initiates the authentication flow
	Prepare(ctx context.Context, req *PrepareRequest) (*PrepareResponse, error)

	// ProcessCredential processes the authentication response
	ProcessCredential(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error)
}

// SimSwapService handles SIM swap detection
type SimSwapService interface {
	// Check verifies if a SIM swap occurred recently
	Check(ctx context.Context, req *SimSwapCheckRequest) (*SimSwapCheckResponse, error)

	// GetLastSwapDate retrieves the last SIM swap date
	GetLastSwapDate(ctx context.Context, req *SimSwapDateRequest) (*SimSwapDateResponse, error)
}

// NumberVerifyService handles number verification
type NumberVerifyService interface {
	// Verify checks if a phone number belongs to the user
	Verify(ctx context.Context, req *NumberVerifyRequest) (*NumberVerifyResponse, error)
}

// KYCService handles KYC (Know Your Customer) verification
type KYCService interface {
	// Match verifies user identity information
	Match(ctx context.Context, req *KYCMatchRequest) (*KYCMatchResponse, error)
}
