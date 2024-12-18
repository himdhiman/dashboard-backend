package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMongoDB(t *testing.T) {
	// Test for successful connection
	t.Run("Valid Connection", func(t *testing.T) {
		mongoURI := "mongodb://localhost:27017"
		dbName := "testDB"

		mongoClient, err := NewMongoDB(mongoURI, dbName)
		assert.NoError(t, err)
		assert.NotNil(t, mongoClient)
		mongoClient.Close()
	})

	// Test for invalid connection
	t.Run("Invalid Connection", func(t *testing.T) {
		mongoURI := "mongodb://invalid:27017"
		dbName := "testDB"

		_, err := NewMongoDB(mongoURI, dbName)
		assert.Error(t, err)
	})
}
