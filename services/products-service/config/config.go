package config

import "golang.org/x/oauth2/google"

type ProductsServiceConfig struct {
	SpreadsheetID string
	SheetName     string
	Credentials   *google.Credentials
}

func NewProductsServiceConfig(spreadsheetID string, sheetName string, credentials *google.Credentials) *ProductsServiceConfig {
	return &ProductsServiceConfig{
		SpreadsheetID: spreadsheetID,
		SheetName:     sheetName,
		Credentials:   credentials,
	}
}
