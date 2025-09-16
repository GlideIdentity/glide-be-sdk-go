package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/glideidentity/glide-go-sdk/glide"
)

func TestDebugPrepareRequest(t *testing.T) {
	// Enable debug logging
	logger := glide.NewDefaultLogger(glide.LogLevelDebug)

	apiKey := os.Getenv("GLIDE_API_KEY")
	if apiKey == "" {
		apiKey = "0NrxSqb7oWXzZXo7Cq25hgvwvwN60lP2MYERLlFrVyaKPiJB"
	}

	client := glide.New(
		glide.WithAPIKey(apiKey),
		glide.WithBaseURL("https://api.glideidentity.app"),
		glide.WithLogger(logger),
	)

	ctx := context.Background()

	t.Log("Testing with debug output...")

	// Test 1: GetPhoneNumber with PLMN (correct)
	t.Run("get_phone_with_plmn", func(t *testing.T) {
		plmn := glide.PLMN{MCC: "310", MNC: "260"}
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &plmn,
			// No phone number - we're trying to get it
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		t.Logf("Request: %+v", prepReq)
		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		if err != nil {
			t.Logf("Error: %v", err)
		} else {
			t.Logf("Success: Strategy=%s, SessionKey=%s", result.AuthenticationStrategy, result.Session.SessionKey)
		}
	})

	// Test 2: VerifyPhoneNumber with phone (correct)
	t.Run("verify_with_phone", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: "+14157400083", // T-Mobile number that TelcoFinder recognizes
			// No PLMN needed - server uses TelcoFinder
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		t.Logf("Request: %+v", prepReq)
		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		if err != nil {
			t.Logf("Error: %v", err)
		} else {
			t.Logf("Success: Strategy=%s, SessionKey=%s", result.AuthenticationStrategy, result.Session.SessionKey)
		}
	})

	// Test 3: Mixed parameters (should fail with right error)
	t.Run("mixed_params_should_fail", func(t *testing.T) {
		plmn := glide.PLMN{MCC: "310", MNC: "260"}
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseGetPhoneNumber,
			PhoneNumber: "+13105551234", // Wrong - shouldn't provide this
			PLMN:        &plmn,
		}

		t.Logf("Request: %+v", prepReq)
		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		if err != nil {
			t.Logf("Expected error: %v", err)
		} else {
			t.Logf("Unexpected success: Strategy=%s, SessionKey=%s", result.AuthenticationStrategy, result.Session.SessionKey)
		}
	})
}
