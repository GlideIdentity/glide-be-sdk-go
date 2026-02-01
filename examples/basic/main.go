// Example: Backend server using the Glide SDK
//
// Architecture:
//   Frontend App (glide-fe-sdk) → Your Backend (this) → Glide Services
//
// This example shows a minimal HTTP server that handles phone authentication
// requests from your frontend and proxies them to Glide's backend services.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	glide "github.com/GlideIdentity/glide-be-sdk-go/v2"
)

var client *glide.Client

func main() {
	client = glide.New(
		glide.WithClientCredentials(
			os.Getenv("GLIDE_CLIENT_ID"),
			os.Getenv("GLIDE_CLIENT_SECRET"),
		),
		glide.WithTimeout(30*time.Second),
	)

	http.HandleFunc("/api/phone-auth/prepare", handlePrepare)
	http.HandleFunc("/api/phone-auth/process", handleProcess)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handlePrepare receives prepare requests from frontend and forwards to Glide
func handlePrepare(w http.ResponseWriter, r *http.Request) {
	var req glide.PrepareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := client.MagicalAuth.Prepare(r.Context(), &req)
	if err != nil {
		handleGlideError(w, err)
		return
	}

	writeJSON(w, resp)
}

// handleProcess receives credentials from frontend and processes with Glide
func handleProcess(w http.ResponseWriter, r *http.Request) {
	var req glide.ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	switch req.UseCase {
	case glide.UseCaseGetPhoneNumber:
		resp, err := client.MagicalAuth.GetPhoneNumber(r.Context(), &glide.GetPhoneNumberRequest{
			Session:    req.Session,
			Credential: req.Credential,
		})
		if err != nil {
			handleGlideError(w, err)
			return
		}
		writeJSON(w, resp)

	case glide.UseCaseVerifyPhoneNumber:
		resp, err := client.MagicalAuth.VerifyPhoneNumber(r.Context(), &glide.VerifyPhoneNumberRequest{
			Session:    req.Session,
			Credential: req.Credential,
		})
		if err != nil {
			handleGlideError(w, err)
			return
		}
		writeJSON(w, resp)

	default:
		writeError(w, http.StatusBadRequest, "Invalid use_case")
	}
}

func handleGlideError(w http.ResponseWriter, err error) {
	if glideErr, ok := err.(*glide.MagicalAuthError); ok {
		status := glideErr.Status
		if status == 0 {
			status = http.StatusUnprocessableEntity
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(glideErr)
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
