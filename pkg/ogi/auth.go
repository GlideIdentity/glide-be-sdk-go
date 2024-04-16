package ogi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	b64 "encoding/base64"

	"github.com/ClearBlockchain/glide-sdk-go/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type BaseAuthConfig struct {
	Scopes []string
	LoginHint string
}

type AuthConfig struct {
	*BaseAuthConfig

	// ciba, oauth2
	Provider SessionType
}

type AuthenticationResponse struct {
	Session *Session
	RedirectUrl string
}

type cibaAuthResponse struct {
	AuthRequestId string `json:"auth_req_id"`
	ExpiresIn int `json:"expires_in"`
	Interval int `json:"interval"`
}

func (c *GlideClient) getBasicAuthHeader() string {
	return fmt.Sprintf(
		"Basic %s",
		b64.StdEncoding.EncodeToString(
			[]byte(
				fmt.Sprintf(
					"%s:%s",
					c.clientId,
					c.clientSecret,
				),
			),
		),
	)
}

func (c *GlideClient) getCibaAuthLoginHint(authConfig *BaseAuthConfig) (authReqId string, expiresIn int, interval int, err error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return "", 0, 0, err
	}

	log.Debugf("Getting ciba auth login hint with config: %+v", authConfig)

	// Prepare the request parameters
	data := url.Values{}
	data.Set("scope", strings.Join(authConfig.Scopes, " "))
	data.Set("login_hint", authConfig.LoginHint)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/oauth2/backchannel-authentication", envConfig.InternalAuthBaseUrl),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		log.Errorf("Error creating ciba auth login hint request: %+v", err)
		return "", 0, 0, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.getBasicAuthHeader())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("Error getting ciba auth login: %+v", err)
		return "", 0, 0, err
	}

	if res.StatusCode != 200 {
		log.Errorf("Error getting ciba auth login: %+v", res.Body)
		return "", 0, 0, fmt.Errorf("error getting ciba auth login %+v", res.Body)
	}

	log.Debugf("raw get ciba login hint response: %+v", res.Body)
	var resData cibaAuthResponse
	if err := utils.GetJsonBody(res, &resData); err != nil {
		log.Errorf("Error parsing ciba auth login response: %+v", err)
		return "", 0, 0, err
	}

	log.Debugf("Ciba auth login hint response: %+v", resData)
	return resData.AuthRequestId, resData.ExpiresIn, resData.Interval, nil
}

func (c *GlideClient) fetchCibaToken(authReqId string) (*Session, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("grant_type", "urn:openid:params:grant-type:ciba")
	data.Set("auth_req_id", authReqId)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/oauth2/token", envConfig.InternalAuthBaseUrl),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		log.Errorf("Error creating ciba token request: %+v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.getBasicAuthHeader())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("Error fetching ciba token: %+v", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		log.Errorf("Error fetching ciba token: %+v", res.Body)
		return nil, fmt.Errorf("error fetching ciba token %+v", res.Body)
	}

	log.Debugf("raw ciba token response: %+v", res.Body)
	session := &Session{}
	if err := utils.GetJsonBody(res, session); err != nil {
		log.Errorf("Error parsing ciba token response: %+v", err)
		return nil, err
	}

	log.Debug("Ciba token fetched successfully")
	return session, nil
}

func (c *GlideClient) pollCibaToken(authReqId string, interval int) (*Session, error) {
	if interval < 1 {
		return nil, fmt.Errorf("invalid interval: %d", interval)
	}

	log.Debugf("Polling ciba token with auth req id: %s", authReqId)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
			case <-ticker.C:
				log.Debugf("Polling ciba token with auth req id: %s", authReqId)
				// make request to ciba token endpoint
				session, err := c.fetchCibaToken(authReqId)
				if err != nil {
					log.Errorf("Error fetching ciba token: %+v", err)
					return nil, err
				}

				if session.AccessToken != "" {
					log.Debug("Ciba token polling completed successfully with session")
					return session, nil
				}

				log.Debugf("Couldn't get ciba access token. Trying again in %d seconds", interval)
			case <-time.After(2 * time.Minute):
				log.Errorf("Ciba token polling timeout")
				return nil, fmt.Errorf("ciba token polling timeout")
		}
	}
}

func (c *GlideClient) getCibaSession(authConfig *BaseAuthConfig) (*Session, error) {
	log.Debug("Starting ciba authentication flow")
	authReqId, _, interval, err := c.getCibaAuthLoginHint(authConfig)
	if err != nil {
		log.Errorf("Error getting ciba auth login hint: %+v", err)
		return nil, err
	}

	session, err := c.pollCibaToken(authReqId, interval)
	if err != nil {
		log.Errorf("Error polling ciba token: %+v", err)
		return nil, err
	}

	session.SessionType = Ciba
	log.Debug("Ciba authentication flow completed successfully with session")
	return session, nil
}

func (c *GlideClient) get3LeggedAuthRedirectUrl(authConfig *BaseAuthConfig) (redirectUrl string, err error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return "", err
	}

	log.Debug("Generating 3-legged auth redirect url")
	nonce := randomString(16)
	state := randomString(10)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/oauth2/auth", envConfig.InternalAuthBaseUrl), nil)
	if err != nil {
		log.Errorf("Error creating 3-legged auth request: %+v", err)
		return "", err
	}

	q := req.URL.Query()
	q.Add("client_id", c.clientId)
	q.Add("redirect_uri", c.redirectUri)
	q.Add("state", state)
	q.Add("response_type", "code")
	q.Add("scope", strings.Join(authConfig.Scopes, " "))
	q.Add("nonce", nonce)
	q.Add("max_age", "0")
	q.Add("purpose", "") // ????
	q.Add("audience", c.clientId)
    if authConfig.LoginHint != "" {
        q.Add("login_hint", authConfig.LoginHint)
    }

	req.URL.RawQuery = q.Encode()
	url := req.URL.String()
	log.Debugf("3-legged auth url: %s", url)
	return url, nil
}

func (c *GlideClient) ExchangeCodeForSession(code string) (*Session, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", c.redirectUri)
	data.Set("code", code)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/oauth2/token", envConfig.InternalAuthBaseUrl),
		strings.NewReader(data.Encode()),
	)

	if err != nil {
		log.Errorf("Error creating code exchange request: %+v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.getBasicAuthHeader())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("Error exchanging code for session: %+v", err)
		return nil, err
	}

	if res.StatusCode != 200 {
		log.Errorf("Error exchanging code for session: %+v", res.Body)
		return nil, fmt.Errorf("error exchanging code for session %+v", res.Body)
	}

	log.Debugf("raw code exchange response: %+v", res.Body)
	session := &Session{}
	if err := utils.GetJsonBody(res, session); err != nil {
		log.Errorf("Error parsing code exchange response: %+v", err)
		return nil, err
	}

	log.Debug("Code exchange completed successfully with session")
    c.session = session
	return session, nil
}

func (c *GlideClient) Authenticate(authConfig *AuthConfig) (response *AuthenticationResponse, err error) {
	// only run auth flow if session type is higher
	if c.session != nil && c.session.SessionType >= authConfig.Provider {
		log.Debugf("Current session type is higher than requested provider. Skipping auth flow.")
		return &AuthenticationResponse{Session: c.session}, nil
	}

	// if no base auth config is provided, use the default one
	if authConfig.BaseAuthConfig == nil {
		authConfig.BaseAuthConfig = &BaseAuthConfig{
			Scopes: []string{"openid"},
		}
	}

	switch authConfig.Provider {
	case Ciba:
		log.Debug("Starting ciba authentication flow")
		session, err := c.getCibaSession(authConfig.BaseAuthConfig)
		if err != nil {
			log.Errorf("Error getting ciba session: %+v", err)
			return nil, err
		}

        c.session = session
		return &AuthenticationResponse{Session: session}, nil

	case ThreeLeggedOAuth2:
		log.Debug("Starting 3-legged oauth2 authentication flow")
		redirectUrl, err := c.get3LeggedAuthRedirectUrl(authConfig.BaseAuthConfig)
		if err != nil {
			log.Errorf("Error getting 3-legged auth redirect url: %+v", err)
			return nil, err
		}

		return &AuthenticationResponse{RedirectUrl: redirectUrl}, nil
	default:
		return nil, fmt.Errorf("invalid provider: %d. Can only be '%d' or '%d'", authConfig.Provider, Ciba, ThreeLeggedOAuth2)
	}
}
