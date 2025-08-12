GetHeadersFromSheet`

**Package:** `resultsspreadsheet`  
**File:** `sheet_utils.go:9`  

### Purpose
Extracts the column header names from a Google Sheets worksheet (`*sheets.Sheet`).  
The returned slice contains every cell value found in the first row of the sheet, preserving order.

### Signature
```go
func GetHeadersFromSheet(sheet *sheets.Sheet) []string
```

| Parameter | Type              | Description |
|-----------|-------------------|-------------|
| `sheet`   | `*sheets.Sheet`   | A Google Sheets API representation of a worksheet. It must contain at least one row; otherwise the function returns an empty slice. |

### Return Value
- `[]string`: Ordered list of header strings from the first row.

### Key Dependencies
- **Google Sheets API (`google.golang.org/api/sheets/v4`)** – the type `sheets.Sheet` and its nested structures are used to access cell values.
- No external packages or global variables are referenced; the function is pure and deterministic.

### Implementation Details (concise)
1. Initialize an empty slice of strings, `headers`.
2. Iterate over each row in `sheet.Data.RowData`.  
   *Stop after the first row* because headers reside only there.
3. For every cell (`CellValue`) in that row, append its value to `headers`.
4. Return the populated slice.

### Side Effects
- None. The function does not modify the input sheet or any global state.

### Context within the Package
`GetHeadersFromSheet` is a small utility used by higher‑level routines that process spreadsheet data.  
It allows callers to retrieve column names before mapping raw cell values to structured fields, facilitating flexible handling of sheets with varying header configurations.

---  

#### Mermaid Diagram (optional)

```mermaid
flowchart TD
    A[GetHeadersFromSheet] --> B{sheet.Data.RowData}
    B -->|first row only| C[Iterate cells]
    C --> D[Append CellValue to headers]
    D --> E[Return []string]
```
