package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/GlideIdentity/glide-be-sdk-go/glide"
)

func main() {
	fmt.Println("🚀 Glide Go SDK - Local Server Example")
	fmt.Println("=======================================")
	fmt.Println("This example connects to a local Glide server instead of production API")
	fmt.Println()

	// Get API key from environment
	apiKey := os.Getenv("GLIDE_API_KEY")
	if apiKey == "" {
		log.Fatal("GLIDE_API_KEY environment variable is required")
	}

	// Initialize client pointing to local server
	client := glide.New(
		glide.WithAPIKey(apiKey),
		glide.WithBaseURL("http://localhost:8080"), // Local server
		glide.WithTimeout(30*time.Second),
		glide.WithDebug(true), // Enable debug logging
	)

	// Test verify phone number flow
	fmt.Println("📱 Testing Verify Phone Number")
	fmt.Println("------------------------------")
	if err := testVerifyPhoneNumber(client); err != nil {
		log.Printf("❌ Error: %v\n", err)
	}

	fmt.Println()

	// Test get phone number flow
	fmt.Println("📞 Testing Get Phone Number")
	fmt.Println("---------------------------")
	if err := testGetPhoneNumber(client); err != nil {
		log.Printf("❌ Error: %v\n", err)
	}
}

func testVerifyPhoneNumber(client *glide.Client) error {
	ctx := context.Background()

	// Prepare authentication for a T-Mobile US number
	prepareReq := &glide.PrepareRequest{
		PhoneNumber: "+13102223333", // Example T-Mobile number
		UseCase:     glide.UseCaseVerifyPhoneNumber,
		ConsentData: &glide.ConsentData{
			ConsentText: "I agree to verify my phone number",
			PolicyLink:  "https://example.com/privacy",
		},
	}

	fmt.Println("🔄 Calling prepare endpoint...")
	prepareResp, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		// Handle known errors gracefully
		if glideErr, ok := err.(*glide.Error); ok {
			switch glideErr.Code {
			case glide.ErrCodeCarrierNotEligible:
				fmt.Printf("⚠️  Carrier not eligible: %s\n", glideErr.Message)
				fmt.Printf("   Request ID: %s\n", glideErr.RequestID)
				return nil // Not really an error for testing
			case glide.ErrCodeInternalServerError:
				return fmt.Errorf("API key is invalid or missing")
			default:
				return fmt.Errorf("API error [%s]: %s", glideErr.Code, glideErr.Message)
			}
		}
		return fmt.Errorf("prepare failed: %v", err)
	}

	fmt.Printf("✅ Prepare successful!\n")
	fmt.Printf("   Strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("   Session Key: %s\n", prepareResp.Session.SessionKey)
	if prepareResp.TTL > 0 {
		fmt.Printf("   TTL: %d seconds\n", prepareResp.TTL)
	}

	// In a real scenario, the client would perform authentication
	// For this example, we'll simulate it
	fmt.Println("\n🔐 Simulating client authentication...")
	time.Sleep(1 * time.Second)

	// Verify the phone number
	verifyReq := &glide.VerifyPhoneNumberRequest{
		Session: &prepareResp.Session,
		Credential: map[string]interface{}{
			"vp_token": "simulated-credential-token",
		},
	}

	fmt.Println("🔄 Calling verify endpoint...")
	verifyResp, err := client.MagicAuth.VerifyPhoneNumber(ctx, verifyReq)
	if err != nil {
		if glideErr, ok := err.(*glide.Error); ok {
			return fmt.Errorf("verify error [%s]: %s", glideErr.Code, glideErr.Message)
		}
		return fmt.Errorf("verification failed: %v", err)
	}

	fmt.Printf("✅ Phone number verified!\n")
	fmt.Printf("   Number: %s\n", verifyResp.PhoneNumber)
	fmt.Printf("   Verified: %v\n", verifyResp.Verified)

	return nil
}

func testGetPhoneNumber(client *glide.Client) error {
	ctx := context.Background()

	// Prepare with PLMN for T-Mobile US
	prepareReq := &glide.PrepareRequest{
		PLMN: &glide.PLMN{
			MCC: "310", // United States
			MNC: "260", // T-Mobile
		},
		UseCase: glide.UseCaseGetPhoneNumber,
	}

	fmt.Println("🔄 Calling prepare endpoint with PLMN...")
	prepareResp, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		if glideErr, ok := err.(*glide.Error); ok {
			switch glideErr.Code {
			case glide.ErrCodeCarrierNotEligible:
				fmt.Printf("⚠️  Carrier not eligible: %s\n", glideErr.Message)
				return nil
			default:
				return fmt.Errorf("API error [%s]: %s", glideErr.Code, glideErr.Message)
			}
		}
		return fmt.Errorf("prepare failed: %v", err)
	}

	fmt.Printf("✅ Prepare successful!\n")
	fmt.Printf("   Strategy: %s\n", prepareResp.AuthenticationStrategy)
	fmt.Printf("   Session Key: %s\n", prepareResp.Session.SessionKey)

	// Simulate authentication
	fmt.Println("\n🔐 Simulating client authentication...")
	time.Sleep(1 * time.Second)

	// Get the phone number
	getReq := &glide.GetPhoneNumberRequest{
		Session: &prepareResp.Session,
		Credential: map[string]interface{}{
			"vp_token": "simulated-token-with-phone-claim",
		},
	}

	fmt.Println("🔄 Calling get phone number endpoint...")
	getResp, err := client.MagicAuth.GetPhoneNumber(ctx, getReq)
	if err != nil {
		if glideErr, ok := err.(*glide.Error); ok {
			return fmt.Errorf("get phone error [%s]: %s", glideErr.Code, glideErr.Message)
		}
		return fmt.Errorf("get phone number failed: %v", err)
	}

	fmt.Printf("✅ Phone number retrieved!\n")
	fmt.Printf("   Number: %s\n", getResp.PhoneNumber)

	return nil
}
