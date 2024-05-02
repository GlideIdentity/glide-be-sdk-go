package ogi_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ClearBlockchain/glide-sdk-go/pkg/ogi"
	log "github.com/sirupsen/logrus"
)

var client *ogi.GlideClient
var testPhoneNumber string
var testPhoneNumber2 string

func setupClient() {
	var err error

	client, err = ogi.NewGlideClient()
	if err != nil {
		log.Fatalf("Error setting up client: %+v", err)
		panic(err)
	}

    testPhoneNumber = os.Getenv("GLIDE_TEST_PHONE_NUMBER")
    testPhoneNumber2 = os.Getenv("GLIDE_TEST_PHONE_NUMBER_2")
}

func TestAuthenticateWithCiba(t *testing.T) {
	setupClient()

	res, err := client.Authenticate(&ogi.AuthConfig{
		Provider: ogi.Ciba,
		BaseAuthConfig: &ogi.BaseAuthConfig{
			Scopes: []string{
				"openid",
				"dpv:FraudPreventionAndDetection:sim-swap",
			},
			LoginHint: fmt.Sprintf("tel:%s", testPhoneNumber2),
		},
	})

	if err != nil {
		t.Fatalf("Error authenticating: %+v", err)
	}

	if res.Session == nil {
		t.Fatalf("Session is nil")
	}
}

func TestAuthenticateWithOAuth2(t *testing.T) {
	setupClient()

	res, err := client.Authenticate(&ogi.AuthConfig{
		Provider: ogi.ThreeLeggedOAuth2,
		BaseAuthConfig: &ogi.BaseAuthConfig{
			Scopes: []string{
				"openid",
				"dpv:FraudPreventionAndDetection:sim-swap",
			},
			LoginHint: fmt.Sprintf("tel:%s", testPhoneNumber),
		},
	})

	if err != nil {
		t.Fatalf("Error authenticating: %+v", err)
	}

	if res.RedirectUrl == "" {
		t.Fatalf("RedirectUrl is missing")
	}
}

// FIXME: unable to test - testPhoneNumber requires a confirmation
// phoneNumber2 does not implement this endpoint
func TestRetrieveDate(t *testing.T) {
	setupClient()

	lastSimChanged, err := client.RetrieveDate(testPhoneNumber2)
	if err != nil {
		t.Fatalf("Error getting phone number: %+v", err)
	}

	if lastSimChanged == "" {
		t.Fatalf("Last sim changed is empty")
	}
}

func TestCheckSimSwap(t *testing.T) {
	setupClient()

	valid, err := client.CheckSimSwap(testPhoneNumber2, 100)
	if err != nil {
		t.Fatalf("Error getting phone number: %+v", err)
	}

	if !valid {
		t.Fatalf("Sim swap is not valid")
	}
}

func TestVerifyLocation(t *testing.T) {
	setupClient()

	location := ogi.LocationBody{
		DeviceID:     testPhoneNumber2,
		DeviceIDType: ogi.PHONE_NUMBER,
		Latitude:     40.416775,
		Longitude:    -3.703790,
		Radius:       1000,
		MaxAge:       3600,
	}

	valid, err := client.VerifyLocation(location)
	if err != nil {
		t.Fatalf("Error verifying location: %+v", err)
	}

	if !valid {
		t.Fatalf("Location is not valid")
	}
}

func TestMagicAuth(t *testing.T) {
	setupClient()
  
	startVerificationDto := &ogi.StartVerificationDto{
	  PhoneNumber: testPhoneNumber,
	  Email: "",
	  FallbackChannel: "SMS",
	}
  
	res, err := client.MagicAuth(startVerificationDto)
  
	if err != nil {
	  t.Fatalf("Error starting verification: %+v", err)
	}
  
	if res == nil {
	  t.Fatalf("Response is nil")
	}
  
	if res.Type != ogi.MAGIC {
	  t.Fatalf("Verification type is incorrect")
	}
  }
  
  func TestVerifyToken(t *testing.T) {
	setupClient()
  
	checkCodeDto := &ogi.CheckCodeDto{
	  PhoneNumber: testPhoneNumber,
	  Email: "",
	  Code: "123456",
	}
  
	valid, err := client.VerifyToken(checkCodeDto)
  
	if err != nil {
	  t.Fatalf("Error verifying token: %+v", err)
	}
  
	if !valid {
	  t.Fatalf("Token is not valid")
	}
  }
  