GetSheetIDByName`

```go
func (*sheets.Spreadsheet, string) (int64, error)
```

### Purpose  
Locates a worksheet inside a Google Sheets document by its **visible name** and returns the underlying numeric sheet ID used by the API.

The function is part of the *resultsspreadsheet* command that uploads test results to a shared spreadsheet. It is called whenever the program needs to reference a particular sheet (e.g., `"Conclusion"`, `"Raw Results"`).  

### Parameters  
| Parameter | Type                     | Description |
|-----------|--------------------------|-------------|
| `s`       | `*sheets.Spreadsheet`   | The Google Sheets API client that already holds the spreadsheet metadata. |
| `name`    | `string`                 | Human‑readable sheet title to search for. |

### Return Values  
| Value | Type   | Description |
|-------|--------|-------------|
| `int64` | Sheet ID of the matching tab. | Zero if not found. |
| `error` | `nil` on success; otherwise an error describing why the lookup failed (e.g., sheet not present). |

### Key Steps
1. **Iterate** over `s.Sheets`, which is a slice of `*sheets.Sheet`.
2. For each sheet, compare its `Properties.Title` with the supplied `name`.
3. If a match is found, return `sheet.Properties.SheetId`.
4. If no sheet matches, construct an error using `fmt.Errorf("sheet %q not found", name)`.

### Dependencies
- **Google Sheets Go SDK** (`"google.golang.org/api/sheets/v4"`): Provides the `Spreadsheet` and `Sheet` types.
- **`fmt` package**: Used only for formatting the “not found” error message.

No other globals or external state are accessed; the function is pure with respect to the spreadsheet data passed in.

### Side‑Effects
None. The function reads from the provided `Spreadsheet`; it does not modify any fields.

### Package Context
`GetSheetIDByName` lives in **resultsspreadsheet** (cmd/certsuite/upload/results_spreadsheet).  
Other parts of the package call this helper when they need to:
- Add rows to a specific sheet.
- Retrieve existing data from a tab by ID.
- Validate that expected sheets exist before proceeding.

Because the spreadsheet object is already populated (via an API call earlier in the program), this function is a lightweight lookup utility used throughout the upload workflow.
