package utils

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
)

type CSVRecord struct {
	Vendor string
	SKU    string
	Price  float64
}

func ReadCSV(filePath string) ([]CSVRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var csvRecords []CSVRecord
	for i, record := range records {
		// Remove BOM from the first field of the first record
		if i == 0 && len(record[0]) > 0 {
			record[0] = strings.TrimPrefix(record[0], "\ufeff")
		}

		price, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, err
		}
		csvRecords = append(csvRecords, CSVRecord{
			SKU:    record[0],
			Vendor: record[1],
			Price:  price,
		})
	}

	return csvRecords, nil
}
