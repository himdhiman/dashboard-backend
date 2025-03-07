package products

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/services/products-service/config"
	"github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
	"github.com/himdhiman/dashboard-backend/services/products-service/routes"
	"github.com/himdhiman/dashboard-backend/services/products-service/schedulers"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"

	cache "github.com/himdhiman/dashboard-backend/libs/cache"
	conflux "github.com/himdhiman/dashboard-backend/libs/conflux/cmd"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
)

func InitializeProductsService(router *gin.Engine, ctx context.Context, config *config.ProductsServiceConfig, logger logger.ILogger, cache cache.Cacher, confluxService *conflux.ConfluxService, mongoClient mongo.IMongoClient) *gin.Engine {

	collection, err := mongoClient.GetCollection(context.Background(), "unicom_products")
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}
	productsRepository := repository.Repository[models.Product]{Collection: collection}

	unicommerceApiClient, err := confluxService.CreateApiClient(constants.UNICOM_API_CODE, conflux.AuthStrategyBasic)
	if err != nil {
		logger.Fatal("Failed to create Unicommerce API client", "error", err)
	}

	googleSheetService := services.NewGoogleSheetsService(config.SpreadsheetID, config.SheetName, config.Credentials, logger)
	unicommerceProductsService := services.NewUnicommerceProductsService(logger, cache, unicommerceApiClient, &productsRepository)
	productsService := services.NewProductsService(unicommerceApiClient, googleSheetService, logger, cache, collection)

	productsServices := routes.ProductsServices{
		UnicommerceProductsService: unicommerceProductsService,
		ProductsService:            productsService,
	}

	schedulerCollection, err := mongoClient.GetCollection(context.Background(), "product_schedulers")
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	exportJobScheduler := schedulers.NewExportJobScheduler(schedulerCollection, unicommerceProductsService, logger)
	exportJobScheduler.Start(ctx)

	// start invetory snapshot scheduler
	inventorySnapShotScheduler := schedulers.NewInventorySnapShotScheduler(schedulerCollection, productsService, logger)
	inventorySnapShotScheduler.Start(ctx)

	router = routes.AddProductsRoutes(router, logger, &productsServices)

	return router
}
