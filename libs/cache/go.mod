module github.com/himdhiman/dashboard-backend/libs/cache

go 1.22.2

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20241218052858-2f8483cbcb4a
)

replace github.com/himdhiman/dashboard-backend/libs/logger => ../logger

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
