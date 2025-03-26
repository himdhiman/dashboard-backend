package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type ProjectConfig struct {
	// Database Configurations
	MongoUser     string
	MongoPassword string
	MongoHost     string

	// Crypto Configurations
	SecretKey            string
	InitializationVector string

	// Google Sheets Configurations
	SpreadsheetID         string
	SheetName             string
	GoogleCredentialsPath string

	// Environment
	AppEnv string
}

// LoadConfig loads configuration from environment variables or .env file
func LoadConfig(envFilePath string) (*ProjectConfig, error) {
	// Check if we're in production
	appEnv := os.Getenv("APP_ENV")

	// Load from .env file in development
	if appEnv == "development" {
		if err := godotenv.Load(envFilePath); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config := &ProjectConfig{
		AppEnv: appEnv,
	}

	// Load MongoDB configurations
	config.MongoUser = os.Getenv("MONGO_USER")
	config.MongoPassword = os.Getenv("MONGO_PASSWORD")
	config.MongoHost = os.Getenv("MONGO_HOST")

	// Load Crypto configurations
	config.SecretKey = os.Getenv("SECRET_KEY")
	config.InitializationVector = os.Getenv("INITIALIZATION_VECTOR")

	// Load Google Sheets configurations
	config.SpreadsheetID = os.Getenv("SPREADSHEET_ID")
	config.SheetName = os.Getenv("SHEET_NAME")
	config.GoogleCredentialsPath = os.Getenv("GOOGLE_CREDENTIALS_PATH")

	// Validate required configurations
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks if all required configurations are present
func (c *ProjectConfig) validate() error {
	required := map[string]string{
		"MONGO_USER":              c.MongoUser,
		"MONGO_PASSWORD":          c.MongoPassword,
		"MONGO_HOST":              c.MongoHost,
		"SECRET_KEY":              c.SecretKey,
		"INITIALIZATION_VECTOR":   c.InitializationVector,
		"SPREADSHEET_ID":          c.SpreadsheetID,
		"SHEET_NAME":              c.SheetName,
		"GOOGLE_CREDENTIALS_PATH": c.GoogleCredentialsPath,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required configuration %s is not set", name)
		}
	}

	return nil
}
