// Example usage of the Magic Auth Go SDK.
//
// This example demonstrates:
// 1. Creating a client with configuration
// 2. Preparing authentication
// 3. Verifying a phone number
// 4. Checking session status
//
// Run the mock server first:
//
//	go run ./example/mockserver
//
// Then run this example:
//
//	go run ./example
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/GlideIdentity/glide-be-sdk-go/v3/magicalauth"
)

func main() {
	// Configuration - use environment variables or defaults for mock server
	clientID := getEnv("MAGIC_AUTH_CLIENT_ID", "test")
	clientSecret := getEnv("MAGIC_AUTH_CLIENT_SECRET", "secret")
	baseURL := getEnv("MAGIC_AUTH_BASE_URL", "http://localhost:8080")

	fmt.Println("=== Magic Auth Go SDK Example ===")
	fmt.Printf("Base URL: %s\n\n", baseURL)

	// Create client with custom logger
	client, err := magicalauth.NewClient(magicalauth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      baseURL,
		Logger:       &exampleLogger{},
		Timeout:      10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// --- Example 1: Prepare Authentication ---
	fmt.Println("--- 1. Prepare Authentication ---")

	prepareResp, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
		PhoneNumber: "+14155551234",
		Nonce:       "example-nonce-" + fmt.Sprint(time.Now().Unix()),
		UseCase:     magicalauth.UseCaseVerifyPhoneNumber,
		ClientInfo: magicalauth.ClientInfo{
			UserAgent: "MagicalAuth-Go-Example/1.0", // In production, use actual browser UA from frontend
		},
	})
	if err != nil {
		handleError("Prepare", err)
		return
	}

	fmt.Printf("Strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("Session Key: %s\n", prepareResp.Session.SessionKey)
	fmt.Printf("Protocol: %s\n", prepareResp.Session.ProtocolType)
	fmt.Println()

	// --- Example 2: Report Invocation ---
	fmt.Println("--- 2. Report Invocation ---")

	reportResp, err := client.ReportInvocation(ctx, &magicalauth.ReportInvocationRequest{
		SessionID: prepareResp.Session.SessionKey,
	})
	if err != nil {
		handleError("ReportInvocation", err)
		return
	}

	fmt.Printf("Report success: %v\n", reportResp.Success)
	fmt.Println()

	// --- Example 3: Check Status (before verification) ---
	fmt.Println("--- 3. Check Status (before verification) ---")

	statusResp, err := client.CheckStatus(ctx, prepareResp.Session.SessionKey)
	if err != nil {
		handleError("CheckStatus", err)
		return
	}

	fmt.Printf("Status: %s\n", statusResp.Status)
	fmt.Printf("Protocol: %s\n", statusResp.Protocol)
	fmt.Println()

	// --- Example 4: Verify Phone Number ---
	fmt.Println("--- 4. Verify Phone Number ---")

	// In real usage, the credential comes from the device after the user
	// completes the TS43/Link flow. Here we use a mock credential.
	verifyResp, err := client.VerifyPhoneNumber(ctx, &magicalauth.VerifyPhoneNumberRequest{
		Session:    prepareResp.Session,
		Credential: "mock-sd-jwt-credential",
	})
	if err != nil {
		handleError("VerifyPhoneNumber", err)
		return
	}

	fmt.Printf("Phone Number: %s\n", verifyResp.PhoneNumber)
	fmt.Printf("Verified: %v\n", verifyResp.Verified)
	if verifyResp.SimSwap != nil {
		fmt.Printf("SIM Swap Risk: %s\n", verifyResp.SimSwap.RiskLevel)
		fmt.Printf("SIM Swap Age: %s\n", verifyResp.SimSwap.AgeBand)
	}
	fmt.Println()

	// --- Example 5: Check Status (after verification) ---
	fmt.Println("--- 5. Check Status (after verification) ---")

	statusResp, err = client.CheckStatusPublic(ctx, prepareResp.Session.SessionKey)
	if err != nil {
		handleError("CheckStatusPublic", err)
		return
	}

	fmt.Printf("Status: %s\n", statusResp.Status)
	fmt.Printf("Protocol: %s\n", statusResp.Protocol)
	fmt.Println()

	// --- Example 6: Get Phone Number (alternative use case) ---
	fmt.Println("--- 6. Get Phone Number (alternative flow) ---")

	// First, prepare for GetPhoneNumber use case
	prepareResp2, err := client.Prepare(ctx, &magicalauth.PrepareRequest{
		Nonce:   "get-phone-nonce-" + fmt.Sprint(time.Now().Unix()),
		UseCase: magicalauth.UseCaseGetPhoneNumber,
		PLMN: &magicalauth.PLMN{
			MCC: "310",
			MNC: "260",
		},
		ClientInfo: magicalauth.ClientInfo{
			UserAgent: "MagicalAuth-Go-Example/1.0", // In production, use actual browser UA from frontend
		},
	})
	if err != nil {
		handleError("Prepare (GetPhoneNumber)", err)
		return
	}

	getPhoneResp, err := client.GetPhoneNumber(ctx, &magicalauth.GetPhoneNumberRequest{
		Session:    prepareResp2.Session,
		Credential: "mock-sd-jwt-credential",
	})
	if err != nil {
		handleError("GetPhoneNumber", err)
		return
	}

	fmt.Printf("Retrieved Phone: %s\n", getPhoneResp.PhoneNumber)
	fmt.Println()

	// --- Example 7: Error Handling ---
	fmt.Println("--- 7. Error Handling Example ---")

	// Try to check status for non-existent session
	_, err = client.CheckStatus(ctx, "nonexistent-session-key")
	if err != nil {
		handleError("CheckStatus (expected error)", err)
	}

	fmt.Println("\n=== Example Complete ===")
}

// handleError demonstrates proper error handling with the SDK
func handleError(operation string, err error) {
	fmt.Printf("[%s] Error: %v\n", operation, err)

	// Check for specific error types using errors.Is()
	switch {
	case errors.Is(err, magicalauth.ErrSessionNotFound):
		fmt.Println("  -> Session not found or expired")
	case errors.Is(err, magicalauth.ErrCarrierNotEligible):
		fmt.Println("  -> Carrier is not supported")
	case errors.Is(err, magicalauth.ErrRateLimit):
		fmt.Println("  -> Rate limited, try again later")
	case errors.Is(err, magicalauth.ErrUnauthorized):
		fmt.Println("  -> Check your credentials")
	}

	// Extract detailed error information using errors.As()
	var apiErr *magicalauth.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("  -> Code: %s\n", apiErr.Code)
		fmt.Printf("  -> Status: %d\n", apiErr.Status)
		if apiErr.RequestID != "" {
			fmt.Printf("  -> RequestID: %s\n", apiErr.RequestID)
		}
		if apiErr.RetryAfter > 0 {
			fmt.Printf("  -> Retry After: %d seconds\n", apiErr.RetryAfter)
		}
	}
	fmt.Println()
}

// exampleLogger implements magicalauth.Logger
type exampleLogger struct{}

func (l *exampleLogger) Debug(msg string, keysAndValues ...any) {
	log.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

func (l *exampleLogger) Info(msg string, keysAndValues ...any) {
	log.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (l *exampleLogger) Warn(msg string, keysAndValues ...any) {
	log.Printf("[WARN] %s %v", msg, keysAndValues)
}

func (l *exampleLogger) Error(msg string, keysAndValues ...any) {
	log.Printf("[ERROR] %s %v", msg, keysAndValues)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
