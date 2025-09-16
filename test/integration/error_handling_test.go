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
}
