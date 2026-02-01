// Example: Backend server using only core types
//
// Architecture:
//   Frontend App (glide-fe-sdk) → Your Backend (this) → Glide Services
//
// This example shows how to use core types with your own HTTP client.
// Use this when you need minimal dependencies or custom HTTP handling.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/GlideIdentity/glide-be-sdk-go/core/v2"
)

var (
	clientID     = os.Getenv("GLIDE_CLIENT_ID")
	clientSecret = os.Getenv("GLIDE_CLIENT_SECRET")
	baseURL      = "https://api.glideidentity.app"

	// Token cache
	cachedToken     string
	cachedExpiresAt time.Time
	tokenMutex      sync.Mutex
)

func main() {
	http.HandleFunc("/api/phone-auth/prepare", handlePrepare)
	http.HandleFunc("/api/phone-auth/process", handleProcess)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handlePrepare receives prepare requests from frontend and forwards to Glide
func handlePrepare(w http.ResponseWriter, r *http.Request) {
	var req core.PrepareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := callGlideAPI("/magic-auth/v2/auth/prepare", &req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// handleProcess receives credentials from frontend and processes with Glide
func handleProcess(w http.ResponseWriter, r *http.Request) {
	var req core.ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var path string
	var glideReq interface{}

	switch req.UseCase {
	case core.UseCaseGetPhoneNumber:
		path = "/magic-auth/v2/auth/get-phone-number"
		glideReq = &core.GetPhoneNumberRequest{
			Session:    req.Session,
			Credential: req.Credential,
		}
	case core.UseCaseVerifyPhoneNumber:
		path = "/magic-auth/v2/auth/verify-phone-number"
		glideReq = &core.VerifyPhoneNumberRequest{
			Session:    req.Session,
			Credential: req.Credential,
		}
	default:
		writeError(w, http.StatusBadRequest, "Invalid use_case")
		return
	}

	resp, err := callGlideAPI(path, glideReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// getAccessToken retrieves or refreshes OAuth2 access token
func getAccessToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	// Check if we have a valid cached token (with 60 second buffer)
	if cachedToken != "" && time.Now().Before(cachedExpiresAt.Add(-60*time.Second)) {
		return cachedToken, nil
	}

	// Fetch new token
	tokenURL := baseURL + "/oauth2-cc/token"
	credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+credentials)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	cachedToken = tokenResp.AccessToken
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	cachedExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return cachedToken, nil
}

// callGlideAPI sends a request to Glide's API with OAuth2 authentication
func callGlideAPI(path string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Get OAuth2 access token
	accessToken, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "glide-be-sdk-go/"+core.GetVersion())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
