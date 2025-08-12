createSingleWorkloadRawResultsSpreadSheet`

| Aspect | Detail |
|--------|--------|
| **Signature** | `func createSingleWorkloadRawResultsSpreadSheet(sheetsService *sheets.Service, driveService *drive.Service, file *drive.File, sheet *sheets.Sheet, workloadName string) (*sheets.Spreadsheet, error)` |
| **Export status** | Unexported (private to the package). |

### Purpose
Creates a Google Sheets spreadsheet that contains *raw* test results for a single workload.  
The function:

1. Builds a new sheet inside an existing spreadsheet (`file`) with headers defined by `conclusionSheetHeaders`.  
2. Adds a filter on the sheet that shows only rows where **Failed** and **Mandatory** flags are set.  
3. Moves the resulting spreadsheet into a Google Drive folder identified by the `rootFolderURL` variable.  

The returned `*sheets.Spreadsheet` contains the fully‑configured sheet ready for further processing or upload.

### Parameters

| Name | Type | Role |
|------|------|------|
| `sheetsService` | `*sheets.Service` | API client used to create and modify sheets. |
| `driveService` | `*drive.Service` | API client used to move the spreadsheet into a Drive folder. |
| `file` | `*drive.File` | The Google Sheet file that will receive the new sheet. |
| `sheet` | `*sheets.Sheet` | Metadata for the sheet being created (e.g., title, index). |
| `workloadName` | `string` | Name of the workload; used to name the sheet and log messages. |

### Return Values

| Value | Type | Meaning |
|-------|------|---------|
| `*sheets.Spreadsheet` | Spreadsheet object | The fully‑configured spreadsheet after adding the new sheet and filter. |
| `error` | error | Non‑nil if any API call fails (sheet creation, filter addition, or Drive move). |

### Key Steps & Dependencies

1. **Sheet Creation**  
   ```go
   createSingleWorkloadRawResultsSheet(sheetsService, file.ID, sheet, conclusionSheetHeaders)
   ```  
   - Builds a new sheet with the specified headers.
   - Uses `conclusionSheetHeaders` (a slice of strings defined elsewhere in the package) as column titles.

2. **Filter Setup**  
   ```go
   addFilterByFailedAndMandatoryToSheet(sheetsService, file.ID, sheet.Properties.Index)
   ```  
   - Applies a filter that shows only rows where both *Failed* and *Mandatory* columns are true.
   - The function is defined elsewhere in the package.

3. **Move to Drive Folder**  
   ```go
   MoveSpreadSheetToFolder(driveService, file.ID, rootFolderURL)
   ```  
   - Moves the spreadsheet into a folder referenced by `rootFolderURL` (a global string).
   - Logs the action with `fmt.Printf`.

4. **Error Handling**  
   Each API call (`Create`, `Do`) is wrapped in error checks; any failure bubbles up as the function’s return value.

### Side Effects

- The original spreadsheet (`file`) receives a new sheet.
- The spreadsheet file is moved to another Google Drive folder (side‑effect on Drive metadata).
- Logs are printed to standard output for progress tracking.

### How It Fits in the Package

The `resultsspreadsheet` package orchestrates creation and organization of test result sheets.  
Functions like `createSingleWorkloadRawResultsSpreadSheet` are helpers that:

- Keep the main flow (e.g., command‑line upload logic) clean.
- Encapsulate specific API interactions.
- Ensure consistent naming, filtering, and folder placement across all generated spreadsheets.

They are called by higher‑level functions that iterate over workloads or test runs, producing a structured set of raw result sheets for further analysis.
