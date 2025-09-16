package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glideidentity/glide-go-sdk/glide"
)

// CustomLogger demonstrates how to implement a custom logger
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) Debug(msg string, fields ...glide.Field) {
	l.log("DEBUG", msg, fields...)
}

func (l *CustomLogger) Info(msg string, fields ...glide.Field) {
	l.log("INFO", msg, fields...)
}

func (l *CustomLogger) Warn(msg string, fields ...glide.Field) {
	l.log("WARN", msg, fields...)
}

func (l *CustomLogger) Error(msg string, fields ...glide.Field) {
	l.log("ERROR", msg, fields...)
}

func (l *CustomLogger) log(level, msg string, fields ...glide.Field) {
	// Custom formatting for your logging system
	fmt.Printf("%s[%s][%s] %s", l.prefix, time.Now().Format("15:04:05"), level, msg)
	for _, f := range fields {
		fmt.Printf(" %s=%v", f.Key, f.Value)
	}
	fmt.Println()
}

func main() {
	fmt.Println("=== Glide SDK Debug Logging Examples ===")

	// Example 1: Enable debug logging with environment variable
	fmt.Println("1. Using environment variable:")
	os.Setenv("GLIDE_DEBUG", "true")
	client1 := glide.New(
		glide.WithAPIKey("test-api-key"),
	)
	// This will log at debug level automatically
	_ = client1 // Example usage
	fmt.Println()

	// Example 2: Enable debug logging programmatically
	fmt.Println("2. Using WithDebug option:")
	client2 := glide.New(
		glide.WithAPIKey("test-api-key"),
		glide.WithDebug(true), // Enable debug logging
	)
	_ = client2
	fmt.Println()

	// Example 3: Set specific log level
	fmt.Println("3. Using WithLogLevel option:")
	client3 := glide.New(
		glide.WithAPIKey("test-api-key"),
		glide.WithLogLevel(glide.LogLevelInfo), // Only log info and above
	)
	_ = client3
	fmt.Println()

	// Example 4: Use custom logger
	fmt.Println("4. Using custom logger:")
	customLogger := &CustomLogger{prefix: "[MyApp]"}
	client4 := glide.New(
		glide.WithAPIKey("test-api-key"),
		glide.WithLogger(customLogger), // Use custom logger
	)
	_ = client4
	fmt.Println()

	// Example 5: Environment variable for log level
	fmt.Println("5. Using GLIDE_LOG_LEVEL environment variable:")
	os.Setenv("GLIDE_LOG_LEVEL", "info")
	client5 := glide.New(
		glide.WithAPIKey("test-api-key"),
	)
	_ = client5
	fmt.Println()

	// Example 6: Disable logging completely
	fmt.Println("6. Disable all logging:")
	client6 := glide.New(
		glide.WithAPIKey("test-api-key"),
		glide.WithLogLevel(glide.LogLevelSilent), // No logging
	)
	_ = client6
	fmt.Println()

	// Example 7: Real usage with Magic Auth
	fmt.Println("7. Real usage example with Magic Auth:")

	// Enable debug for troubleshooting
	client := glide.New(
		glide.WithAPIKey("your-api-key"),
		glide.WithDebug(true),
		glide.WithBaseURL("https://api.glideidentity.app"),
	)

	ctx := context.Background()

	// This will log:
	// - Request preparation
	// - HTTP request details
	// - Response status
	// - Any errors or retries
	// - Performance metrics

	prepareReq := &glide.PrepareRequest{
		PhoneNumber: "+1234567890", // Will be automatically redacted in logs
		UseCase:     glide.UseCaseVerifyPhoneNumber,
	}

	fmt.Println("\nMaking API call with debug logging enabled...")
	_, err := client.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		// Error will be logged automatically with context
		log.Printf("API call failed: %v", err)
	}

	fmt.Println("\n=== Debug Logging Benefits ===")
	fmt.Println("• Troubleshoot integration issues quickly")
	fmt.Println("• Monitor API performance and latency")
	fmt.Println("• Debug authentication problems")
	fmt.Println("• Track retry behavior")
	fmt.Println("• Identify rate limiting issues")
	fmt.Println("• Integrate with your existing logging infrastructure")

	fmt.Println("\n=== Security Features ===")
	fmt.Println("• Automatic redaction of sensitive data:")
	fmt.Println("  - API keys show only first 4 chars")
	fmt.Println("  - Phone numbers show only area code")
	fmt.Println("  - Tokens and passwords are fully redacted")
	fmt.Println("  - Emails show only domain")
}
