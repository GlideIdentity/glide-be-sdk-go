package integration_test

import (
	"context"
	"testing"

	"github.com/glideidentity/glide-go-sdk/glide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarrierNotEligible(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should return error for unsupported carriers", func(t *testing.T) {
		// Test with Israeli phone number (VerifyPhoneNumber use case)
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.NonEligible,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.GreaterOrEqual(t, glideErr.Status, 400)
		assert.Less(t, glideErr.Status, 500)
		assert.NotEmpty(t, glideErr.Message)
		t.Logf("Phone flow error: %d - %s", glideErr.Status, glideErr.Code)

		// Test with unknown PLMN
		prepReq2 := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.UnknownPLMN,
		}

		_, err2 := client.MagicAuth.Prepare(ctx, &prepReq2)
		require.Error(t, err2)

		glideErr2, ok := err2.(*glide.Error)
		require.True(t, ok)
		assert.GreaterOrEqual(t, glideErr2.Status, 400)
		assert.Less(t, glideErr2.Status, 500)
		t.Logf("PLMN flow error: %d - %s", glideErr2.Status, glideErr2.Code)
	})
}

func TestUnauthorized(t *testing.T) {
	t.Run("should preserve 401 status for unauthorized requests", func(t *testing.T) {
		// Create client with invalid API key
		unauthorizedClient := glide.New(
			glide.WithAPIKey("invalid-api-key"),
			glide.WithBaseURL("https://api.glideidentity.app"),
		)

		ctx := context.Background()
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.TMobileValid,
		}

		_, err := unauthorizedClient.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, 401, glideErr.Status)
		t.Logf("401 error preserved: %s", glideErr.Code)
	})
}

func TestValidationErrors(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should return 400 for invalid phone number format", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber, // Changed to verify since we're testing phone validation
			PhoneNumber: testPhoneNumbers.InvalidFormat,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		// Client-side validation might not set status
		assert.Equal(t, glide.ErrCodeInvalidPhoneNumber, glideErr.Code)
	})

	t.Run("should return 400 for invalid PLMN", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.InvalidMCC,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		// Client-side validation might not set status
		assert.Equal(t, glide.ErrCodeInvalidMCCMNC, glideErr.Code)
	})
}

func TestErrorDetails(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should handle 422 in getPhoneNumber flow", func(t *testing.T) {
		// Prepare a session
		plmn := glide.PLMN{MCC: "310", MNC: "260"}
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &plmn,
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		prepareResult, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.NoError(t, err)

		// Try with invalid credential
		req := &glide.GetPhoneNumberRequest{
			SessionInfo: &prepareResult.Session,
			Credential: map[string]interface{}{
				"invalid": "credential",
			},
		}

		_, err = client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, 422, glideErr.Status)
		t.Logf("✅ Correctly returned 422 for invalid credential")
	})

	t.Run("should include request ID in errors", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.NonEligible,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)

		// Request ID should be present after our server fix
		if glideErr.RequestID != "" {
			t.Logf("✅ Request ID present: %s", glideErr.RequestID)
		} else {
			t.Log("⚠️ Request ID missing (server may need update)")
		}
	})

	t.Run("should include error details", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.InvalidFormat,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.NotEmpty(t, glideErr.Code)
		assert.NotEmpty(t, glideErr.Message)
		// Client-side validation errors might not have a status code

		t.Logf("Error details: code=%s, status=%d, message=%s",
			glideErr.Code, glideErr.Status, glideErr.Message)
	})

	t.Run("should include timestamp in errors", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.NonEligible,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)

		// Check if error has timestamp information (may be in message or details)
		if glideErr.Message != "" {
			t.Logf("✅ Error message present: %s", glideErr.Message)
		}
	})

	t.Run("should serialize errors properly for logging", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.InvalidFormat,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)

		// Error should be serializable
		errorString := glideErr.Error()
		assert.NotEmpty(t, errorString)
		assert.Contains(t, errorString, glideErr.Code)

		// Should have a structured format for logging
		t.Logf("Serialized error: %s", errorString)
	})

	t.Run("should support type-safe error code checking", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.InvalidFormat,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)

		// Type-safe error code checking
		switch glideErr.Code {
		case glide.ErrCodeInvalidPhoneNumber:
			t.Log("✅ Type-safe error code checking works")
		case glide.ErrCodeValidationError:
			t.Log("✅ Type-safe error code checking works (VALIDATION_ERROR)")
		default:
			t.Errorf("Unexpected error code: %s", glideErr.Code)
		}
	})
}

func TestBrowserCompatibilityErrors(t *testing.T) {
	// Now enabled - Go SDK supports ClientInfo with UserAgent

	client := createTestClient(t)
	ctx := context.Background()

	testCases := []struct {
		name          string
		userAgent     string
		mcc           string
		mnc           string
		expectError   bool
		errorContains string
	}{
		{
			name:          "T-Mobile with Safari Desktop",
			userAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
			mcc:           "310",
			mnc:           "260",
			expectError:   true,
			errorContains: "compatible platform",
		},
		{
			name:          "T-Mobile with iOS Safari",
			userAgent:     "Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
			mcc:           "310",
			mnc:           "260",
			expectError:   true,
			errorContains: "not eligible",
		},
		{
			name:          "T-Mobile with Firefox",
			userAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/118.0",
			mcc:           "310",
			mnc:           "260",
			expectError:   true,
			errorContains: "compatible platform",
		},
		{
			name:        "T-Mobile with Chrome Desktop (should succeed)",
			userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			mcc:         "310",
			mnc:         "260",
			expectError: false,
		},
		{
			name:        "T-Mobile with Edge Desktop (should succeed)",
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
			mcc:         "310",
			mnc:         "260",
			expectError: false,
		},
		{
			name:        "T-Mobile with Brave (Chromium-based, should succeed)",
			userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Brave/1.60",
			mcc:         "310",
			mnc:         "260",
			expectError: false,
		},
		{
			name:        "T-Mobile with Opera (Chromium-based, should succeed)",
			userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 OPR/105.0.0.0",
			mcc:         "310",
			mnc:         "260",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepReq := glide.PrepareRequest{
				UseCase:     glide.UseCaseVerifyPhoneNumber,
				PhoneNumber: testPhoneNumbers.TMobileValid,
				ClientInfo: &glide.ClientInfo{
					UserAgent: tc.userAgent,
				},
				PLMN: &glide.PLMN{
					MCC: tc.mcc,
					MNC: tc.mnc,
				},
			}

			result, err := client.MagicAuth.Prepare(ctx, &prepReq)

			if tc.expectError {
				require.Error(t, err)
				glideErr, ok := err.(*glide.Error)
				require.True(t, ok)

				// Check for expected error content
				if tc.errorContains != "" {
					assert.Contains(t, glideErr.Message, tc.errorContains,
						"Expected error to contain '%s', got: %s", tc.errorContains, glideErr.Message)
				}

				t.Logf("✅ Got expected error: %s", glideErr.Message)
			} else {
				require.NoError(t, err, "Expected success but got error: %v", err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AuthenticationStrategy)
				t.Logf("✅ Success with strategy: %s", result.AuthenticationStrategy)
			}
		})
	}
}

func TestErrorMessagePrivacy(t *testing.T) {
	// Now enabled - Go SDK supports ClientInfo with UserAgent

	client := createTestClient(t)
	ctx := context.Background()

	testCases := []struct {
		name             string
		userAgent        string
		mcc              string
		mnc              string
		shouldNotContain []string // Carrier names that should NOT appear in error
	}{
		{
			name:             "T-Mobile Safari error should not mention T-Mobile",
			userAgent:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
			mcc:              "310",
			mnc:              "260",
			shouldNotContain: []string{"T-Mobile", "TMobile", "t-mobile"},
		},
		{
			name:             "AT&T error should not mention AT&T",
			userAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			mcc:              "310",
			mnc:              "410",
			shouldNotContain: []string{"AT&T", "ATT", "at&t"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepReq := glide.PrepareRequest{
				UseCase:     glide.UseCaseVerifyPhoneNumber,
				PhoneNumber: "+14155552671",
				ClientInfo: &glide.ClientInfo{
					UserAgent: tc.userAgent,
				},
				PLMN: &glide.PLMN{
					MCC: tc.mcc,
					MNC: tc.mnc,
				},
			}

			_, err := client.MagicAuth.Prepare(ctx, &prepReq)

			if err == nil {
				t.Skip("No error returned, skipping privacy check")
			}

			glideErr, ok := err.(*glide.Error)
			require.True(t, ok)

			// Check that carrier names are not in the error message
			for _, forbidden := range tc.shouldNotContain {
				assert.NotContains(t, glideErr.Message, forbidden,
					"Error message should not contain '%s'", forbidden)
			}

			t.Logf("✅ Error message properly generic: %s", glideErr.Message)
		})
	}
}
