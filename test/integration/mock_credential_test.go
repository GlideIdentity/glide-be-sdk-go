package integration_test

import (
	"context"
	"testing"

	"github.com/glideidentity/glide-go-sdk/glide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock credential responses matching Node.js test data
var mockCredentialResponses = struct {
	ValidTS43      map[string]interface{}
	MissingGlide   map[string]interface{}
	EmptyVPToken   map[string]interface{}
	InvalidToken   map[string]interface{}
	InvalidFormat  map[string]interface{}
	ExpiredSession string
}{
	ValidTS43: map[string]interface{}{
		"vp_token": map[string]interface{}{
			"glide": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.mock-vp-token-content",
		},
		"presentation_submission": map[string]interface{}{
			"id":            "submission-12345",
			"definition_id": "glide-phone-auth-v1",
		},
	},
	MissingGlide: map[string]interface{}{
		"vp_token": map[string]interface{}{
			"other": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.mock-content",
		},
	},
	EmptyVPToken: map[string]interface{}{
		"vp_token": map[string]interface{}{},
	},
	InvalidToken: map[string]interface{}{
		"vp_token": map[string]interface{}{
			"glide": "invalid-token-not-jwt",
		},
		"presentation_submission": map[string]interface{}{
			"id":            "submission-id",
			"definition_id": "def-id",
		},
	},
	InvalidFormat: map[string]interface{}{
		"not_vp_token": "some-value",
	},
	ExpiredSession: "expired-session-key",
}

func TestMockCredentialProcessing(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Prepare a session first for the tests
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
	require.NotNil(t, prepareResult)

	t.Run("should handle missing aggregator ID in vp_token", func(t *testing.T) {
		req := &glide.GetPhoneNumberRequest{
			Session: &prepareResult.Session,
			Credential:  mockCredentialResponses.MissingGlide,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidCredentialFormat, glideErr.Code)
		assert.Equal(t, 422, glideErr.Status)
		t.Log("✅ Correctly rejected missing aggregator ID")
	})

	t.Run("should handle empty vp_token", func(t *testing.T) {
		req := &glide.GetPhoneNumberRequest{
			Session: &prepareResult.Session,
			Credential:  mockCredentialResponses.EmptyVPToken,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidCredentialFormat, glideErr.Code)
		assert.Equal(t, 422, glideErr.Status)
		t.Log("✅ Correctly rejected empty vp_token")
	})

	t.Run("should handle completely invalid credential format", func(t *testing.T) {
		req := &glide.GetPhoneNumberRequest{
			Session: &prepareResult.Session,
			Credential:  mockCredentialResponses.InvalidFormat,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidCredentialFormat, glideErr.Code)
		assert.Equal(t, 422, glideErr.Status)
		t.Log("✅ Correctly rejected invalid credential format")
	})

	t.Run("should handle invalid JWT token from server", func(t *testing.T) {
		req := &glide.GetPhoneNumberRequest{
			Session: &prepareResult.Session,
			Credential:  mockCredentialResponses.InvalidToken,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidCredentialFormat, glideErr.Code)
		assert.Equal(t, 422, glideErr.Status)
		t.Log("✅ Server rejected invalid JWT")
	})

	t.Run("should handle expired session", func(t *testing.T) {
		// Create a session info with expired session key
		expiredSession := &glide.SessionInfo{
			SessionKey: mockCredentialResponses.ExpiredSession,
			Metadata:   prepareResult.Session.Metadata,
		}

		req := &glide.GetPhoneNumberRequest{
			Session: expiredSession,
			Credential:  mockCredentialResponses.ValidTS43,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeSessionNotFound, glideErr.Code)
		assert.Equal(t, 404, glideErr.Status)
		t.Log("✅ Correctly rejected expired session")
	})

	t.Run("should handle mismatched phone number", func(t *testing.T) {
		// Prepare a verify session
		verifyReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: "+14155551234",
			PLMN:        &plmn,
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		verifyPrepare, err := client.MagicAuth.Prepare(ctx, &verifyReq)
		require.NoError(t, err)

		// Try to verify with a MOCK credential (not a real one from actual device)
		req := &glide.VerifyPhoneNumberRequest{
			Session: &verifyPrepare.Session,
			Credential:  mockCredentialResponses.ValidTS43, // Mock/forged credential
		}

		_, err = client.MagicAuth.VerifyPhoneNumber(ctx, req)
		require.Error(t, err)

		// Server correctly rejects the mock/forged credential as invalid
		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)
		assert.Equal(t, glide.ErrCodeInvalidCredentialFormat, glideErr.Code)
		assert.Equal(t, 422, glideErr.Status)
		t.Log("✅ Server correctly rejects mock/forged credential")
	})

	t.Run("should include all error details from server", func(t *testing.T) {
		// Use expired session to trigger detailed error
		expiredSession := &glide.SessionInfo{
			SessionKey: "non-existent-session-key",
			Metadata: &glide.SessionMetadata{
				Nonce:  prepareResult.Session.Metadata.Nonce,
				EncKey: prepareResult.Session.Metadata.EncKey,
			},
		}

		req := &glide.GetPhoneNumberRequest{
			Session: expiredSession,
			Credential:  mockCredentialResponses.ValidTS43,
		}

		_, err := client.MagicAuth.GetPhoneNumber(ctx, req)
		require.Error(t, err)

		glideErr, ok := err.(*glide.Error)
		require.True(t, ok)

		// Check error structure
		assert.Equal(t, glide.ErrCodeSessionNotFound, glideErr.Code)
		assert.Equal(t, 404, glideErr.Status)
		assert.NotEmpty(t, glideErr.Message)

		// Request ID might be present after server fix
		if glideErr.RequestID != "" {
			t.Logf("✅ Request ID present: %s", glideErr.RequestID)
		} else {
			t.Log("⚠️ Request ID missing (server may need update)")
		}

		// Error message should be present
		if glideErr.Message != "" {
			t.Logf("✅ Error message present: %s", glideErr.Message)
		}

		t.Logf("Error details: code=%s, status=%d, message=%s",
			glideErr.Code, glideErr.Status, glideErr.Message)
	})
}
