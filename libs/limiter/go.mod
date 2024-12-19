module github.com/himdhiman/dashboard-backend/libs/limiter

go 1.22.2

require (
	github.com/himdhiman/dashboard-backend/libs/cache v0.0.0-20241218092156-3cd9c315706d
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20241218091630-a9eb4a8a3c99
	github.com/himdhiman/dashboard-backend/libs/mongo v0.0.0-20241218093130-0e369eff53ae
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/himdhiman/dashboard-backend/libs/cache => ../cache
replace github.com/himdhiman/dashboard-backend/libs/logger => ../logger
replace github.com/himdhiman/dashboard-backend/libs/mongo => ../mongo
