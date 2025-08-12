createRawResultsSheet`

| Item | Details |
|------|---------|
| **Package** | `resultsspreadsheet` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload/results_spreadsheet`) |
| **Signature** | `func createRawResultsSheet(filePath string) (*sheets.Sheet, error)` |
| **Visibility** | unexported (internal helper) |

### Purpose
Creates a *raw results* sheet that will later be embedded into the final Google Sheet.  
The function:

1. Reads a CSV file containing raw test results.
2. Transforms each record so it fits the spreadsheet schema (`prepareRecordsForSpreadSheet`).
3. Returns a `sheets.Sheet` instance ready for insertion.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `filePath` | `string` | Path to the CSV file that contains raw results. The path is typically derived from the global variable `resultsFilePath`, but callers may supply a custom location. |

### Outputs
| Return | Type | Meaning |
|--------|------|---------|
| `*sheets.Sheet` | *pointer to `Sheet` (from the Google Sheets API)* | Represents a single sheet that can be added to the spreadsheet. The sheet contains the processed rows from the CSV. |
| `error` | `error` | Non‑nil if any step fails: file read, CSV parsing, or record preparation. |

### Key Dependencies & Calls
| Called Function | Role |
|-----------------|------|
| `readCSV(filePath)` | Reads the CSV into a slice of string slices (`[][]string`). |
| `prepareRecordsForSpreadSheet(records [][]string) ([][]interface{}, error)` | Converts raw CSV rows into the format required by Google Sheets (`[]interface{}` per row). Handles type conversion, truncation to `cellContentLimit`, and other formatting rules. |
| `Errorf` (from `fmt`) | Wraps errors with context for easier debugging. |

### Side‑Effects
* No global state is mutated; the function works purely on its input.
* The returned sheet contains no side‑effects on external resources beyond what `prepareRecordsForSpreadSheet` performs.

### Integration Flow
1. **Command Setup** – The CLI command (`uploadResultSpreadSheetCmd`) collects configuration (e.g., `resultsFilePath`, credentials) and initiates spreadsheet creation.
2. **Raw Sheet Creation** – `createRawResultsSheet` is invoked with the path to the CSV. It produces a sheet that represents raw test outcomes.
3. **Sheet Assembly** – The resulting sheet is passed to higher‑level logic that stitches together other sheets (e.g., conclusion, single‑workload) and uploads the final spreadsheet to Google Drive.

### Diagram (Mermaid)

```mermaid
flowchart TD
    A[User provides CSV path] -->|createRawResultsSheet(filePath)| B[Read CSV]
    B --> C{Parse success?}
    C -- Yes --> D[Prepare Records]
    D --> E[Return *sheets.Sheet]
    C -- No --> F[Error: fmt.Errorf("failed to read csv")]
```

### Summary
`createRawResultsSheet` is a small, focused helper that bridges raw CSV data and the Google Sheets API. It encapsulates file I/O, data sanitization, and sheet construction while keeping side‑effects minimal. This function is invoked as part of the larger spreadsheet generation pipeline within the `certsuite` upload command.
