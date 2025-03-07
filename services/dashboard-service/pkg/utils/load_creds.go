package utils

import (
	"context"
	"os"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"golang.org/x/oauth2/google"
)

func LoadGoogleCreds(ctx context.Context, path string, logger logger.ILogger) (*google.Credentials, error) {
	credBytes, err := os.ReadFile(path)
	if err != nil {
		logger.Fatal("Failed to read credentials file", "error", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		logger.Fatal("Failed to parse credentials", "error", err)
	}

	return creds, nil
}
