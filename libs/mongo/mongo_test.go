package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

func TestNewMongoDB(t *testing.T) {
	// Test for successful connection
	t.Run("Valid Connection", func(t *testing.T) {
		mongoURI := "mongodb://localhost:27017"
		dbName := "testDB"
		loggerInstance := logger.New(logger.DefaultConfig())

		mongoClient, err := NewMongoDB(mongoURI, dbName, loggerInstance)
		assert.NoError(t, err)
		assert.NotNil(t, mongoClient)
		mongoClient.Close()
	})

	// Test for invalid connection
	t.Run("Invalid Connection", func(t *testing.T) {
		mongoURI := "mongodb://invalid:27017"
		dbName := "testDB"
		loggerInstance := logger.New(logger.DefaultConfig())

		_, err := NewMongoDB(mongoURI, dbName, loggerInstance)
		assert.Error(t, err)
	})
}
