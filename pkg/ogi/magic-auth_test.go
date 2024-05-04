package ogi_test

import (
	"os"
	"testing"

	"github.com/ClearBlockchain/glide-sdk-go/pkg/ogi"
	log "github.com/sirupsen/logrus"
)

var magicAuth *ogi.MagicAuth

func setupMagicAuth() {
	var err error

	magicAuth, err = ogi.NewMagicAuth()
	if err != nil {
		log.Fatalf("Error setting up client: %+v", err)
		panic(err)
	}

	testPhoneNumber = os.Getenv("GLIDE_TEST_PHONE_NUMBER")
}

func TestMagicAuth(t *testing.T) {
	setupMagicAuth()

	res, err := magicAuth.Authenticate(&ogi.StartVerificationDto{
		PhoneNumber: testPhoneNumber,
	})

	if err != nil {
		t.Fatalf("Error authenticating: %+v", err)
	}

	if res.Type != "MAGIC" {
		t.Fatalf("Expected type to be MAGIC, got %s", res.Type)
	}

	if !res.Verified {
		t.Fatalf("Expected verification to be successful")
	}
}

func TestMagicAuthFallback(t *testing.T) {
	setupMagicAuth()

	res, err := magicAuth.Authenticate(&ogi.StartVerificationDto{
		PhoneNumber: testPhoneNumber,
	})

	if err != nil {
		t.Fatalf("Error authenticating: %+v", err)
	}

	if res.Type != "SMS" {
		t.Fatalf("Expected type to be SMS, got %s", res.Type)
	}
}

func TestMagicAuthCheckCode(t *testing.T) {
	setupMagicAuth()

	res, err := magicAuth.CheckCode(&ogi.CheckCodeDto{
		PhoneNumber: testPhoneNumber,
		Code:        "981673",
	})

	if err != nil {
		t.Fatalf("Error checking code: %+v", err)
	}

	if !res {
		t.Fatalf("Expected code to be verified")
	}
}
