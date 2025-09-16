package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/glideidentity/glide-go-sdk/glide"
)

func main() {
	// Initialize the Glide client with optional rate limiting
	client := glide.New(
		glide.WithAPIKey("your-api-key"),
		glide.WithBaseURL("https://api.glideidentity.app"),
		glide.WithTimeout(30*time.Second),
		glide.WithRetry(3, time.Second),
		glide.WithRateLimit(100, time.Second), // 100 requests per second
	)

	// Example 1: Verify a phone number
	if err := verifyPhoneNumber(client); err != nil {
		log.Printf("Verification failed: %v", err)
	}

	// Example 2: Get phone number
	if err := getPhoneNumber(client); err != nil {
		log.Printf("Get phone number failed: %v", err)
	}

	// Example 3: Check SIM swap
	if err := checkSimSwap(client); err != nil {
		log.Printf("SIM swap check failed: %v", err)
	}

	// Example 4: KYC matching
	if err := performKYCMatch(client); err != nil {
		log.Printf("KYC match failed: %v", err)
	}
}

func verifyPhoneNumber(client *glide.Client) error {
	ctx := context.Background()

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

	fmt.Printf("Authentication prepared with strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("Session: %s\n", prepareResp.Session)

	// Step 2: Client performs authentication (browser/app)
	// This would happen in your frontend using the Web SDK
	// For this example, we'll simulate the credential response
	credentialResponse := map[string]interface{}{
		"vp_token": "simulated-token",
	}

	// Step 3: Verify the phone number
	verifyReq := &glide.VerifyPhoneNumberRequest{
		SessionInfo: &prepareResp.Session,
		Credential:  credentialResponse,
	}

	verifyResp, err := client.MagicAuth.VerifyPhoneNumber(ctx, verifyReq)
	if err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	fmt.Printf("Phone number verified: %s (verified: %v)\n",
		verifyResp.PhoneNumber, verifyResp.Verified)

	return nil
}

func getPhoneNumber(client *glide.Client) error {
	ctx := context.Background()

	// Prepare with PLMN instead of phone number
	prepareReq := &glide.PrepareRequest{
		PLMN: &glide.PLMN{
			MCC: "310", // USA
			MNC: "260", // T-Mobile
		},
		UseCase: glide.UseCaseGetPhoneNumber,
	}

	prepareResp, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		return fmt.Errorf("prepare failed: %v", err)
	}

	fmt.Printf("Ready to get phone number with strategy: %s\n", prepareResp.AuthenticationStrategy)

	// After client authentication...
	credentialResponse := map[string]interface{}{
		"vp_token": "simulated-token",
	}

	getReq := &glide.GetPhoneNumberRequest{
		SessionInfo: &prepareResp.Session,
		Credential:  credentialResponse,
	}

	getResp, err := client.MagicAuth.GetPhoneNumber(ctx, getReq)
	if err != nil {
		return fmt.Errorf("get phone number failed: %v", err)
	}

	fmt.Printf("Retrieved phone number: %s\n", getResp.PhoneNumber)

	return nil
}

func checkSimSwap(client *glide.Client) error {
	ctx := context.Background()

	// Check if SIM was swapped in last 48 hours
	checkReq := &glide.SimSwapCheckRequest{
		PhoneNumber: "+14155552671",
		MaxAge:      48, // hours
	}

	checkResp, err := client.SimSwap.Check(ctx, checkReq)
	if err != nil {
		return fmt.Errorf("SIM swap check failed: %v", err)
	}

	if checkResp.Swapped {
		fmt.Printf("⚠️  SIM was swapped at: %v\n", checkResp.SwappedAt)
	} else {
		fmt.Println("✅ No recent SIM swap detected")
	}

	// Get last swap date
	dateReq := &glide.SimSwapDateRequest{
		PhoneNumber: "+14155552671",
	}

	dateResp, err := client.SimSwap.GetLastSwapDate(ctx, dateReq)
	if err != nil {
		return fmt.Errorf("get swap date failed: %v", err)
	}

	if dateResp.LastSwapDate != nil {
		fmt.Printf("Last SIM swap: %v\n", dateResp.LastSwapDate)
	} else {
		fmt.Println("No SIM swap history available")
	}

	return nil
}

func performKYCMatch(client *glide.Client) error {
	ctx := context.Background()

	matchReq := &glide.KYCMatchRequest{
		PhoneNumber: "+14155552671",
		Name:        "John Doe",
		BirthDate:   "1990-01-15",
		Email:       "john.doe@example.com",
		Address: &glide.Address{
			Street:     "123 Main St",
			City:       "San Francisco",
			State:      "CA",
			PostalCode: "94102",
			Country:    "US",
		},
	}

	matchResp, err := client.KYC.Match(ctx, matchReq)
	if err != nil {
		return fmt.Errorf("KYC match failed: %v", err)
	}

	fmt.Printf("Overall KYC match: %v\n", matchResp.OverallMatch)

	// Show individual field matches
	for field, result := range matchResp.MatchResults {
		fmt.Printf("  %s: matched=%v (confidence: %s)\n",
			field, result.Matched, result.Confidence)
	}

	return nil
}

// Example: Handling errors properly
func handleErrors() {
	client := glide.New(glide.WithAPIKey("test"))
	ctx := context.Background()

	_, err := client.MagicAuth.Prepare(ctx, &glide.PrepareRequest{
		PhoneNumber: "+1234567890",
		UseCase:     glide.UseCaseVerifyPhoneNumber,
	})

	if err != nil {
		// Type assert to get detailed error info
		if glideErr, ok := err.(*glide.Error); ok {
			fmt.Printf("Error Code: %s\n", glideErr.Code)
			fmt.Printf("Message: %s\n", glideErr.Message)
			fmt.Printf("Request ID: %s\n", glideErr.RequestID)

			// Check if retryable
			if glideErr.IsRetryable() {
				fmt.Println("This error is retryable")
			}

			// Handle specific errors
			switch glideErr.Code {
			case glide.ErrCodeCarrierNotEligible:
				// Use alternative verification method
			case glide.ErrCodeInvalidSessionState:
				// Restart the flow
			case glide.ErrCodeRateLimitExceeded:
				// Wait and retry
				if retryAfter, ok := glideErr.Details["retry_after"].(float64); ok {
					fmt.Printf("Retry after %v seconds\n", retryAfter)
				}
			}
		}
	}
}

// Example: Batch processing with concurrency control
func batchVerification(client *glide.Client, phoneNumbers []string) {
	ctx := context.Background()
	results := make(chan string, len(phoneNumbers))

	// Limit concurrent requests
	semaphore := make(chan struct{}, 10)

	for _, phone := range phoneNumbers {
		go func(phoneNumber string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			req := &glide.NumberVerifyRequest{
				PhoneNumber: phoneNumber,
			}

			resp, err := client.NumberVerify.Verify(ctx, req)
			if err != nil {
				results <- fmt.Sprintf("%s: error - %v", phoneNumber, err)
			} else {
				results <- fmt.Sprintf("%s: verified=%v", phoneNumber, resp.Verified)
			}
		}(phone)
	}

	// Collect results
	for i := 0; i < len(phoneNumbers); i++ {
		fmt.Println(<-results)
	}
}

// Example: Pretty print responses for debugging
func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}
