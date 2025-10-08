package products

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/services/products-service/config"
	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
	"github.com/himdhiman/dashboard-backend/services/products-service/routes"
	"github.com/himdhiman/dashboard-backend/services/products-service/schedulers"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"

	cache "github.com/himdhiman/dashboard-backend/libs/cache"
	conflux "github.com/himdhiman/dashboard-backend/libs/conflux/cmd"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
)

type ProductsServices struct {
	UnicommerceProductsService *services.UnicommerceProductsService
	ProductsService            *services.ProductsService
	GoogleSheetService         *services.GoogleSheetsService
}

func InitializeProductsService(router *gin.Engine, ctx context.Context, config *config.ProductsServiceConfig, logger logger.ILogger, cache cache.Cacher, confluxService *conflux.ConfluxService, mongoClient mongo.IMongoClient) (*ProductsServices, error) {
	// Initialize repositories, services, controllers, and routes here
	collection, err := mongoClient.GetCollection(context.Background(), constants.UnicommerceProductsCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}
	productsRepository := repository.Repository[models.Product]{Collection: collection}

	unicommerceApiClient, err := confluxService.CreateApiClient(products_constants.UNICOM_API_CODE, conflux.AuthStrategyBasic)
	if err != nil {
		logger.Fatal("Failed to create Unicommerce API client", "error", err)
		return nil, err
	}

	productsBundleCollection, err := mongoClient.GetCollection(context.Background(), constants.ProductsBundlesCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}
	productsBundlesRepository := repository.Repository[models.ProductBundle]{Collection: productsBundleCollection}

	googleSheetService := services.NewGoogleSheetsService(config.SpreadsheetID, config.SheetName, config.Credentials, logger)
	unicommerceProductsService := services.NewUnicommerceProductsService(logger, cache, unicommerceApiClient, &productsRepository, &productsBundlesRepository)
	productsService := services.NewProductsService(unicommerceApiClient, googleSheetService, logger, cache, collection)

	productsServices := routes.ProductsServices{
		UnicommerceProductsService: unicommerceProductsService,
		ProductsService:            productsService,
	}

	schedulerCollection, err := mongoClient.GetCollection(context.Background(), constants.ProductsSchedulersCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	jobCodes := []string{
		products_constants.PRODUCTS_EXPORT_JOB_CODE,
		products_constants.BUNDLES_EXPORT_JOB_CODE,
		// Add more codes if needed
	}
	cronExpr := "0 */5 * * * *" // Every 5 minutes

	exportJobScheduler := schedulers.NewExportJobScheduler(schedulerCollection, unicommerceProductsService, logger, jobCodes, cronExpr)
	exportJobScheduler.Start(ctx)

	// start inventory snapshot scheduler
	inventorySnapShotScheduler := schedulers.NewInventorySnapShotScheduler(schedulerCollection, productsService, logger)
	inventorySnapShotScheduler.Start(ctx)

	routes.AddProductsRoutes(router, logger, &productsServices)

	return &ProductsServices{
		UnicommerceProductsService: unicommerceProductsService,
		ProductsService:            productsService,
		GoogleSheetService:         googleSheetService,
	}, nil
}
