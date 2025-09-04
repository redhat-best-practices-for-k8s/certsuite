package resultsspreadsheet

import (
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// GetHeadersFromSheet Retrieves header names from a spreadsheet sheet
//
// The function accesses the first row of the provided sheet, extracts each
// cell's string value, and returns them as a slice of strings. It assumes that
// the sheet contains at least one row with headers. The returned slice
// preserves the order of columns as they appear in the sheet.
func GetHeadersFromSheet(sheet *sheets.Sheet) []string {
	headers := []string{}
	for _, val := range sheet.Data[0].RowData[0].Values {
		headers = append(headers, *val.UserEnteredValue.StringValue)
	}
	return headers
}

// GetHeadersFromValueRange extracts header names from the first row of a spreadsheet
//
// The function receives a ValueRange object containing cell values, accesses
// its first row, and converts each entry to a string using formatting logic. It
// collects these strings into a slice that represents column headers for later
// lookup operations. The returned slice is used by other utilities to map
// header names to column indices.
func GetHeadersFromValueRange(sheetsValues *sheets.ValueRange) []string {
	headers := []string{}
	for _, val := range sheetsValues.Values[0] {
		headers = append(headers, fmt.Sprint(val))
	}
	return headers
}

// GetHeaderIndicesByColumnNames Finds header positions for specified column names
//
// The function scans a slice of header strings to locate the index of each
// requested column name. It returns an integer slice containing the indices in
// the same order as the input names or an error if any name is missing from the
// headers. The returned indices can be used to reference columns when
// manipulating spreadsheet data.
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

// GetSheetIDByName Retrieves a sheet's numeric identifier by its title
//
// This function scans the list of sheets in a spreadsheet for one whose title
// matches the provided name. If found, it returns that sheet's unique ID and no
// error; otherwise it returns -1 and an error describing the missing sheet.
func GetSheetIDByName(spreadsheet *sheets.Spreadsheet, name string) (int64, error) {
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == name {
			return sheet.Properties.SheetId, nil
		}
	}
	return -1, fmt.Errorf("there is no sheet named %s in spreadsheet %s", name, spreadsheet.SpreadsheetUrl)
}

// addBasicFilterToSpreadSheet Adds a basic filter to every sheet in the spreadsheet
//
// The function iterates over each sheet in the provided spreadsheet, creating a
// request that sets a basic filter covering the entire sheet range. It then
// sends all requests as a batch update to the Google Sheets API. If the update
// succeeds it returns nil; otherwise it propagates the error.
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

// addDescendingSortFilterToSheet applies a descending sort filter to a specified column in a spreadsheet sheet
//
// This routine retrieves the values of the target sheet, determines the index
// of the requested column header, obtains the sheet ID, and then constructs a
// batch update request that sorts all rows below the header in descending order
// based on that column. It returns an error if any step fails, otherwise
// completes silently.
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

// addFilterByFailedAndMandatoryToSheet applies a filter to show only failed mandatory tests
//
// This function retrieves the specified sheet’s data, identifies the columns
// for test state and mandatory status, then builds a request to set a basic
// filter that displays rows where the state is "failed" and the test is marked
// as "Mandatory". It executes this filter through a batch update on the
// spreadsheet. If any step fails, it returns an error describing the issue.
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
