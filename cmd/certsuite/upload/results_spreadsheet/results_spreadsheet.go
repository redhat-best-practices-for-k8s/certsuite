package resultsspreadsheet

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var stringToPointer = func(s string) *string { return &s }
var conclusionSheetHeaders = []string{categoryConclusionsCol, workloadVersionConclusionsCol, ocpVersionConclusionsCol, WorkloadNameConclusionsCol, ResultsConclusionsCol}

var (
	resultsFilePath string
	rootFolderURL   string
	ocpVersion      string
	credentials     string
)

var (
	uploadResultSpreadSheetCmd = &cobra.Command{
		Use:   "results-spreadsheet",
		Short: "Generates a google spread sheets with test suite results.",
		Run: func(cmd *cobra.Command, args []string) {
			generateResultsSpreadSheet()
		},
	}
)

// NewCommand Creates a command for uploading results spreadsheets
//
// This function configures flags for the spreadsheet upload command, including
// paths to the results file, destination URL, optional OCP version, and
// credentials file. It marks the required flags and handles errors by logging
// fatal messages if flag validation fails. The configured command is then
// returned for use in the larger CLI.
func NewCommand() *cobra.Command {
	uploadResultSpreadSheetCmd.Flags().StringVarP(&resultsFilePath, "results-file", "f", "", "Required: path to results file")
	uploadResultSpreadSheetCmd.Flags().StringVarP(&rootFolderURL, "dest-url", "d", "", "Required: Destination drive folder's URL")
	uploadResultSpreadSheetCmd.Flags().StringVarP(&ocpVersion, "version", "v", "", "Optional: OCP Version")
	uploadResultSpreadSheetCmd.Flags().StringVarP(&credentials, "credentials", "c", "credentials.json", "Optional: Google credentials file path, default path: credentials.json")

	err := uploadResultSpreadSheetCmd.MarkFlagRequired("results-file")
	if err != nil {
		log.Fatalf("Failed to mark results file path as required parameter: %v", err)
		return nil
	}

	err = uploadResultSpreadSheetCmd.MarkFlagRequired("dest-url")
	if err != nil {
		log.Fatalf("Failed to mark dest url path as required parameter: %v", err)
		return nil
	}

	return uploadResultSpreadSheetCmd
}

// readCSV loads CSV file contents into a two-dimensional string slice
//
// The function opens the specified file path, reads all rows using the csv
// package, and returns them as a slice of records where each record is a slice
// of fields. It propagates any I/O or parsing errors to the caller. The file is
// closed automatically via defer before returning.
func readCSV(fp string) ([][]string, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateSheetsAndDriveServices Initializes Google Sheets and Drive services
//
// This function takes a path to credentials and uses it to create authenticated
// clients for both the Sheets and Drive APIs. It returns the two service
// instances or an error if either creation fails.
func CreateSheetsAndDriveServices(credentials string) (sheetService *sheets.Service, driveService *drive.Service, err error) {
	ctx := context.TODO()

	sheetSrv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve Sheets service: %v", err)
	}

	driveSrv, err := drive.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve Drive service: %v", err)
	}

	return sheetSrv, driveSrv, nil
}

// prepareRecordsForSpreadSheet Converts CSV rows into spreadsheet row data
//
// This routine takes a two‑dimensional string slice, representing CSV
// records, and transforms each cell into a CellData object suitable for Google
// Sheets. It trims overly long content to a predefined limit, replaces empty
// cells with a single space to preserve layout, and removes line breaks from
// text. Each processed row is wrapped in a RowData structure; the function
// returns a slice of these rows for use in sheet creation.
func prepareRecordsForSpreadSheet(records [][]string) []*sheets.RowData {
	var rows []*sheets.RowData
	for _, row := range records {
		var rowData []*sheets.CellData
		for _, col := range row {
			var val string
			// cell content cannot exceed 50,000 letters.
			if len(col) > cellContentLimit {
				col = col[:cellContentLimit]
			}
			// use space for empty values to avoid cells overlapping
			if col == "" {
				val = " "
			}
			// avoid line breaks in cell
			val = strings.ReplaceAll(strings.ReplaceAll(col, "\r\n", " "), "\n", " ")

			rowData = append(rowData, &sheets.CellData{
				UserEnteredValue: &sheets.ExtendedValue{StringValue: &val},
			})
		}
		rows = append(rows, &sheets.RowData{Values: rowData})
	}
	return rows
}

// createSingleWorkloadRawResultsSheet Creates a new sheet containing only the rows for a specified workload
//
// The function filters an existing raw results sheet to include only the test
// case rows that match the given workload name, adding two empty columns for
// owner/tech lead conclusion and next step actions. It retains the original
// header row from the raw sheet while inserting the new headers at the
// beginning. The resulting sheet is returned along with any error encountered
// during processing.
func createSingleWorkloadRawResultsSheet(rawResultsSheet *sheets.Sheet, workloadName string) (*sheets.Sheet, error) {
	// Initialize sheet with the two new column headers only.
	filteredRows := []*sheets.RowData{{Values: []*sheets.CellData{
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: stringToPointer(conclusionIndividualSingleWorkloadSheetCol)}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: stringToPointer(nextStepAIIfFailSingleWorkloadSheetCol)}},
	}}}

	// Add existing column headers from the rawResultsSheet
	filteredRows[0].Values = append(filteredRows[0].Values, rawResultsSheet.Data[0].RowData[0].Values...)

	headers := GetHeadersFromSheet(rawResultsSheet)
	indices, err := GetHeaderIndicesByColumnNames(headers, []string{"CNFName"})
	if err != nil {
		return nil, err
	}
	workloadNameIndex := indices[0]

	// add to sheet only rows of given workload name
	for _, row := range rawResultsSheet.Data[0].RowData[1:] {
		if len(row.Values) <= workloadNameIndex {
			return nil, fmt.Errorf("workload %s not found in raw spreadsheet", workloadName)
		}
		curWorkloadName := *row.Values[workloadNameIndex].UserEnteredValue.StringValue
		if curWorkloadName == workloadName {
			// add empty values in 2 added columns
			newRow := &sheets.RowData{
				Values: append([]*sheets.CellData{{}, {}}, row.Values...),
			}
			filteredRows = append(filteredRows, newRow)
		}
	}

	workloadResultsSheet := &sheets.Sheet{
		Properties: &sheets.SheetProperties{
			Title: "results",
		},
		Data: []*sheets.GridData{{RowData: filteredRows}},
	}

	return workloadResultsSheet, nil
}

// createSingleWorkloadRawResultsSpreadSheet Creates a Google Sheets spreadsheet containing raw results for a specific workload
//
// The function builds a new sheet from the provided raw results, then creates a
// spreadsheet titled with the workload name. It applies a filter to show only
// failed or mandatory entries and moves the file into the designated Drive
// folder. Errors are returned if any step fails, and a log message confirms
// creation.
func createSingleWorkloadRawResultsSpreadSheet(sheetService *sheets.Service, driveService *drive.Service, folder *drive.File, rawResultsSheet *sheets.Sheet, workloadName string) (*sheets.Spreadsheet, error) {
	workloadResultsSheet, err := createSingleWorkloadRawResultsSheet(rawResultsSheet, workloadName)
	if err != nil {
		return nil, err
	}

	workloadResultsSpreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: fmt.Sprintf("%s Best Practices Test Results", workloadName),
		},
		Sheets: []*sheets.Sheet{workloadResultsSheet},
	}

	workloadResultsSpreadsheet, err = sheetService.Spreadsheets.Create(workloadResultsSpreadsheet).Do()
	if err != nil {
		return nil, err
	}

	if err := addFilterByFailedAndMandatoryToSheet(sheetService, workloadResultsSpreadsheet, "results"); err != nil {
		return nil, err
	}

	if err := MoveSpreadSheetToFolder(driveService, folder, workloadResultsSpreadsheet); err != nil {
		return nil, err
	}

	log.Printf("%s workload's results sheet has been created.\n", workloadName)

	return workloadResultsSpreadsheet, nil
}

// createConclusionsSheet Creates a conclusion sheet summarizing unique workloads
//
// The function builds a new Google Sheets tab that lists each distinct workload
// from the raw results, along with its category, version, OCP release, and a
// hyperlink to a dedicated results spreadsheet. It first creates a folder for
// per‑workload sheets, then iterates over the raw data rows, extracting
// unique names and assembling row values. For every new workload it generates
// an individual results file and inserts a link; if any step fails it returns
// an error.
//
//nolint:funlen
func createConclusionsSheet(sheetsService *sheets.Service, driveService *drive.Service, rawResultsSheet *sheets.Sheet, mainResultsFolderID string) (*sheets.Sheet, error) {
	workloadsFolderName := "Results Per Workload"
	workloadsResultsFolder, err := createDriveFolder(driveService, workloadsFolderName, mainResultsFolderID)
	if err != nil {
		return nil, fmt.Errorf("unable to create workloads results folder: %v", err)
	}

	rawSheetHeaders := GetHeadersFromSheet(rawResultsSheet)
	colsIndices, err := GetHeaderIndicesByColumnNames(rawSheetHeaders, []string{workloadNameRawResultsCol, workloadTypeRawResultsCol, operatorVersionRawResultsCol})
	if err != nil {
		return nil, err
	}

	workloadNameColIndex := colsIndices[0]
	workloadTypeColIndex := colsIndices[1]
	operatorVersionColIndex := colsIndices[2]

	// Initialize sheet with headers
	conclusionsSheetRowsValues := []*sheets.CellData{}
	for _, colHeader := range conclusionSheetHeaders {
		headerCellData := &sheets.CellData{UserEnteredValue: &sheets.ExtendedValue{StringValue: &colHeader}}
		conclusionsSheetRowsValues = append(conclusionsSheetRowsValues, headerCellData)
	}
	conclusionsSheetRows := []*sheets.RowData{{Values: conclusionsSheetRowsValues}}

	// If rawResultsSheet has now workloads data, return an error
	if len(rawResultsSheet.Data[0].RowData) <= 1 {
		return nil, fmt.Errorf("raw results has no workloads data")
	}

	// Extract unique values from the CNFName column and fill sheet
	uniqueWorkloadNames := make(map[string]bool)
	for _, rawResultsSheetrow := range rawResultsSheet.Data[0].RowData[1:] {
		workloadName := *rawResultsSheetrow.Values[workloadNameColIndex].UserEnteredValue.StringValue
		// if workload has already been added to sheet, skip it
		if uniqueWorkloadNames[workloadName] {
			continue
		}
		uniqueWorkloadNames[workloadName] = true

		curConsclusionRowValues := []*sheets.CellData{}
		for _, colHeader := range conclusionSheetHeaders {
			curCellData := &sheets.CellData{UserEnteredValue: &sheets.ExtendedValue{}}

			switch colHeader {
			case categoryConclusionsCol:
				curCellData.UserEnteredValue.StringValue = rawResultsSheetrow.Values[workloadTypeColIndex].UserEnteredValue.StringValue

			case workloadVersionConclusionsCol:
				curCellData.UserEnteredValue.StringValue = rawResultsSheetrow.Values[operatorVersionColIndex].UserEnteredValue.StringValue

			case ocpVersionConclusionsCol:
				curCellData.UserEnteredValue.StringValue = stringToPointer(ocpVersion + " ")

			case WorkloadNameConclusionsCol:
				curCellData.UserEnteredValue.StringValue = &workloadName

			case ResultsConclusionsCol:
				workloadResultsSpreadsheet, err := createSingleWorkloadRawResultsSpreadSheet(sheetsService, driveService, workloadsResultsFolder, rawResultsSheet, workloadName)
				if err != nil {
					return nil, fmt.Errorf("error has occurred while creating %s results file: %v", workloadName, err)
				}

				hyperlinkFormula := fmt.Sprintf("=HYPERLINK(%q, %q)", workloadResultsSpreadsheet.SpreadsheetUrl, "Results")
				curCellData.UserEnteredValue.FormulaValue = &hyperlinkFormula

			default:
				// use space for empty values to avoid cells overlapping
				curCellData.UserEnteredValue.StringValue = stringToPointer(" ")
			}

			curConsclusionRowValues = append(curConsclusionRowValues, curCellData)
		}
		conclusionsSheetRows = append(conclusionsSheetRows, &sheets.RowData{Values: curConsclusionRowValues})
	}

	conclusionSheet := &sheets.Sheet{
		Properties: &sheets.SheetProperties{
			Title:          ConclusionSheetName,
			GridProperties: &sheets.GridProperties{FrozenRowCount: 1},
		},
		Data: []*sheets.GridData{{RowData: conclusionsSheetRows}},
	}

	return conclusionSheet, nil
}

// createRawResultsSheet parses a CSV file into a Google Sheets sheet
//
// The function reads the specified CSV file, converts each row into spreadsheet
// rows while trimming overly long cell content and normalizing empty cells and
// line breaks. It builds a Sheet object with a title and frozen header row,
// then returns this sheet or an error if reading fails.
func createRawResultsSheet(fp string) (*sheets.Sheet, error) {
	records, err := readCSV(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file: %v", err)
	}

	rows := prepareRecordsForSpreadSheet(records)

	rawResultsSheet := &sheets.Sheet{
		Properties: &sheets.SheetProperties{
			Title:          RawResultsSheetName,
			GridProperties: &sheets.GridProperties{FrozenRowCount: 1},
		},
		Data: []*sheets.GridData{{RowData: rows}},
	}

	return rawResultsSheet, nil
}

// generateResultsSpreadSheet Creates a Google Sheets document with raw results and conclusions
//
// This routine establishes Google Sheets and Drive services, extracts the root
// folder ID from a URL, and creates a main results folder named with the OCP
// version and timestamp. It then builds a raw results sheet from a CSV file and
// a conclusions sheet that aggregates workload data, moves the new spreadsheet
// into the created folder, applies basic filtering, sorts by category, and
// prints the final URL.
func generateResultsSpreadSheet() {
	sheetService, driveService, err := CreateSheetsAndDriveServices(credentials)
	if err != nil {
		log.Fatalf("Unable to create services: %v", err)
	}

	rootFolderID, err := extractFolderIDFromURL(rootFolderURL)
	if err != nil {
		log.Fatalf("error getting folder ID from URL")
	}
	mainFolderName := strings.TrimLeft(fmt.Sprintf("%s Redhat Best Practices for K8 Test Results %s", ocpVersion, time.Now().Format("2006-01-02T15:04:05Z07:00")), " ")
	mainResultsFolder, err := createDriveFolder(driveService, mainFolderName, rootFolderID)
	if err != nil {
		log.Fatalf("Unable to create main results folder: %v", err)
	}

	log.Printf("Generating raw results sheet...")
	rawResultsSheet, err := createRawResultsSheet(resultsFilePath)
	if err != nil {
		log.Fatalf("Unable to create raw results sheet: %v", err)
	}
	log.Printf("Raw results sheet has been generated.")

	log.Printf("Generating conclusion sheet...")
	conclusionSheet, err := createConclusionsSheet(sheetService, driveService, rawResultsSheet, mainResultsFolder.Id)
	if err != nil {
		log.Fatalf("Unable to create conclusions sheet: %v", err)
	}
	log.Printf("Conclusion sheet has been generated.")

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: strings.TrimLeft(fmt.Sprintf("%s Redhat Best Practices for K8 Test Results", ocpVersion), " "),
		},
		Sheets: []*sheets.Sheet{rawResultsSheet, conclusionSheet},
	}

	spreadsheet, err = sheetService.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		log.Fatalf("Unable to create spreadsheet: %v", err)
	}

	if err := MoveSpreadSheetToFolder(driveService, mainResultsFolder, spreadsheet); err != nil {
		log.Fatal(err)
	}

	if err = addBasicFilterToSpreadSheet(sheetService, spreadsheet); err != nil {
		log.Fatalf("Unable to apply filter to the spread sheet: %v", err)
	}

	if err = addDescendingSortFilterToSheet(sheetService, spreadsheet, conclusionSheet.Properties.Title, "Category"); err != nil {
		log.Fatalf("Unable to apply filter to the spread sheet: %v", err)
	}

	fmt.Printf("Results spreadsheet was created successfully: %s\n", spreadsheet.SpreadsheetUrl)
}
