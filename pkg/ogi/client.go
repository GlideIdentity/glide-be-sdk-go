package ogi

import (
	"errors"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
)

type GlideClient struct {
	clientId     string
	clientSecret string
	redirectUri  string
	session 	 *Session
}

func init() {
	logLevel := log.ErrorLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")
	if logLevelSet {
		switch logLevelEnv {
		case "debug":
			logLevel = log.DebugLevel
		case "info":
			logLevel = log.InfoLevel
		case "warn":
			logLevel = log.WarnLevel
		case "error":
			logLevel = log.ErrorLevel
		default:
			logLevel = log.ErrorLevel
		}
	}

	log.SetLevel(logLevel)
}


func NewGlideClient() (*GlideClient, error) {
	// parse client id, client secret and base url from environment variables
	env, err := ReadEnv()
	if err != nil {
		return nil, errors.New("failed to read environment variables: " + err.Error())
	}

	// validate base url
	if _, err := url.ParseRequestURI(env.RedirectURI); err != nil {
		return nil, errors.New("invalid base url: " + env.RedirectURI)
	}

	return &GlideClient{
		clientId:     env.ClientID,
		clientSecret: env.ClientSecret,
		redirectUri:      env.RedirectURI,
	}, nil
}

