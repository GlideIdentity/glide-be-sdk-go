package ogi

import (
	"fmt"

	"github.com/opensaucerer/goaxios"
	log "github.com/sirupsen/logrus"
)

type NumberVerificationResponse struct {
	DevicePhoneNumberVerified bool `json:"devicePhoneNumberVerified"`
}

type GetPhoneNumberResponse struct {
	DevicePhoneNumber string `json:"devicePhoneNumber"`
}

func (c *GlideClient) VerifyByNumber(phoneNumber string) (bool, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return false, err
	}

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: ThreeLeggedOAuth2,
        BaseAuthConfig: &BaseAuthConfig{
            Scopes: []string{
                "openid",
                "dpv:FraudDetectionAndPrevention:number-verification",
            },
            LoginHint: fmt.Sprintf("tel:%s", FormatPhoneNumber(phoneNumber)),
        },
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return false, err
	}

	if authRes.RedirectUrl != "" {
		log.Error("Doesn't have a ThreeLeggedOAuth2 session.")
		return false, fmt.Errorf("threeleggedoauth2 session is required to verify number - please call the authenticate method first")
	}

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/number-verification/verify", envConfig.InternalApiBaseUrl),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]string{
			"phoneNumber": phoneNumber,
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &NumberVerificationResponse{},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error verifying number: %+v", res.Error)
		return false, res.Error
	}

	return res.Body.(*NumberVerificationResponse).DevicePhoneNumberVerified, nil
}

func (c *GlideClient) VerifyByNumberHash(hasedPhoneNumber string) (bool, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return false, err
	}

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: ThreeLeggedOAuth2,
        BaseAuthConfig: &BaseAuthConfig{
            Scopes: []string{
                "openid",
                "dpv:FraudDetectionAndPrevention:number-verification",
            },
        },
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return false, err
	}

	if authRes.RedirectUrl != "" {
		log.Error("Doesn't have a ThreeLeggedOAuth2 session.")
		return false, fmt.Errorf("threeleggedoauth2 session is required to verify number - please call the authenticate method first")
	}

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/number-verification/verify", envConfig.InternalApiBaseUrl),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]string{
			"hasedPhoneNumber": hasedPhoneNumber,
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &NumberVerificationResponse{},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error verifying number: %+v", res.Error)
		return false, res.Error
	}

	return res.Body.(*NumberVerificationResponse).DevicePhoneNumberVerified, nil
}

func (c *GlideClient) GetPhoneNumber() (string, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return "", err
	}

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: ThreeLeggedOAuth2,
        BaseAuthConfig: &BaseAuthConfig{
            Scopes: []string{
                "openid",
                "dpv:FraudDetectionAndPrevention:number-verification",
            },
        },
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return "", err
	}

	if authRes.RedirectUrl != "" {
		log.Error("Doesn't have a ThreeLeggedOAuth2 session.")
		return "", fmt.Errorf("threeleggedoauth2 session is required to verify number - please call the authenticate method first")
	}

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/number-verification/device-phone-number", envConfig.InternalApiBaseUrl),
		Method: "GET",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &GetPhoneNumberResponse{},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error getting phone number: %+v", res.Error)
		return "", res.Error
	}

	return res.Body.(*GetPhoneNumberResponse).DevicePhoneNumber, nil
}
