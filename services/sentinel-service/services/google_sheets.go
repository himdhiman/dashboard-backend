package services

import (
	"context"
	"errors"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	SpreadsheetID string
	SheetName     string
	Credentials   string
	Logger        logger.ILogger
}

func NewGoogleSheetsService(spreadsheetID, sheetName, credentials string, logger logger.ILogger) *GoogleSheetsService {
	return &GoogleSheetsService{
		SpreadsheetID: spreadsheetID,
		SheetName:     sheetName,
		Credentials:   credentials,
		Logger:        logger,
	}
}

func (s *GoogleSheetsService) UpdateGoogleSheet(ctx context.Context, data []map[string]interface{}) error {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		return errors.New("correlationID not found in context")
	}
	s.Logger.Infof("CorrelationID: %s - Starting UpdateGoogleSheet", correlationID)

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(s.Credentials))
	if err != nil {
		s.Logger.Errorf("CorrelationID: %s - Error creating Google Sheets service: %v", correlationID, err)
		return err
	}

	if len(data) == 0 {
		s.Logger.Infof("CorrelationID: %s - No data provided to update", correlationID)
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

	s.Logger.Infof("CorrelationID: %s - Prepared value range for update", correlationID)

	_, err = srv.Spreadsheets.Values.Update(s.SpreadsheetID, s.SheetName, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		s.Logger.Errorf("CorrelationID: %s - Error updating Google Sheets: %v", correlationID, err)
		return err
	}

	s.Logger.Infof("CorrelationID: %s - Successfully updated Google Sheets", correlationID)
	return nil
}

func (s *GoogleSheetsService) FetchGoogleSheetData(ctx context.Context) ([]map[string]interface{}, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Infof("CorrelationID: %s - Starting FetchGoogleSheetData", correlationID)

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(s.Credentials))
	if err != nil {
		s.Logger.Errorf("CorrelationID: %s - Error creating Google Sheets service: %v", correlationID, err)
		return nil, err
	}

	resp, err := srv.Spreadsheets.Values.Get(s.SpreadsheetID, s.SheetName).Do()
	if err != nil {
		s.Logger.Errorf("CorrelationID: %s - Error fetching data from Google Sheets: %v", correlationID, err)
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

	s.Logger.Infof("CorrelationID: %s - Successfully fetched data from Google Sheets", correlationID)
	return data, nil
}
