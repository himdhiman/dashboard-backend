module github.com/himdhiman/dashboard-backend/services/shipping-service

go 1.23.0

toolchain go1.23.7

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/himdhiman/dashboard-backend/libs/constants v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/mappers v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/mongo v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/task v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/services/products-service v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/services/purchaseOrder-service v0.0.0-20250227135410-83fbb743dc5f

)

require (
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator v9.31.0+incompatible // indirect
	github.com/go-playground/validator/v10 v10.25.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/himdhiman/dashboard-backend/libs/cache v0.0.0-20250227135410-83fbb743dc5f // indirect
	github.com/himdhiman/dashboard-backend/libs/conflux v0.0.0-20250227135410-83fbb743dc5f // indirect
	github.com/himdhiman/dashboard-backend/libs/crypto v0.0.0-20250227135410-83fbb743dc5f // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/arch v0.15.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/api v0.227.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/himdhiman/dashboard-backend/libs/logger => ../../libs/logger

replace github.com/himdhiman/dashboard-backend/libs/mongo => ../../libs/mongo

replace github.com/himdhiman/dashboard-backend/libs/constants => ../../libs/constants

replace github.com/himdhiman/dashboard-backend/libs/conflux => ../../libs/conflux

replace github.com/himdhiman/dashboard-backend/libs/mappers => ../../libs/mappers

replace github.com/himdhiman/dashboard-backend/libs/task => ../../libs/task

replace github.com/himdhiman/dashboard-backend/services/purchaseOrder-service => ../purchaseOrder-service

replace github.com/himdhiman/dashboard-backend/services/products-service => ../products-service

replace golang.org/x/net => golang.org/x/net v0.7.0
