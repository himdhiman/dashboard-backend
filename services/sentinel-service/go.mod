module github.com/himdhiman/dashboard-backend/services/sentinel-service

go 1.22.2

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/himdhiman/dashboard-backend/libs/cache v0.0.0-20241218093311-5bed961e82ae
	github.com/himdhiman/dashboard-backend/libs/crypto v0.0.0-20241220153702-23a782a7858d
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20241218052858-2f8483cbcb4a
	github.com/himdhiman/dashboard-backend/libs/mongo v0.0.0-20241218093311-5bed961e82ae
	github.com/himdhiman/dashboard-backend/libs/scheduler v0.0.0-20241218052858-2f8483cbcb4a
	github.com/himdhiman/dashboard-backend/libs/task v0.0.0-20241218093311-5bed961e82ae
	github.com/joho/godotenv v1.5.1
)

replace github.com/himdhiman/dashboard-backend/libs/logger => ../../libs/logger

replace github.com/himdhiman/dashboard-backend/libs/mongo => ../../libs/mongo

replace github.com/himdhiman/dashboard-backend/libs/cache => ../../libs/cache

replace github.com/himdhiman/dashboard-backend/libs/crypto => ../../libs/crypto

replace github.com/himdhiman/dashboard-backend/libs/task => ../../libs/task

replace github.com/himdhiman/dashboard-backend/libs/scheduler => ../../libs/scheduler

require (
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.1 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
