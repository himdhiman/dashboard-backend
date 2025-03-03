module github.com/himdhiman/dashboard-backend/libs/conflux

go 1.22.2

require (
	github.com/himdhiman/dashboard-backend/libs/cache v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/crypto v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/mongo v0.0.0-20250227135410-83fbb743dc5f
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
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

replace github.com/himdhiman/dashboard-backend/libs/logger => ../../libs/logger

replace github.com/himdhiman/dashboard-backend/libs/cache => ../../libs/cache

replace github.com/himdhiman/dashboard-backend/libs/crypto => ../../libs/crypto

replace github.com/himdhiman/dashboard-backend/libs/mongo => ../../libs/mongo
