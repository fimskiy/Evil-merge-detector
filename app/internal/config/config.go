package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	AppID              int64
	PrivateKey         []byte
	WebhookSecret      []byte
	DatabaseURL        string
	OAuthClientID      string
	OAuthClientSecret  string
	SessionSecret      []byte
}

func Load() (*Config, error) {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID is required")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_APP_ID: %w", err)
	}

	privateKey := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKey == "" {
		// Try reading from file path
		path := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
		if path == "" {
			return nil, fmt.Errorf("GITHUB_PRIVATE_KEY or GITHUB_PRIVATE_KEY_PATH is required")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading private key: %w", err)
		}
		privateKey = string(data)
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}

	return &Config{
		Port:              port,
		AppID:             appID,
		PrivateKey:        []byte(privateKey),
		WebhookSecret:     []byte(webhookSecret),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		OAuthClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		OAuthClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		SessionSecret:     []byte(sessionSecret),
	}, nil
}
