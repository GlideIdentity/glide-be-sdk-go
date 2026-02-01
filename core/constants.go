package core

// UseCase represents the authentication use case.
type UseCase string

const (
	UseCaseGetPhoneNumber    UseCase = "GetPhoneNumber"
	UseCaseVerifyPhoneNumber UseCase = "VerifyPhoneNumber"
)

// AuthenticationStrategy represents the authentication method.
type AuthenticationStrategy string

const (
	AuthenticationStrategyTS43    AuthenticationStrategy = "ts43"
	AuthenticationStrategyLink    AuthenticationStrategy = "link"
	AuthenticationStrategyDesktop AuthenticationStrategy = "desktop"
)
