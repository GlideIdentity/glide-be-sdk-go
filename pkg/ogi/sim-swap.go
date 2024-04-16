package ogi

import (
	"fmt"

	"github.com/opensaucerer/goaxios"
	log "github.com/sirupsen/logrus"
)

type LastSimChangeResponse struct {
	LatestSimChange string `json:"lastSimChange"`
}

type SimSwapResponse struct {
	Swapped bool `json:"swapped"`
}

func (c *GlideClient) RetrieveDate(phoneNumber string) (string, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return "", err
	}

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: Ciba,
        BaseAuthConfig: &BaseAuthConfig{
            Scopes: []string{
                "openid",
                "dpv:FraudPreventionAndDetection:sim-swap",
            },
            LoginHint: fmt.Sprintf("tel:%s", FormatPhoneNumber(phoneNumber)),
        },
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return "", err
	}

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/sim-swap/retrieve-date", envConfig.InternalApiBaseUrl),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]string{
			"phoneNumber": FormatPhoneNumber(phoneNumber),
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &LastSimChangeResponse{},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error retrieving date: %+v", res.Error)
		return "", res.Error
	}

    if res.Response.StatusCode != 200 {
        return "", fmt.Errorf("error retrieving date: %+v", res.Response.Status)
    }

	return res.Body.(*LastSimChangeResponse).LatestSimChange, nil
}

func (c *GlideClient) CheckSimSwap(phoneNumber string, maxAge int) (bool, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return false, err
	}

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: Ciba,
        BaseAuthConfig: &BaseAuthConfig{
            Scopes: []string{
                "openid",
                "dpv:FraudPreventionAndDetection:sim-swap",
            },
            LoginHint: fmt.Sprintf("tel:%s", FormatPhoneNumber(phoneNumber)),
        },
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return false, err
	}

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/sim-swap/check", envConfig.InternalApiBaseUrl),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"phoneNumber": FormatPhoneNumber(phoneNumber),
			"maxAge":     maxAge,
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &SimSwapResponse{},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error checking sim swap: %+v", res.Error)
		return false, res.Error
	}

    if res.Response.StatusCode != 200 {
        return false, fmt.Errorf("error retrieving date: %+v", res.Response.Status)
    }

	return res.Body.(*SimSwapResponse).Swapped, nil
}
