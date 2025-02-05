package services

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	SpreadsheetID string
	SheetName     string
	Credentials   string
}

func NewGoogleSheetsService(spreadsheetID, sheetName, credentials string) *GoogleSheetsService {
	return &GoogleSheetsService{
		SpreadsheetID: spreadsheetID,
		SheetName:     sheetName,
		Credentials:   credentials,
	}
}

func (s *GoogleSheetsService) UpdateGoogleSheet(ctx context.Context, data []map[string]interface{}) error {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(s.Credentials))
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	// Extract headers from the first record and store their positions
	headers := make([]interface{}, 0, len(data[0]))
	headerPositions := make(map[string]int)
	for key := range data[0] {
		headers = append(headers, key)
		headerPositions[key] = len(headers) - 1
	}

	// Prepare the value range
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, headers)

	for _, record := range data {
		row := make([]interface{}, len(headers))
		for key, value := range record {
			if pos, ok := headerPositions[key]; ok {
				row[pos] = value
			}
		}
		vr.Values = append(vr.Values, row)
	}

	_, err = srv.Spreadsheets.Values.Update(s.SpreadsheetID, s.SheetName, &vr).ValueInputOption("RAW").Do()
	return err
}

func (s *GoogleSheetsService) FetchGoogleSheetData(ctx context.Context) ([]map[string]interface{}, error) {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(s.Credentials))
	if err != nil {
		return nil, err
	}

	resp, err := srv.Spreadsheets.Values.Get(s.SpreadsheetID, s.SheetName).Do()
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	headers := resp.Values[0]
	for _, row := range resp.Values[1:] {
		record := make(map[string]interface{})
		for i, cell := range row {
			record[headers[i].(string)] = cell
		}
		data = append(data, record)
	}

	return data, nil
}
