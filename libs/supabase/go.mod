module github.com/himdhiman/dashboard-backend/libs/supabase

go 1.23.0

toolchain go1.23.7

require (
	github.com/gorilla/websocket v1.5.0
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20241218052858-2f8483cbcb4a
	github.com/stretchr/testify v1.8.4
)

replace github.com/himdhiman/dashboard-backend/libs/logger => ../logger

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
