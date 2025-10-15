package glide

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogFormat specifies the output format for logs
type LogFormat string

const (
	// LogFormatPretty provides colored, formatted output with boxes (default)
	LogFormatPretty LogFormat = "pretty"
	// LogFormatJSON outputs structured JSON logs
	LogFormatJSON LogFormat = "json"
	// LogFormatSimple provides simple text output without colors
	LogFormatSimple LogFormat = "simple"
)

// ANSI color codes for terminal output
var colors = struct {
	Reset   string
	Bright  string
	Dim     string
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	White   string
	Gray    string
}{
	Reset:   "\x1b[0m",
	Bright:  "\x1b[1m",
	Dim:     "\x1b[2m",
	Red:     "\x1b[31m",
	Green:   "\x1b[32m",
	Yellow:  "\x1b[33m",
	Blue:    "\x1b[34m",
	Magenta: "\x1b[35m",
	Cyan:    "\x1b[36m",
	White:   "\x1b[37m",
	Gray:    "\x1b[90m",
}

// supportsColor checks if the environment supports colored output
func supportsColor() bool {
	// Check for NO_COLOR env variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Force colors if explicitly requested
	if os.Getenv("FORCE_COLOR") == "true" || os.Getenv("FORCE_COLOR") == "1" {
		return true
	}

	// For npm scripts and Node.js environments
	if os.Getenv("npm_config_color") != "false" {
		return true
	}

	// Check common terminal emulators that support color
	term := os.Getenv("TERM")
	if term != "" && term != "dumb" {
		// Most modern terminals support color
		return true
	}

	// Check if running on Windows with color support
	if runtime.GOOS == "windows" {
		// Windows Terminal and modern terminals
		if os.Getenv("WT_SESSION") != "" ||
			os.Getenv("TERMINAL_EMULATOR") != "" ||
			os.Getenv("ANSICON") != "" {
			return true
		}
	}

	// Check for common color-supporting environments
	if os.Getenv("COLORTERM") != "" {
		return true
	}

	// Check for CI environments that support color
	if os.Getenv("CI") == "true" {
		// GitHub Actions, CircleCI, Travis all support colors
		return os.Getenv("GITHUB_ACTIONS") == "true" ||
			os.Getenv("CIRCLECI") == "true" ||
			os.Getenv("TRAVIS") == "true"
	}

	// Default to true for Unix-like systems with TTY
	if runtime.GOOS != "windows" {
		return true // Always enable on macOS/Linux
	}

	return false
}

var useColors = supportsColor()

// colorize applies color to text if colors are supported
func colorize(text, color string) string {
	if !useColors {
		return text
	}
	return color + text + colors.Reset
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

// formatDuration formats duration to human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// createBox creates a formatted box around content
func createBox(title string, content []string, color string) string {
	// Calculate max width
	// Account for special unicode characters that may display wider
	titleDisplayLen := getDisplayWidth(title)
	maxLen := titleDisplayLen + 2 // Add padding for title

	for _, line := range content {
		lineLen := getDisplayWidth(line) + 2 // Add padding for content
		if lineLen > maxLen {
			maxLen = lineLen
		}
	}
	width := min(maxLen+2, 80) // +2 for borders

	var lines []string

	// Top border
	lines = append(lines, colorize("┌"+strings.Repeat("─", width-2)+"┐", color))

	// Title with proper padding
	// For the title line, we need to match the width exactly
	// Structure: "│ " + title + " " + padding + "│"
	titleText := " " + title + " "
	titleLen := len(stripANSI(titleText))

	// Calculate padding needed to reach the target width
	// The title line needs to match the width of the top border
	// Top border has width characters total
	// Title line: │ (1) + titleText + padding + │ (1) = width
	// So: 1 + titleLen + padding + 1 = width
	// Therefore: padding = width - titleLen - 2
	// But we need two extra spaces for proper alignment
	titlePadding := width - titleLen // No subtraction, adds the needed spaces

	if titlePadding < 0 {
		// If title is too long, truncate it
		maxTitleLen := width - 4 // Leave room for borders and spaces
		if maxTitleLen > 0 && len(title) > maxTitleLen-2 {
			title = title[:maxTitleLen-5] + "..."
			titleText = " " + title + " "
			titleLen = len(titleText)
			titlePadding = width - titleLen // Also no subtraction here for consistency
		}
		if titlePadding < 0 {
			titlePadding = 0
		}
	}

	// Build the title line
	lines = append(lines,
		colorize("│", color)+
			colorize(titleText, colors.Bright+color)+
			strings.Repeat(" ", titlePadding)+
			colorize("│", color))

	// Separator
	lines = append(lines, colorize("├"+strings.Repeat("─", width-2)+"┤", color))

	// Content
	for _, line := range content {
		lineText := " " + line + " "
		cleanLine := stripANSI(lineText)
		padding := width - len(cleanLine) - 2 // -2 for borders
		if padding < 0 {
			padding = 0
		}
		lines = append(lines,
			colorize("│", color)+
				lineText+
				strings.Repeat(" ", padding)+
				colorize("│", color))
	}

	// Bottom border
	lines = append(lines, colorize("└"+strings.Repeat("─", width-2)+"┘", color))

	return strings.Join(lines, "\n")
}

// stripANSI removes ANSI color codes from text
func stripANSI(text string) string {
	// Simple ANSI stripping - removes color codes
	result := text
	for {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// getDisplayWidth calculates the display width of a string accounting for unicode characters
func getDisplayWidth(text string) int {
	// Strip ANSI codes first
	clean := stripANSI(text)

	// For now, just return the length - the checkmark seems to be 1 char wide in most terminals
	// We may need to revisit this for other unicode characters
	return len(clean)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LogFormatter provides formatted console output
type LogFormatter struct {
	format LogFormat
	prefix string
}

// NewLogFormatter creates a new log formatter
func NewLogFormatter(format LogFormat, prefix string) *LogFormatter {
	return &LogFormatter{
		format: format,
		prefix: prefix,
	}
}

// FormatRequest formats an API request for logging
func (f *LogFormatter) FormatRequest(method, url string, details map[string]interface{}) {
	switch f.format {
	case LogFormatJSON:
		f.formatRequestJSON(method, url, details)
	case LogFormatSimple:
		f.formatRequestSimple(method, url, details)
	case LogFormatPretty:
		fallthrough
	default:
		f.formatRequestPretty(method, url, details)
	}
}

// FormatResponse formats an API response for logging
func (f *LogFormatter) FormatResponse(operation string, status int, details map[string]interface{}) {
	switch f.format {
	case LogFormatJSON:
		f.formatResponseJSON(operation, status, details)
	case LogFormatSimple:
		f.formatResponseSimple(operation, status, details)
	case LogFormatPretty:
		fallthrough
	default:
		f.formatResponsePretty(operation, status, details)
	}
}

// Pretty format implementations
func (f *LogFormatter) formatRequestPretty(method, url string, details map[string]interface{}) {
	// Extract operation name from URL
	operation := f.getOperationName(url)

	// Build content lines
	var content []string

	// Add timestamp
	content = append(content, fmt.Sprintf("Time: %s", time.Now().Format("15:04:05")))

	// Add URL (truncate if too long)
	urlDisplay := url
	if len(url) > 70 {
		urlDisplay = url[:67] + "..."
	}
	content = append(content, fmt.Sprintf("URL: %s", urlDisplay))

	// Add method
	content = append(content, fmt.Sprintf("Method: %s", method))

	// Add details
	if useCase, ok := details["use_case"].(string); ok {
		content = append(content, fmt.Sprintf("Use Case: %s", useCase))
	}
	if plmn, ok := details["plmn"].(map[string]interface{}); ok {
		if mcc, ok := plmn["mcc"].(string); ok {
			if mnc, ok := plmn["mnc"].(string); ok {
				content = append(content, fmt.Sprintf("PLMN: MCC=%s, MNC=%s", mcc, mnc))
			}
		}
	}
	if bodySize, ok := details["body_size"].(int); ok {
		content = append(content, fmt.Sprintf("Body Size: %s", formatBytes(bodySize)))
	}

	// Create and print the box
	box := createBox("→ "+operation, content, colors.Cyan)
	fmt.Println()
	fmt.Println(box)
}

func (f *LogFormatter) formatResponsePretty(operation string, status int, details map[string]interface{}) {
	// Determine success/error and color
	var symbol, color string
	if status >= 200 && status < 300 {
		symbol = "✓"
		color = colors.Green
	} else {
		symbol = "✗"
		color = colors.Red
	}

	// Build content lines
	var content []string
	content = append(content, fmt.Sprintf("Status: %d", status))

	// Add operation-specific details
	if phoneNumber, ok := details["phone_number"].(string); ok {
		content = append(content, fmt.Sprintf("Phone Number: %s", phoneNumber))
	}
	if verified, ok := details["verified"].(bool); ok {
		content = append(content, fmt.Sprintf("Verified: %t", verified))
	}
	if strategy, ok := details["strategy"].(string); ok {
		content = append(content, fmt.Sprintf("Strategy: %s", strategy))
	}
	if sessionKey, ok := details["session_key"].(string); ok {
		// Format like Node SDK: show full key twice with ellipsis
		content = append(content, fmt.Sprintf("Session Key: %s...%s", sessionKey, sessionKey))
	}

	// Create and print the box
	title := fmt.Sprintf("%s %s Response", symbol, operation)
	box := createBox(title, content, color)
	fmt.Println(box)
	fmt.Println()
}

// Simple format implementations
func (f *LogFormatter) formatRequestSimple(method, url string, details map[string]interface{}) {
	fmt.Printf("[%s] %s %s", time.Now().Format("15:04:05"), method, url)
	if len(details) > 0 {
		if jsonBytes, err := json.Marshal(details); err == nil {
			fmt.Printf(" %s", string(jsonBytes))
		}
	}
	fmt.Println()
}

func (f *LogFormatter) formatResponseSimple(operation string, status int, details map[string]interface{}) {
	fmt.Printf("[%s] Response %d", time.Now().Format("15:04:05"), status)
	if len(details) > 0 {
		if jsonBytes, err := json.Marshal(details); err == nil {
			fmt.Printf(" %s", string(jsonBytes))
		}
	}
	fmt.Println()
}

// JSON format implementations
func (f *LogFormatter) formatRequestJSON(method, url string, details map[string]interface{}) {
	logObj := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      "request",
		"method":    method,
		"url":       url,
		"details":   details,
	}
	if jsonBytes, err := json.Marshal(logObj); err == nil {
		fmt.Println(string(jsonBytes))
	}
}

func (f *LogFormatter) formatResponseJSON(operation string, status int, details map[string]interface{}) {
	logObj := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      "response",
		"operation": operation,
		"status":    status,
		"details":   details,
	}
	if jsonBytes, err := json.Marshal(logObj); err == nil {
		fmt.Println(string(jsonBytes))
	}
}

// getOperationName extracts operation name from URL
func (f *LogFormatter) getOperationName(url string) string {
	if strings.Contains(url, "prepare") {
		return "MagicAuth PREPARE"
	} else if strings.Contains(url, "verify-phone-number") {
		return "MagicAuth VERIFY PHONE"
	} else if strings.Contains(url, "get-phone-number") {
		return "MagicAuth GET PHONE"
	} else if strings.Contains(url, "sim-swap") {
		if strings.Contains(url, "check") {
			return "SimSwap CHECK"
		}
		return "SimSwap RETRIEVE DATE"
	} else if strings.Contains(url, "kyc-match") {
		return "KYC MATCH"
	}
	return "API Request"
}
