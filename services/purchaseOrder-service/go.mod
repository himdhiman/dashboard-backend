module github.com/himdhiman/dashboard-backend/services/purchaseOrder-service

go 1.22.2


require (
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/himdhiman/dashboard-backend/libs/logger v0.0.0-20250227135410-83fbb743dc5f
	github.com/himdhiman/dashboard-backend/libs/mongo v0.0.0-20250227135410-83fbb743dc5f
)


replace github.com/himdhiman/dashboard-backend/libs/logger => ../libs/logger

replace github.com/himdhiman/dashboard-backend/libs/mongo => ../libs/mongo
