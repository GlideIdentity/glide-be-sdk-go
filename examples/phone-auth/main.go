package main

import (
	"context"
	"fmt"
	"log"
	"time"

	glide "github.com/glideidentity/glide-go-sdk"
)

// This example demonstrates MagicAuth service specifically
func main() {
	// Initialize the Glide client
	client := glide.New(
		glide.WithAPIKey("your-api-key"),
		glide.WithBaseURL("https://api.glideidentity.app"),
		glide.WithTimeout(30*time.Second),
		glide.WithRetry(3, time.Second),
		glide.WithRateLimit(100, time.Second), // Optional rate limiting
	)

	ctx := context.Background()

	// Example 1: Verify a phone number
	fmt.Println("=== Verifying Phone Number ===")
	if err := verifyPhoneNumber(ctx, client); err != nil {
		log.Printf("Verification failed: %v", err)
	}

	// Example 2: Get phone number using PLMN
	fmt.Println("\n=== Getting Phone Number ===")
	if err := getPhoneNumber(ctx, client); err != nil {
		log.Printf("Get phone number failed: %v", err)
	}
}

func verifyPhoneNumber(ctx context.Context, client *glide.Client) error {
	// Step 1: Prepare authentication
	prepareReq := &glide.PrepareRequest{
		PhoneNumber: "+14155552671",
		UseCase:     glide.UseCaseVerifyPhoneNumber,
		ConsentData: &glide.ConsentData{
			ConsentText: "I agree to verify my phone number",
			PolicyLink:  "https://example.com/privacy",
		},
	}

	prepareResp, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		// Handle specific error types
		if glideErr, ok := err.(*glide.Error); ok {
			switch glideErr.Code {
			case glide.ErrCodeCarrierNotEligible:
				return fmt.Errorf("your device is not eligible for this verification method")
			case glide.ErrCodeRateLimitExceeded:
				return fmt.Errorf("too many requests, please try again later")
			default:
				return fmt.Errorf("prepare failed: %v", glideErr)
			}
		}
		return err
	}

	fmt.Printf("✅ Authentication prepared\n")
	fmt.Printf("   Strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("   Session: %s\n", prepareResp.Session)
	if prepareResp.TTL > 0 {
		fmt.Printf("   TTL: %d seconds\n", prepareResp.TTL)
	}

	// Check which strategy to use
	switch prepareResp.AuthenticationStrategy {
	case glide.AuthenticationStrategyTS43:
		fmt.Println("   → Using TS43 (Digital Credentials API)")
	case glide.AuthenticationStrategyLink:
		fmt.Println("   → Using Link (Deep link/App clip)")
	default:
		fmt.Printf("   → Unknown strategy: %s\n", prepareResp.AuthenticationStrategy)
	}

	// Step 2: Client performs authentication (browser/app)
	// This would happen in your frontend using the Web SDK
	// The frontend would use the Digital Credentials API or deep link
	// based on the strategy returned
	fmt.Println("\n⏳ Client authentication would happen here...")
	fmt.Println("   (In production, this happens in the browser/app)")

	// For this example, we'll simulate the credential response
	credentialResponse := map[string]interface{}{
		"vp_token": "simulated-token-from-carrier",
	}

	// Step 3: Verify the phone number
	verifyReq := &glide.VerifyPhoneNumberRequest{
		Session:    &prepareResp.Session,
		Credential: credentialResponse,
	}

	verifyResp, err := client.MagicAuth.VerifyPhoneNumber(ctx, verifyReq)
	if err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	fmt.Printf("\n✅ Phone number verified!\n")
	fmt.Printf("   Number: %s\n", verifyResp.PhoneNumber)
	fmt.Printf("   Verified: %v\n", verifyResp.Verified)

	return nil
}

func getPhoneNumber(ctx context.Context, client *glide.Client) error {
	// Prepare with PLMN instead of phone number
	// This is useful when you don't know the phone number yet
	prepareReq := &glide.PrepareRequest{
		PLMN: &glide.PLMN{
			MCC: "310", // USA
			MNC: "260", // T-Mobile
		},
		UseCase: glide.UseCaseGetPhoneNumber,
		ConsentData: &glide.ConsentData{
			ConsentText: "I agree to share my phone number",
			PolicyLink:  "https://example.com/privacy",
		},
	}

	prepareResp, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		return fmt.Errorf("prepare failed: %v", err)
	}

	fmt.Printf("✅ Ready to get phone number\n")
	fmt.Printf("   Strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("   Session: %s\n", prepareResp.Session)

	// After client authentication...
	fmt.Println("\n⏳ Client authentication would happen here...")

	// Simulated credential response
	credentialResponse := map[string]interface{}{
		"vp_token": "simulated-token-with-phone-claim",
	}

	getReq := &glide.GetPhoneNumberRequest{
		Session:    &prepareResp.Session,
		Credential: credentialResponse,
	}

	getResp, err := client.MagicAuth.GetPhoneNumber(ctx, getReq)
	if err != nil {
		return fmt.Errorf("get phone number failed: %v", err)
	}

	fmt.Printf("\n✅ Phone number retrieved!\n")
	fmt.Printf("   Number: %s\n", getResp.PhoneNumber)

	return nil
}

// Example: Error handling patterns
func demonstrateErrorHandling(ctx context.Context, client *glide.Client) {
	_, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
		PhoneNumber: "+1234567890",
		UseCase:     glide.UseCaseVerifyPhoneNumber,
	})

	if err != nil {
		// Type assert to get detailed error info
		if glideErr, ok := err.(*glide.Error); ok {
			fmt.Printf("Error Details:\n")
			fmt.Printf("  Code: %s\n", glideErr.Code)
			fmt.Printf("  Message: %s\n", glideErr.Message)
			fmt.Printf("  Request ID: %s\n", glideErr.RequestID)
			fmt.Printf("  Retryable: %v\n", glideErr.IsRetryable())

			// Handle specific errors
			switch glideErr.Code {
			case glide.ErrCodeCarrierNotEligible:
				fmt.Println("→ Fallback: Use alternative verification method")

			case glide.ErrCodeSessionNotFound:
				fmt.Println("→ Action: Restart the authentication flow")

			case glide.ErrCodeRateLimitExceeded:
				if retryAfter, ok := glideErr.Details["retry_after"].(float64); ok {
					fmt.Printf("→ Wait: Retry after %v seconds\n", retryAfter)
				}

			case glide.ErrCodeValidationError:
				fmt.Println("→ Fix: Check request parameters")

			default:
				fmt.Printf("→ Unexpected error: %s\n", glideErr.Code)
			}
		}
	}
}
