package ogi

import (
	"math/rand"
	"os/exec"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// EnvConfig represents the environment configuration for Glide.
type EnvConfig struct {
	RedirectURI string `env:"GLIDE_REDIRECT_URI,required"`
	ClientID    string `env:"GLIDE_CLIENT_ID,required"`
	ClientSecret string `env:"GLIDE_CLIENT_SECRET,required"`
	InternalAuthBaseUrl string `env:"GLIDE_AUTH_BASE_URL" envDefault:"https://oidc.gateway-x.io"`
	InternalApiBaseUrl string `env:"GLIDE_API_BASE_URL" envDefault:"https://api.gateway-x.io"`
}

var envConfig *EnvConfig

// ReadEnv reads the .env file from the root directory of the current git repository.
// It returns an EnvConfig struct containing the required environment variables.
// If any of the required variables are missing, it returns an error.
func ReadEnv() (*EnvConfig, error) {
	if envConfig != nil {
		return envConfig, nil
	}

	rootDir, err := FindGitRepoDir()
	if err != nil {
		return nil, err
	}

	err = godotenv.Load(rootDir + "/.env")
	if err != nil {
		return nil, err
	}

	config := &EnvConfig{}
	err = env.Parse(config)
	if err != nil {
		return nil, err
	}

	envConfig = config
	return config, nil
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FindGitRepoDir() (string, error) {
	// check if the current directory is a git repo
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("Failed to find git repo: %+v", err)
		return "", err
	}

	// remove the newline character
	filePath := strings.Trim(string(out), "\n")
	return filePath, nil
}

func FormatPhoneNumber(phoneNumber string) string {
    phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
    if !strings.HasPrefix(phoneNumber, "+") {
        phoneNumber = "+" + phoneNumber
    }
    return phoneNumber
}
