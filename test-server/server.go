package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GlideIdentity/glide-be-sdk-go/glide"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Server configuration
type Config struct {
	Port          string
	GlideAPIKey   string
	GlideBaseURL  string
	AllowedOrigin string
	Debug         bool
}

// API response types
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"requestId,omitempty"`
	Timestamp string                 `json:"timestamp"`
	TraceID   string                 `json:"traceId,omitempty"`
	SpanID    string                 `json:"spanId,omitempty"`
	Status    int                    `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type PhoneAuthPrepareRequest struct {
	UseCase     string `json:"use_case"`
	PhoneNumber string `json:"phone_number,omitempty"`
	PLMN        struct {
		MCC string `json:"mcc,omitempty"`
		MNC string `json:"mnc,omitempty"`
	} `json:"plmn,omitempty"`
}

type PhoneAuthPrepareResponse struct {
	Strategy string      `json:"strategy"`
	Session  interface{} `json:"session"`
	Data     interface{} `json:"data,omitempty"`
	TTL      int         `json:"ttl,omitempty"`
}

type PhoneAuthProcessRequest struct {
	Response    interface{}            `json:"response"`
	Session     string                 `json:"session,omitempty"`
	SessionInfo map[string]interface{} `json:"sessionInfo,omitempty"` // New: full session info
}

type PhoneAuthProcessResponse struct {
	Success     bool   `json:"success"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Verified    bool   `json:"verified"`
	Error       string `json:"error,omitempty"`
}

// Global variables
var (
	cfg         Config
	glideClient *glide.Client
)

func main() {
	// Initialize configuration
	cfg = Config{
		Port:          getEnv("PORT", "3001"),
		GlideAPIKey:   getEnv("GLIDE_API_KEY", ""),
		GlideBaseURL:  getEnv("GLIDE_API_BASE_URL", "https://api.glideidentity.app"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		Debug:         getEnv("GLIDE_DEBUG", "false") == "true" || getEnv("GLIDE_LOG_LEVEL", "") == "debug",
	}

	if cfg.GlideAPIKey == "" {
		log.Println("Warning: GLIDE_API_KEY not set, using test mode")
		cfg.GlideAPIKey = "test-api-key"
	}

	// Initialize Glide client
	clientOpts := []glide.Option{
		glide.WithAPIKey(cfg.GlideAPIKey),
		glide.WithBaseURL(cfg.GlideBaseURL),
		glide.WithTimeout(30 * time.Second),
		glide.WithRetry(3, time.Second),
	}

	if cfg.Debug {
		clientOpts = append(clientOpts, glide.WithDebug(true))
		log.Println("Debug mode enabled")
	}

	glideClient = glide.New(clientOpts...)

	// Set up routes
	router := mux.NewRouter()

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/phone-auth/prepare", prepareHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/phone-auth/verify", verifyHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/phone-auth/get", getPhoneHandler).Methods("POST", "OPTIONS")

	// CORS middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{cfg.AllowedOrigin}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Logging middleware
	loggingHandler := handlers.LoggingHandler(os.Stdout, corsHandler(router))

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Go SDK Test Server starting on http://localhost%s", addr)
	log.Printf("📝 API endpoint: http://localhost%s/api", addr)
	log.Printf("🔗 Glide API URL: %s", cfg.GlideBaseURL)
	log.Printf("🔍 Debug mode: %v", cfg.Debug)

	if err := http.ListenAndServe(addr, loggingHandler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// Prepare authentication endpoint
func prepareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req PhoneAuthPrepareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Log the request if debug is enabled
	if cfg.Debug {
		log.Printf("Prepare request: %+v", req)
	}

	// Map use case string to enum
	var useCase glide.UseCase
	switch strings.ToLower(req.UseCase) {
	case "verifyphonenumber", "verify_phone_number":
		useCase = glide.UseCaseVerifyPhoneNumber
	case "getphonenumber", "get_phone_number":
		useCase = glide.UseCaseGetPhoneNumber
	default:
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_USE_CASE",
			fmt.Sprintf("Invalid use case: %s", req.UseCase), nil)
		return
	}

	// Create Glide prepare request
	prepareReq := &glide.PrepareRequest{
		UseCase:     useCase,
		PhoneNumber: req.PhoneNumber,
	}

	// Add PLMN if provided
	if req.PLMN.MCC != "" && req.PLMN.MNC != "" {
		prepareReq.PLMN = &glide.PLMN{
			MCC: req.PLMN.MCC,
			MNC: req.PLMN.MNC,
		}
	}

	// Call Glide API
	ctx := context.Background()
	prepareResp, err := glideClient.MagicAuth.Prepare(ctx, prepareReq)
	if err != nil {
		handleGlideError(w, err)
		return
	}

	// Return response
	response := PhoneAuthPrepareResponse{
		Strategy: string(prepareResp.AuthenticationStrategy),
		Session:  prepareResp.Session,
		Data:     prepareResp.Data,
		TTL:      prepareResp.TTL,
	}

	if cfg.Debug {
		log.Printf("Prepare response: %+v", response)
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Verify phone number endpoint
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req PhoneAuthProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Log the request if debug is enabled
	if cfg.Debug {
		log.Printf("Verify request: session=%s, has_response=%v",
			req.Session, req.Response != nil)
	}

	// Convert response to map if needed
	responseMap, ok := req.Response.(map[string]interface{})
	if !ok {
		// Try to convert
		responseBytes, _ := json.Marshal(req.Response)
		json.Unmarshal(responseBytes, &responseMap)
	}

	// Create Glide verify request
	// Handle both session formats
	var sessionInfo *glide.SessionInfo
	if req.SessionInfo != nil {
		// New format with full session info
		sessionKey, _ := req.SessionInfo["session_key"].(string)
		nonce, _ := req.SessionInfo["nonce"].(string)
		encKey, _ := req.SessionInfo["enc_key"].(string)
		sessionInfo = &glide.SessionInfo{
			SessionKey: sessionKey,
			Metadata: &glide.SessionMetadata{
				Nonce:  nonce,
				EncKey: encKey,
			},
		}
	} else if req.Session != "" {
		sessionInfo = &glide.SessionInfo{
			SessionKey: req.Session,
		}
	}

	verifyReq := &glide.VerifyPhoneNumberRequest{
		Session:    sessionInfo,
		Credential: responseMap,
	}

	// Call Glide API
	ctx := context.Background()
	verifyResp, err := glideClient.MagicAuth.VerifyPhoneNumber(ctx, verifyReq)
	if err != nil {
		handleGlideError(w, err)
		return
	}

	// Return response
	response := PhoneAuthProcessResponse{
		Success:     verifyResp.Verified,
		PhoneNumber: verifyResp.PhoneNumber,
		Verified:    verifyResp.Verified,
	}

	if cfg.Debug {
		log.Printf("Verify response: %+v", response)
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Get phone number endpoint
func getPhoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req PhoneAuthProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Log the request if debug is enabled
	if cfg.Debug {
		log.Printf("Get phone request: session=%s, has_response=%v",
			req.Session, req.Response != nil)
	}

	// Convert response to map if needed
	responseMap, ok := req.Response.(map[string]interface{})
	if !ok {
		// Try to convert
		responseBytes, _ := json.Marshal(req.Response)
		json.Unmarshal(responseBytes, &responseMap)
	}

	// Create Glide get phone request
	// Handle both session formats
	var sessionInfo *glide.SessionInfo
	if req.SessionInfo != nil {
		// New format with full session info
		sessionKey, _ := req.SessionInfo["session_key"].(string)
		nonce, _ := req.SessionInfo["nonce"].(string)
		encKey, _ := req.SessionInfo["enc_key"].(string)
		sessionInfo = &glide.SessionInfo{
			SessionKey: sessionKey,
			Metadata: &glide.SessionMetadata{
				Nonce:  nonce,
				EncKey: encKey,
			},
		}
	} else if req.Session != "" {
		sessionInfo = &glide.SessionInfo{
			SessionKey: req.Session,
		}
	}

	getReq := &glide.GetPhoneNumberRequest{
		Session:    sessionInfo,
		Credential: responseMap,
	}

	// Call Glide API
	ctx := context.Background()
	getResp, err := glideClient.MagicAuth.GetPhoneNumber(ctx, getReq)
	if err != nil {
		handleGlideError(w, err)
		return
	}

	// Return response
	response := PhoneAuthProcessResponse{
		Success:     true,
		PhoneNumber: getResp.PhoneNumber,
		Verified:    false, // GetPhoneNumber doesn't verify
	}

	if cfg.Debug {
		log.Printf("Get phone response: %+v", response)
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Handle Glide SDK errors
func handleGlideError(w http.ResponseWriter, err error) {
	if glideErr, ok := err.(*glide.Error); ok {
		// Map Glide error codes to HTTP status codes
		status := http.StatusInternalServerError
		switch glideErr.Code {
		case glide.ErrCodeCarrierNotEligible:
			status = http.StatusUnprocessableEntity
		case glide.ErrCodeBadRequest, glide.ErrCodeValidationError:
			status = http.StatusBadRequest
		case glide.ErrCodeInternalServerError:
			status = http.StatusUnauthorized
		case glide.ErrCodeRateLimitExceeded:
			status = http.StatusTooManyRequests
		case glide.ErrCodeSessionNotFound:
			status = http.StatusConflict
		}

		// Build details map
		details := make(map[string]interface{})
		if glideErr.Details != nil {
			details = glideErr.Details
		}
		if glideErr.RequestID != "" {
			details["requestId"] = glideErr.RequestID
		}
		if glideErr.Status != 0 {
			details["status"] = glideErr.Status
		}

		sendErrorResponse(w, status, string(glideErr.Code), glideErr.Message, details)
		return
	}

	// Generic error
	sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
}

// Send error response
func sendErrorResponse(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	response := ErrorResponse{
		Error:     code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    status,
		Details:   details,
	}

	// Extract specific fields from details if present
	if details != nil {
		if reqID, ok := details["requestId"].(string); ok {
			response.RequestID = reqID
		}
		if traceID, ok := details["traceId"].(string); ok {
			response.TraceID = traceID
		}
		if spanID, ok := details["spanId"].(string); ok {
			response.SpanID = spanID
		}
	}

	if cfg.Debug {
		log.Printf("Error response: %+v", response)
	}

	sendJSONResponse(w, status, response)
}

// Send JSON response
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Get environment variable with default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
