package services

import (
	"context"
	"strings"
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

func BundlesExportProcessor(ctx context.Context, s *UnicommerceProductsService, records [][]string) error {
	for i, record := range records {
		skuCode := record[0]
		products := record[1]

		// Skip header row
		if i == 0 {
			continue
		}

		// Ignore independent products (not bundles)
		if products == "" {
			continue
		}

		// Split comma-separated SKUs and trim spaces
        productList := []string{}
        for _, p := range strings.Split(products, ",") {
            p = strings.TrimSpace(p)
            if p != "" {
                productList = append(productList, p)
            }
        }

		bundles, err := s.ProductsBundlesRepository.Find(ctx, map[string]interface{}{"bundleSku": skuCode})
		if err != nil {
			return err
		}
		if len(bundles) > 0 {
			_, err = s.ProductsBundlesRepository.Update(ctx, map[string]interface{}{"products": productList, "updatedAt": time.Now()}, bundles[0])
			if err != nil {
				return err
			}
			continue
		}
		bundle := models.ProductBundle{
			BundleSKU: skuCode,
			Products:  productList,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err = s.ProductsBundlesRepository.Create(ctx, &bundle)
		if err != nil {
			return err
		}
	}
	return nil
}
