package services

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/services/products-service/models"
)

// FileProcessorFunc defines the processor signature
type FileProcessorFunc func(ctx context.Context, s *UnicommerceProductsService, records [][]string) error

// Example processor for products export job
func ProductsExportProcessor(ctx context.Context, s *UnicommerceProductsService, records [][]string) error {
	for _, record := range records {
		if record[3] != "SIMPLE" {
			continue
		}
		skuCode := record[0]
		name := record[1]
		imageURL := record[2]
		primaryVendor := record[5]

		products, err := s.ProductsRepository.Find(ctx, map[string]interface{}{"skuCode": skuCode, "primaryVendor": primaryVendor})
		if err != nil {
			return err
		}
		if len(products) > 0 {
			_, err = s.ProductsRepository.Update(ctx, map[string]interface{}{"name": name, "imageUrl": imageURL, "updatedAt": time.Now()}, products[0])
			if err != nil {
				return err
			}
			continue
		}
		product := models.Product{
			SKUCode:       skuCode,
			Name:          name,
			ImageURL:      imageURL,
			PrimaryVendor: primaryVendor,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err = s.ProductsRepository.Create(ctx, &product)
		if err != nil {
			return err
		}
	}
	return nil
}
