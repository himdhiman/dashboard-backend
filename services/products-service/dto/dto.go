package dto

type ProductPayloadDTO struct {
	SheetName string      `json:"sheetName"` // Name of the sheet
	RowNumber int         `json:"rowNumber"` // Row number in the sheet
	Data      ProductData `json:"data"`      // Data object containing product details
}

type ProductData struct {
	// LotNumber             string `json:"Lot Number"`                      // Lot number
	// PONumber              string `json:"PO Number"`                       // Purchase order number
	// Channel               string `json:"Channel"`                         // Sales channel
	SKU         string `json:"SKU"`          // SKU identifier
	Quantity    int    `json:"Quantity"`     // Quantity of the product
	ShelfNumber string `json:"Shelf Number"` // Shelf number
	// InvoiceNumber         string `json:"Invoice Number"`                  // Invoice number
	// InvoiceDate           string `json:"Invoice Date"`                    // Invoice date
	// Remarks               string `json:"Remarks"`                         // Remarks or comments
	// Date                  string `json:"Date"`                            // General date field
	// Status                string `json:"Status"`                          // Status of the product
	// InventoryAdjustedInUC string `json:"Inventory adjusted in UC or Not"` // Inventory adjustment status
}
