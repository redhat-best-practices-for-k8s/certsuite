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

// NewCommand creates the cobra command that uploads test results to a spreadsheet.
//
// It defines flags for the results file path, root folder URL, OCP version, and credentials,
// marks required flags, and returns the constructed *cobra.Command instance ready for use in
// the certsuite CLI.
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

// readCSV opens a CSV file at the given path, reads all records, and returns them as a slice of string slices.
// It returns an error if the file cannot be opened or read.
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

// CreateSheetsAndDriveServices creates Google Sheets and Drive services for uploading results.
//
// It takes a path to a credentials file and returns initialized *sheets.Service, *drive.Service,
// or an error if the services cannot be created. The function uses the provided credentials
// to authenticate with the Google APIs and prepares the services for subsequent spreadsheet
// and file operations.
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

// prepareRecordsForSpreadSheet converts raw CSV rows into spreadsheet row data.
//
// It accepts a slice of string slices, each representing a row of
// values. For every cell it removes carriage returns and line breaks,
// then appends the cleaned value to a new RowData structure.
// The resulting slice of *sheets.RowData is returned for use when
// populating the results spreadsheet.
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

// createSingleWorkloadRawResultsSheet creates a new sheet containing test case results for a single workload extracted from a raw results sheet that may contain multiple workloads.  
// It copies the original header columns and adds two additional columns: "Owner/TechLead Conclusion" and "Next Step Actions". The function returns the newly created sheet or an error if the input data is empty or headers cannot be retrieved.
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

// createSingleWorkloadRawResultsSpreadSheet creates a new Google Sheets spreadsheet containing raw results for a single workload, applies filtering to show only failed or mandatory entries, and moves the file into the designated folder.
//
// createSingleWorkloadRawResultsSpreadSheet creates a new Google Sheets spreadsheet containing raw results for a single workload. It first builds the sheet structure using createSingleWorkloadRawResultsSheet, then populates it with data via the sheets service. After the sheet is populated, a filter is applied to display only failed or mandatory rows. The spreadsheet file is created in Drive, moved to the target folder, and its URL is logged. The function returns the resulting Spreadsheet object and any error that occurred during the process.
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

// createConclusionsSheet creates a new Google Sheets tab that summarizes unique workload results extracted from raw result sheets.
//
// It builds a sheet with headers such as Category, Workload Version, OCP Version, Workload Name and Results. The Results column contains hyperlinks to the corresponding raw results spreadsheets for each workload. The function takes a Google Sheets service, a Drive service, a pointer to an existing Sheet object and a string identifier, then returns the newly created Sheet and any error that occurs during creation.
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

// createRawResultsSheet creates a spreadsheet sheet containing raw test results from a CSV file.
//
// It reads the CSV located at the given path, prepares the data for insertion into
// Google Sheets by calling prepareRecordsForSpreadSheet, and returns a pointer to
// the resulting sheets.Sheet along with any error that occurs during reading or
// preparation. The function does not expose the sheet name; callers use the
// returned sheet directly in the spreadsheet generation workflow.
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

// generateResultsSpreadSheet creates a Google Sheets document that aggregates raw test results and their conclusions, then moves it to the specified Drive folder.
//
// It initializes the Drive and Sheets services, parses the root folder URL for its ID, and creates a new spreadsheet with separate sheets for raw data and summarized conclusions. The function writes headers, applies basic filtering and sorting, and names the file with a timestamp. Finally, it relocates the spreadsheet into the target folder and reports any errors via logging.
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
