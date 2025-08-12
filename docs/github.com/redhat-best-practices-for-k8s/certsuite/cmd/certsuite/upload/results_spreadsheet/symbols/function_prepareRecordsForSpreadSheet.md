prepareRecordsForSpreadSheet`

```go
func prepareRecordsForSpreadSheet(data [][]string) []*sheets.RowData
```

### Purpose
Transforms raw tabular results into a format suitable for writing to a Google‑Sheets document.  
The function:

1. Normalises cell content by removing newlines (`"\n"`) and replacing them with spaces.
2. Truncates any value that exceeds the maximum allowed length (`cellContentLimit`).
3. Wraps each processed row in a `*sheets.RowData`, which is the type expected by the
   Google Sheets API.

The resulting slice of `RowData` objects can be fed directly into the spreadsheet‑writing logic
(`uploadResultsToSheet` or similar) without further manipulation.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `data` | `[][]string` | A 2‑dimensional string slice where each inner slice represents a row of raw results. |

### Return value

| Type | Description |
|------|-------------|
| `[]*sheets.RowData` | A slice of pointers to `RowData`, one per input row, ready for API consumption. |

### Key operations

| Step | Implementation detail |
|------|-----------------------|
| 1️⃣ Normalise cell content | Each string is processed with `strings.ReplaceAll(s, "\n", " ")`. This prevents multiline values from breaking the spreadsheet layout. |
| 2️⃣ Truncate long cells | If a value’s length exceeds `cellContentLimit` (defined in `const.go`), it is sliced to that limit. This protects against API limits or UI issues. |
| 3️⃣ Build RowData | The cleaned slice of strings is wrapped in `sheets.RowData{Values: []interface{}{...}}`. Each string becomes an interface{} entry, matching the Sheets API schema. |

### Dependencies

* **`cellContentLimit`** – constant from `const.go`, defines the maximum allowed length for a cell.
* **`strings.ReplaceAll`** – standard library function for newline replacement.
* **`sheets.RowData`** – type from the Google Sheets client library (`google.golang.org/api/sheets/v4`).

### Side‑effects

None. The function is pure: it does not modify global state, I/O, or external resources.

### How it fits in `resultsspreadsheet`

The package orchestrates uploading test results to a spreadsheet:

1. Raw CSV data → parsed into `[][]string`.
2. `prepareRecordsForSpreadSheet` converts that raw matrix into the API‑friendly format.
3. The formatted rows are then written to the target sheet(s) (e.g., `ConclusionSheetName`, `RawResultsSheetName`) by higher‑level functions.

This separation keeps the data transformation logic isolated, making it easier to test and reuse when new sheets or formats are added.
