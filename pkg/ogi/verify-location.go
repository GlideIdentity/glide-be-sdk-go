package ogi

import (
	"fmt"

	"github.com/opensaucerer/goaxios"
	log "github.com/sirupsen/logrus"
)

type LocationBody struct {
	Latitude     float64       `json:"latitude"`
	Longitude    float64       `json:"longitude"`
	Radius       int           `json:"radius"`
	DeviceID     string        `json:"deviceId"`
	DeviceIDType DeviceIDType  `json:"deviceIdType"`
	// Accuracy     int           `json:"accuracy"` // TODO: - not used in the original sdk
	MaxAge       int           `json:"maxAge"`
}

type LocationResponse struct {
	VerificationResult string `json:"verificationResult"` //TODO: - why is this string and not boolean?
}

type DeviceIDType string

const (
	IPV4 DeviceIDType = "ipv4Address"
	IPV6 DeviceIDType = "ipv6Address"
	PHONE_NUMBER DeviceIDType = "phoneNumber"
	NAI DeviceIDType = "networkAccessIdentifier"
)

func (c *GlideClient) VerifyLocation(location LocationBody) (bool, error) {
	envConfig, err := ReadEnv()
	if err != nil {
		return false, err
	}

    baseAuthConfig := &BaseAuthConfig{
        Scopes: []string{
            "openid",
            "dpv:FraudPreventionAndDetection:device-location",
        },
    }

	authRes, err := c.Authenticate(&AuthConfig{
		Provider: Ciba,
        BaseAuthConfig: baseAuthConfig,
	})

	if err != nil {
		log.Errorf("Error authenticating: %+v", err)
		return false, err
	}

	// set default device id type
	if location.DeviceIDType == "" {
		location.DeviceIDType = PHONE_NUMBER
	}

	// set default radius to 2000
	if location.Radius == 0 {
		location.Radius = 2000
	}

	// set default max age to 3600
	if location.MaxAge == 0 {
		location.MaxAge = 3600
	}

    if location.DeviceIDType == PHONE_NUMBER {
        baseAuthConfig.LoginHint = fmt.Sprintf("tel:%s", FormatPhoneNumber(location.DeviceID))
    } else if location.DeviceIDType == IPV4 || location.DeviceIDType == IPV6 {
        baseAuthConfig.LoginHint = fmt.Sprintf("ipport:%s", location.DeviceID)
    }

	req := goaxios.GoAxios{
		Url: fmt.Sprintf("%s/device-location/verify", envConfig.InternalApiBaseUrl),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		BearerToken: authRes.Session.AccessToken,
		ResponseStruct: &LocationResponse{},
		Body: map[string]interface{}{
			"device": map[string]string{
				string(location.DeviceIDType): location.DeviceID,
			},
			"area": map[string]interface{}{
				"areaType": "CIRCLE",
				"center": map[string]float64{
					"latitude": location.Latitude,
					"longitude": location.Longitude,
				},
				"radius": location.Radius,
			},
			"maxAge": location.MaxAge,
		},
	}

	res := req.RunRest()
	if res.Error != nil {
		log.Errorf("Error verifying location: %+v", res.Error)
		return false, res.Error
	}

	locationRes, ok := res.Body.(*LocationResponse)
	if !ok {
		log.Errorf("Error parsing location response: %+v", res.Error)
		return false, fmt.Errorf("error parsing location response")
	}

	if locationRes.VerificationResult == "FALSE" {
		return false, fmt.Errorf("location verification failed")
	}

	return true, nil
}
