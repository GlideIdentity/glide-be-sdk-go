package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glideidentity/glide-go-sdk/glide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data
var (
	testPhoneNumbers = struct {
		TMobileValid  string
		NonEligible   string
		InvalidFormat string
		ShortNumber   string
		LongNumber    string
		NonNumeric    string
		MissingPlus   string
		WithSpaces    string
		WithDashes    string
	}{
		TMobileValid:  "+14157400083",  // T-Mobile number verified with TelcoFinder
		NonEligible:   "+972549982913", // Israeli number
		InvalidFormat: "1234567890",
		ShortNumber:   "+1234",
		LongNumber:    "+123456789012345678",
		NonNumeric:    "+1abc5551234",
		MissingPlus:   "13105551234",
		WithSpaces:    "+1 310 555 1234",
		WithDashes:    "+1-310-555-1234",
	}

	testPLMN = struct {
		TMobileUS   glide.PLMN
		InvalidMCC  glide.PLMN
		InvalidMNC  glide.PLMN
		UnknownPLMN glide.PLMN
		TestLab     glide.PLMN
	}{
		TMobileUS:   glide.PLMN{MCC: "310", MNC: "260"},
		InvalidMCC:  glide.PLMN{MCC: "31", MNC: "260"}, // Too short
		InvalidMNC:  glide.PLMN{MCC: "310", MNC: "2"},  // Too short
		UnknownPLMN: glide.PLMN{MCC: "425", MNC: "01"}, // Israeli PLMN
		TestLab:     glide.PLMN{MCC: "999", MNC: "99"}, // Unofficial MCC
	}
)

// createTestClient creates a client connected to the real API
func createTestClient(t *testing.T) *glide.Client {
	// Try to get API key from environment or use the one from quickstart .env
	apiKey := os.Getenv("GLIDE_API_KEY")
	if apiKey == "" {
		// Try to read from quickstart .env file
		envFile := "../../magical-auth-quickstart-react/.env"
		if data, err := os.ReadFile(envFile); err == nil {
			// Simple parsing - just find GLIDE_API_KEY line
			lines := string(data)
			if idx := strings.Index(lines, "GLIDE_API_KEY="); idx != -1 {
				start := idx + len("GLIDE_API_KEY=")
				end := strings.IndexByte(lines[start:], '\n')
				if end == -1 {
					end = len(lines[start:])
				}
				apiKey = lines[start : start+end]
			}
		}
	}

	if apiKey == "" {
		apiKey = "0NrxSqb7oWXzZXo7Cq25hgvwvwN60lP2MYERLlFrVyaKPiJB" // Default test key
	}

	apiBaseURL := os.Getenv("GLIDE_API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.glideidentity.app"
	}

	client := glide.New(
		glide.WithAPIKey(apiKey),
		glide.WithBaseURL(apiBaseURL),
	)

	return client
}

func TestGetPhoneNumberFlow(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should successfully prepare get phone number request", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			// No phone number - we're trying to get it
			PLMN: &testPLMN.TMobileUS,
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AuthenticationStrategy)
		assert.NotEmpty(t, result.Session)
		assert.NotNil(t, result.Data)

		// Should be either TS43 or Link strategy
		assert.Contains(t, []glide.AuthenticationStrategy{
			glide.AuthenticationStrategyTS43,
			glide.AuthenticationStrategyLink,
		}, result.AuthenticationStrategy)
	})

	t.Run("should handle prepare with consent data", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TMobileUS,
			ConsentData: &glide.ConsentData{
				ConsentText: "I agree to the terms",
				PolicyLink:  "https://example.com/privacy",
				PolicyText:  "Privacy policy text",
			},
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.NoError(t, err)
		assert.NotNil(t, result)
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
	})

	// Removed "should reject phone number for GetPhoneNumber" test
	// to match Node.js SDK test suite - Node.js doesn't have this test
}

func TestVerifyPhoneNumberFlow(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should successfully prepare verify phone number request", func(t *testing.T) {
		prepReq := glide.PrepareRequest{
			UseCase:     glide.UseCaseVerifyPhoneNumber,
			PhoneNumber: testPhoneNumbers.TMobileValid,
			// No PLMN needed - server uses TelcoFinder with phone number
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		result, err := client.MagicAuth.Prepare(ctx, &prepReq)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AuthenticationStrategy)
		assert.NotEmpty(t, result.Session)
	})
}

func TestPerformance(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	t.Run("should complete prepare request within reasonable time", func(t *testing.T) {
		start := time.Now()

		prepReq := glide.PrepareRequest{
			UseCase: glide.UseCaseGetPhoneNumber,
			PLMN:    &testPLMN.TMobileUS,
			ClientInfo: &glide.ClientInfo{
				UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			},
		}

		_, err := client.MagicAuth.Prepare(ctx, &prepReq)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 2*time.Second, "Request took too long: %v", duration)
	})

	t.Run("should handle multiple concurrent requests", func(t *testing.T) {
		const numRequests = 3
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				prepReq := glide.PrepareRequest{
					UseCase: glide.UseCaseGetPhoneNumber,
					PLMN:    &testPLMN.TMobileUS,
					ClientInfo: &glide.ClientInfo{
						UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
					},
				}
				_, err := client.MagicAuth.Prepare(ctx, &prepReq)
				results <- err
			}()
		}

		successCount := 0
		for i := 0; i < numRequests; i++ {
			if err := <-results; err == nil {
				successCount++
			}
		}

		assert.Greater(t, successCount, 0, "At least one request should succeed")
	})
}
