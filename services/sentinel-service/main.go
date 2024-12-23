package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)

func main() {

	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig()).WithContext(ctx)

	mongoConfig := mongo.NewMongoConfig("mongodb://localhost:27017", "Dashboard")
	mongoClient, err := mongo.NewMongoClient(mongoConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return
	}

	defer mongoClient.Disconnect(ctx)

	collectionName := "sentinel_apis"
	collection, err := mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	cacheConfig := cache.NewCacheConfig("localhost", 6379, "", 0, 1, "sentinel")
	cache := cache.NewCacheClient(cacheConfig, logger)

	ctx = context.Background()
	err = cache.Ping(ctx)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

	worker.StartConfigSync(collection, cache, logger, 1000)

	secretKey := "Rvdf8NYhzKLpsRrMb7th34bW8bqh4HdT"
	initializationVector := "zUzT1iPfLMw80idf"

	cryptoInstance := crypto.NewCrypto(secretKey, initializationVector)

	authentication := auth.NewAuthentication(cache, logger, cryptoInstance)
	tokenManager := auth.NewTokenManager(cache, logger, cryptoInstance, "unicommerce", authentication)

	payload := map[string]string{
		"skuCode": "NS11716",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Error encoding payload for token request", "error", err)
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://salty.unicommerce.com/services/rest/v1/catalog/itemType/get", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Error creating request for FetchTokens", "error", err)
		panic(err)
	}

	tokenManager.AuthenticateRequest(ctx, req)

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error making request to endpoint", "error", err)
		panic(err)
	}
	defer resp.Body.Close()

	// read the response
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	logger.Info("Response", "response", buf.String())

	// tokens, err := authentication.FetchTokens(ctx, "unicommerce")

	// logger.Info("Tokens", "tokens", tokens)

}
