package resultsspreadsheet

import (
	"fmt"

	"google.golang.org/api/sheets/v4"
)

func getHeadersAsInterfaceList(sheet *sheets.Sheet) []interface{} {
	headersValues := sheet.Data[0].RowData[0].Values
	headers := make([]interface{}, len(headersValues))
	for i, cell := range headersValues {
		headers[i] = *cell.UserEnteredValue.StringValue
	}
	return headers
}

func getIndicesFromListByName(headers []interface{}, wantedColumnsNames []string) ([]int, error) {
	indices := []int{}
	for _, wantedColName := range wantedColumnsNames {
		found := false
		for i, val := range headers {
			if wantedColName == val {
				found = true
				indices = append(indices, i)
				continue
			}
		}
		if !found {
			return nil, fmt.Errorf("column %s doesn't exist in given headers list", wantedColName)
		}
	}
	return indices, nil
}

func getSheetIDByName(spreadsheet *sheets.Spreadsheet, wantedSheetName string) (int64, error) {
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == wantedSheetName {
			return sheet.Properties.SheetId, nil
		}
	}
	return -1, fmt.Errorf("there is no sheet named %s in spreadsheet %s", wantedSheetName, spreadsheet.SpreadsheetUrl)
}

func addBasicFilterToSpreadSheet(srv *sheets.Service, spreadsheet *sheets.Spreadsheet) error {
	requests := []*sheets.Request{}
	for _, sheet := range spreadsheet.Sheets {
		requests = append(requests, &sheets.Request{
			SetBasicFilter: &sheets.SetBasicFilterRequest{
				Filter: &sheets.BasicFilter{
					Range: &sheets.GridRange{SheetId: sheet.Properties.SheetId},
				},
			},
		})
	}

	_, err := srv.Spreadsheets.BatchUpdate(spreadsheet.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return err
	}
	return nil
}

func addDescendingSortFilterToSheet(srv *sheets.Service, spreadsheet *sheets.Spreadsheet, sheetName, colName string) error {
	sheetsValues, err := srv.Spreadsheets.Values.Get(spreadsheet.SpreadsheetId, sheetName).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s values: %v", sheetName, err)
	}
	indices, err := getIndicesFromListByName(sheetsValues.Values[0], []string{colName})
	if err != nil {
		return nil
	}

	sheetID, err := getSheetIDByName(spreadsheet, sheetName)
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s id: %v", sheetName, err)
	}

	requests := []*sheets.Request{
		{
			SortRange: &sheets.SortRangeRequest{
				Range: &sheets.GridRange{
					SheetId:       sheetID,
					StartRowIndex: 1,
				},
				SortSpecs: []*sheets.SortSpec{
					{
						DimensionIndex: int64(indices[0]),
						SortOrder:      "DESCENDING",
					},
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheet.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return err
	}
	return nil
}

func addFilterByFailedAndMandatoryToSheet(srv *sheets.Service, spreadsheet *sheets.Spreadsheet, sheetName string) error {
	sheetsValues, err := srv.Spreadsheets.Values.Get(spreadsheet.SpreadsheetId, sheetName).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s values: %v", sheetName, err)
	}
	indices, err := getIndicesFromListByName(sheetsValues.Values[0], []string{"State", "Mandatory/Optional"})
	if err != nil {
		return nil
	}

	stateColIndex := indices[0]
	isMandatoryColIndex := indices[1]

	sheetID, err := getSheetIDByName(spreadsheet, sheetName)
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s id: %v", sheetName, err)
	}

	requests := []*sheets.Request{
		{
			SetBasicFilter: &sheets.SetBasicFilterRequest{
				Filter: &sheets.BasicFilter{
					Range: &sheets.GridRange{SheetId: sheetID},
					Criteria: map[string]sheets.FilterCriteria{
						fmt.Sprint(stateColIndex): {
							Condition: &sheets.BooleanCondition{
								Type: "TEXT_EQ",
								Values: []*sheets.ConditionValue{
									{UserEnteredValue: "failed"},
								},
							},
						},
						fmt.Sprint(isMandatoryColIndex): {
							Condition: &sheets.BooleanCondition{
								Type: "TEXT_EQ",
								Values: []*sheets.ConditionValue{
									{UserEnteredValue: "Mandatory"},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheet.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do()
	if err != nil {
		return err
	}

	return nil
}
