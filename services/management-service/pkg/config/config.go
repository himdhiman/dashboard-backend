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

	// Validate required configurations
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks if all required configurations are present
func (c *ProjectConfig) validate() error {
	required := map[string]string{
		"MONGO_USER":     c.MongoUser,
		"MONGO_PASSWORD": c.MongoPassword,
		"MONGO_HOST":     c.MongoHost,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required configuration %s is not set", name)
		} else {
			fmt.Printf("Configuration %s: %s\n", name, value)
		}
	}

	return nil
}
