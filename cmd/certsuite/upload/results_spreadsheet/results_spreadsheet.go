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

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	resultsFilePath string
	mainFolderURL   string
	ocpVersion      string
)

var workloadsResultsFolder *drive.File
var stringToPointer = func(s string) *string { return &s }

const (
	categoryConclusionsCol        = "Category"
	workloadVersionConclusionsCol = "Workload Version"
	ocpVersionConclusionsCol      = "OCP Version"
	workloadNameConclusionsCol    = "Workload Name"
	resultsConclusionsCol         = "Results"
	cellContantLimit              = 50000
)

var conclusionSheetHeaders = []string{categoryConclusionsCol, workloadVersionConclusionsCol, ocpVersionConclusionsCol, workloadNameConclusionsCol, resultsConclusionsCol}

var (
	uploadResultSpreadSheetCmd = &cobra.Command{
		Use:   "results-spreadsheet",
		Short: "Generates a google spread sheets with test suite results.",
		Run: func(cmd *cobra.Command, args []string) {
			generateResultsSpreadSheet()
		},
	}
)

func NewCommand() *cobra.Command {
	uploadResultSpreadSheetCmd.Flags().StringVarP(&resultsFilePath, "results-file", "f", "", "Required: path to results file")
	uploadResultSpreadSheetCmd.Flags().StringVarP(&mainFolderURL, "dest-url", "d", "", "Required: Destination drive folder's URL")
	uploadResultSpreadSheetCmd.Flags().StringVarP(&ocpVersion, "version", "v", "", "Optional: OCP Version")

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

func createSheetsAndDriveServices() (sheetService *sheets.Service, driveService *drive.Service, err error) {
	ctx := context.Background()
	b, err := os.ReadFile("cmd/certsuite/upload/results_spreadsheet/credentials.json")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope, drive.DriveScope)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get client: %v", err)
	}

	sheetSrv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return sheetSrv, driveSrv, nil
}

func prepareRecordsForSpreadSheet(records [][]string) []*sheets.RowData {
	var rows []*sheets.RowData
	for _, row := range records {
		var rowData []*sheets.CellData
		for _, col := range row {
			var val string
			// cell content cannot exceed 50,000 letters.
			if len(col) > cellContantLimit {
				col = col[:cellContantLimit]
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

func createSingleWorkloadRawResultsSheet(rawResultsSheet *sheets.Sheet, wantedWorkloadName string) (*sheets.Sheet, error) {
	// Initialize sheet with headers
	filteredRows := []*sheets.RowData{{Values: []*sheets.CellData{
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: stringToPointer("Owner/TechLead Conclusion")}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: stringToPointer("Next Step Actions")}},
	}}}
	filteredRows[0].Values = append(filteredRows[0].Values, rawResultsSheet.Data[0].RowData[0].Values...)

	headers := getHeadersAsInterfaceList(rawResultsSheet)
	indices, err := getIndicesFromListByName(headers, []string{"CNFName"})
	if err != nil {
		return nil, err
	}
	wantedWorkloadIndex := indices[0]

	// add to sheet only rows of given workload name
	for _, row := range rawResultsSheet.Data[0].RowData[1:] {
		if len(row.Values) > wantedWorkloadIndex {
			curWorkloadName := *row.Values[wantedWorkloadIndex].UserEnteredValue.StringValue
			if curWorkloadName == wantedWorkloadName {
				// add empty values in 2 added columns
				newRow := &sheets.RowData{
					Values: append([]*sheets.CellData{{}, {}}, row.Values...),
				}
				filteredRows = append(filteredRows, newRow)
			}
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

func createSingleWorkloadRawResultsSpreadSheet(sheetService *sheets.Service, driveService *drive.Service, folder *drive.File, rawResultsSheet *sheets.Sheet, wantedWorkloadName string) (*sheets.Spreadsheet, error) {
	workloadResultsSheet, err := createSingleWorkloadRawResultsSheet(rawResultsSheet, wantedWorkloadName)
	if err != nil {
		return nil, err
	}

	workloadResultsSpreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: fmt.Sprintf("%s Best Practices Test Results", wantedWorkloadName),
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

	return workloadResultsSpreadsheet, nil
}

//nolint:funlen
func createConclusionsSheet(sheetsService *sheets.Service, driveService *drive.Service, rawResultsSheet *sheets.Sheet) (*sheets.Sheet, error) {
	headers := getHeadersAsInterfaceList(rawResultsSheet)
	colsIndices, err := getIndicesFromListByName(headers, []string{"CNFName", "CNFType", "OperatorVersion"})
	if err != nil {
		return nil, err
	}

	workloadNameColIndex := colsIndices[0]
	workloadTypeColIndex := colsIndices[1]
	operatorVersionColIndex := colsIndices[2]

	// Initialize sheet with headers
	conclusionsSheetRowsValues := []*sheets.CellData{}
	for _, colHeader := range conclusionSheetHeaders {
		headerCellData := &sheets.CellData{UserEnteredValue: &sheets.ExtendedValue{StringValue: stringToPointer(colHeader)}}
		conclusionsSheetRowsValues = append(conclusionsSheetRowsValues, headerCellData)
	}
	conclusionsSheetRows := []*sheets.RowData{{Values: conclusionsSheetRowsValues}}

	// Extract unique values from the CNFName column and fill sheet
	uniqueMap := make(map[string]bool)
	for _, rawResultsSheetrow := range rawResultsSheet.Data[0].RowData[1:] {
		workloadName := *rawResultsSheetrow.Values[workloadNameColIndex].UserEnteredValue.StringValue
		// if workload has already been added to sheet, skip it
		if uniqueMap[workloadName] {
			continue
		}
		uniqueMap[workloadName] = true

		curConsclusionRowValues := []*sheets.CellData{}
		for _, colHeader := range conclusionSheetHeaders {
			curCellData := &sheets.CellData{UserEnteredValue: &sheets.ExtendedValue{}}
			var value string

			switch colHeader {
			case categoryConclusionsCol:
				value = *rawResultsSheetrow.Values[workloadTypeColIndex].UserEnteredValue.StringValue

			case workloadVersionConclusionsCol:
				value = *rawResultsSheetrow.Values[operatorVersionColIndex].UserEnteredValue.StringValue

			case ocpVersionConclusionsCol:
				value = ocpVersion + " "

			case workloadNameConclusionsCol:
				value = workloadName

			case resultsConclusionsCol:
				workloadResultsSpreadsheet, err := createSingleWorkloadRawResultsSpreadSheet(sheetsService, driveService, workloadsResultsFolder, rawResultsSheet, workloadName)
				if err != nil {
					return nil, fmt.Errorf("error has occurred while creating %s results file: %v", workloadName, err)
				}

				hyperlinkFormula := fmt.Sprintf("=HYPERLINK(%q, %q)", workloadResultsSpreadsheet.SpreadsheetUrl, "Results")
				curCellData.UserEnteredValue.FormulaValue = &hyperlinkFormula

			default:
				// use space for empty values to avoid cells overlapping
				value = " "
			}

			if colHeader != resultsConclusionsCol {
				curCellData.UserEnteredValue.StringValue = &value
			}

			curConsclusionRowValues = append(curConsclusionRowValues, curCellData)
		}
		conclusionsSheetRows = append(conclusionsSheetRows, &sheets.RowData{Values: curConsclusionRowValues})
	}

	conclusionSheet := &sheets.Sheet{
		Properties: &sheets.SheetProperties{
			Title:          "conclusions",
			GridProperties: &sheets.GridProperties{FrozenRowCount: 1},
		},
		Data: []*sheets.GridData{{RowData: conclusionsSheetRows}},
	}

	return conclusionSheet, nil
}

func createRawResultsSheet(fp string) (*sheets.Sheet, error) {
	records, err := readCSV(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file: %v", err)
	}

	rows := prepareRecordsForSpreadSheet(records)

	rawResultsSheet := &sheets.Sheet{
		Properties: &sheets.SheetProperties{
			Title:          "raw results",
			GridProperties: &sheets.GridProperties{FrozenRowCount: 1},
		},
		Data: []*sheets.GridData{{RowData: rows}},
	}

	return rawResultsSheet, err
}

func generateResultsSpreadSheet() {
	sheetService, driveService, err := createSheetsAndDriveServices()
	if err != nil {
		log.Fatalf("Unable to create services: %v", err)
	}

	mainFolderName := strings.TrimLeft(fmt.Sprintf("%s Redhat Best Practices for K8 Test Results %s", ocpVersion, time.Now().Format("2006-01-02T15:04:05Z07:00")), " ")
	mainFolderParentID, err := extractFolderIDFromURL(mainFolderURL)
	if err != nil {
		log.Fatalf("error getting folder ID from URL")
	}
	mainResultsFolder, err := createDriveFolder(driveService, mainFolderName, mainFolderParentID)
	if err != nil {
		log.Fatalf("Unable to create main results folder: %v", err)
	}

	workloadsFolderName := "Results Per Workload"
	workloadsResultsFolder, err = createDriveFolder(driveService, workloadsFolderName, mainResultsFolder.Id)
	if err != nil {
		log.Fatalf("Unable to create workloads results folder: %v", err)
	}

	rawResultsSheet, err := createRawResultsSheet(resultsFilePath)
	if err != nil {
		log.Fatalf("Unable to create raw results sheet: %v", err)
	}

	conclusionSheet, err := createConclusionsSheet(sheetService, driveService, rawResultsSheet)
	if err != nil {
		log.Fatalf("Unable to create conclusions sheet: %v", err)
	}

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
