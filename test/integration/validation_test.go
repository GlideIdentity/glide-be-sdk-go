package integration_test

import (
	"context"
	"testing"

	"github.com/glideidentity/glide-go-sdk/glide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhoneNumberValidation(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	testCases := []struct {
		name        string
		phoneNumber string
		expectError bool
		errorCode   string
	}{
		{
			name:        "reject phone without + prefix",
			phoneNumber: testPhoneNumbers.MissingPlus,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
		{
			name:        "reject phone with spaces",
			phoneNumber: testPhoneNumbers.WithSpaces,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
		{
			name:        "reject phone with dashes",
			phoneNumber: testPhoneNumbers.WithDashes,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
		{
			name:        "reject too short phone",
			phoneNumber: testPhoneNumbers.ShortNumber,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
		{
			name:        "reject too long phone",
			phoneNumber: testPhoneNumbers.LongNumber,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
		{
			name:        "reject non-numeric characters",
			phoneNumber: testPhoneNumbers.NonNumeric,
			expectError: true,
			errorCode:   glide.ErrCodeInvalidPhoneNumber,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepReq := glide.PrepareRequest{
				UseCase:     glide.UseCaseVerifyPhoneNumber, // Changed to VerifyPhoneNumber since we're testing phone numbers
				PhoneNumber: tc.phoneNumber,
			}

			_, err := client.MagicAuth.Prepare(ctx, &prepReq)

			if tc.expectError {
				require.Error(t, err)
				glideErr, ok := err.(*glide.Error)
				require.True(t, ok)
				assert.Equal(t, tc.errorCode, glideErr.Code)
				t.Logf("✅ Correctly rejected: %s", tc.name)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPLMNValidation(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should reject invalid MCC format", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.InvalidMCC,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidMCCMNC, glideErr.Code)
		assert.Contains(t, glideErr.Message, "MCC")
		t.Log("✅ Correctly rejected invalid MCC")
	})

	t.Run("should reject invalid MNC format", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.InvalidMNC,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidMCCMNC, glideErr.Code)
		assert.Contains(t, glideErr.Message, "MNC")
		t.Log("✅ Correctly rejected invalid MNC")
	})

	t.Run("should accept unofficial MCC for telco labs", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TestLab,
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)

		// Client should accept unofficial MCC
		// Server might reject if not in carrier config
		if err != nil {
			glideErr, ok := err.(*glide.Error)
			require.True(t, ok)
			assert.GreaterOrEqual(t, glideErr.Status, 400)
			assert.Less(t, glideErr.Status, 500)
			t.Logf("✅ Client accepted MCC, server rejected as expected: %s", glideErr.Code)
		} else {
			t.Log("✅ Unofficial MCC found in carrier config")
		}
	})
}

func TestUseCaseValidation(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should require phone number for VerifyPhoneNumber", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseVerifyPhoneNumber,
			// Missing phone_number - should fail
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeMissingParameters, glideErr.Code)
		t.Log("✅ Correctly requires phone for VerifyPhoneNumber")
	})

	t.Run("should require PLMN for GetPhoneNumber", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			// No PLMN - should fail
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeMissingParameters, glideErr.Code)
		assert.Contains(t, glideErr.Message, "PLMN")
		t.Log("✅ Correctly requires PLMN for GetPhoneNumber")
	})

	t.Run("should allow GetPhoneNumber without phone number if PLMN provided", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TMobileUS,
			// No phone number - correct for GetPhoneNumber
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AuthenticationStrategy)
		t.Log("✅ GetPhoneNumber works correctly with PLMN only")
	})
}

func TestConsentDataValidation(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should reject consent data with missing required fields", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TMobileUS,
			ConsentData: &glide.ConsentData{
				ConsentText: "I agree", // Missing policy fields
			},
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeValidationError, glideErr.Code)
		t.Log("✅ Correctly rejected incomplete consent data")
	})

	t.Run("should reject consent data with invalid URL format", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TMobileUS,
			ConsentData: &glide.ConsentData{
				ConsentText: "I agree to the terms",
				PolicyLink:  "not-a-url", // Invalid URL
				PolicyText:  "Privacy policy text",
			},
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeValidationError, glideErr.Code)
		t.Log("✅ Correctly rejected invalid URL in consent")
	})
}
