package resultsspreadsheet

import (
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// GetHeadersFromSheet extracts header names from a Google Sheets worksheet.
//
// It reads the first row of the provided Sheet and returns each cell value as
// a string slice. The returned slice preserves the order of columns as they
// appear in the sheet. If the sheet has no rows, an empty slice is returned.
func GetHeadersFromSheet(sheet *sheets.Sheet) []string {
	headers := []string{}
	for _, val := range sheet.Data[0].RowData[0].Values {
		headers = append(headers, *val.UserEnteredValue.StringValue)
	}
	return headers
}

// GetHeadersFromValueRange extracts header names from a Google Sheets ValueRange.
//
// It expects the first row of the provided ValueRange to contain column titles.
// The function iterates over each cell in that first row, converts the cell value to
// a string using fmt.Sprint, and returns a slice of these strings. If the ValueRange
// is nil or has no rows, an empty slice is returned. This header slice can be used
// to map column indices to meaningful names when processing spreadsheet data.
func GetHeadersFromValueRange(sheetsValues *sheets.ValueRange) []string {
	headers := []string{}
	for _, val := range sheetsValues.Values[0] {
		headers = append(headers, fmt.Sprint(val))
	}
	return headers
}

// GetHeaderIndicesByColumnNames returns the column indices for a list of header names within a sheet row.
//
// It scans the provided slice of header strings and records the index of each
// requested column name in order. If any requested name is not found, it
// returns an error describing which columns were missing. The function
// outputs a slice of integers corresponding to the positions of the found
// headers and may return nil for the indices slice if an error occurs.
func GetHeaderIndicesByColumnNames(headers, names []string) ([]int, error) {
	indices := []int{}
	for _, name := range names {
		found := false
		for i, val := range headers {
			if name == val {
				found = true
				indices = append(indices, i)
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %s doesn't exist in given headers list", name)
		}
	}
	return indices, nil
}

// GetSheetIDByName retrieves the ID of a sheet in a spreadsheet by its name.
//
// It takes a pointer to a Spreadsheet object and the target sheet name,
// then searches the spreadsheet's sheets for a match.
// If found, it returns the sheet's ID as an int64.
// If no matching sheet exists or an error occurs during lookup, it returns an error.
func GetSheetIDByName(spreadsheet *sheets.Spreadsheet, name string) (int64, error) {
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == name {
			return sheet.Properties.SheetId, nil
		}
	}
	return -1, fmt.Errorf("there is no sheet named %s in spreadsheet %s", name, spreadsheet.SpreadsheetUrl)
}

// addBasicFilterToSpreadSheet applies a basic filter to the first sheet of a Google Sheets spreadsheet.
//
// It takes a sheets.Service client and a pointer to a Spreadsheet object.
// The function constructs a request that adds an AutoFilter covering all columns
// in the first sheet, enabling quick filtering by column headers.
// It sends this batch update via the service's BatchUpdate method.
// If the operation fails, it returns the encountered error.
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

// addDescendingSortFilterToSheet adds a descending sort filter to the specified sheet in a Google Sheets spreadsheet.
//
// It retrieves the sheet ID by name, obtains header indices for the provided column names,
// constructs a sort request that orders rows in descending order based on those columns,
// and applies the filter using a batch update. The function returns an error if any
// step fails or if the required headers cannot be found.
func addDescendingSortFilterToSheet(srv *sheets.Service, spreadsheet *sheets.Spreadsheet, sheetName, colName string) error {
	sheetsValues, err := srv.Spreadsheets.Values.Get(spreadsheet.SpreadsheetId, sheetName).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s values: %v", sheetName, err)
	}
	headers := GetHeadersFromValueRange(sheetsValues)
	indices, err := GetHeaderIndicesByColumnNames(headers, []string{colName})
	if err != nil {
		return nil
	}

	sheetID, err := GetSheetIDByName(spreadsheet, sheetName)
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

// addFilterByFailedAndMandatoryToSheet adds a filter to the specified sheet that shows only rows where the test has failed or is mandatory.
//
// It receives a Google Sheets service, the spreadsheet object and the name of the sheet.
// The function retrieves the sheet ID by name, obtains the header indices for the columns
// that indicate failure status and mandatory flag, and then builds a filter view
// that includes only those rows where either column is true. The filter is applied
// via a batch update request to the spreadsheet API. Any errors encountered during
// retrieval of sheet metadata or during the update are returned.
func addFilterByFailedAndMandatoryToSheet(srv *sheets.Service, spreadsheet *sheets.Spreadsheet, sheetName string) error {
	sheetsValues, err := srv.Spreadsheets.Values.Get(spreadsheet.SpreadsheetId, sheetName).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve sheet %s values: %v", sheetName, err)
	}
	headers := GetHeadersFromValueRange(sheetsValues)
	indices, err := GetHeaderIndicesByColumnNames(headers, []string{"State", "Mandatory/Optional"})
	if err != nil {
		return nil
	}

	stateColIndex := indices[0]
	isMandatoryColIndex := indices[1]

	sheetID, err := GetSheetIDByName(spreadsheet, sheetName)
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
	return err
}
