package ogi

type Session struct {
	AccessToken string `json:"access_token"`
	SessionType SessionType
}

type SessionType int

const (
	Ciba SessionType = iota
	ThreeLeggedOAuth2
)

func (s *Session) GetScopes() []string {
	if s.AccessToken == "" {
		return []string{}
	}

	// TODO: implement the parse
	return []string{}
}
