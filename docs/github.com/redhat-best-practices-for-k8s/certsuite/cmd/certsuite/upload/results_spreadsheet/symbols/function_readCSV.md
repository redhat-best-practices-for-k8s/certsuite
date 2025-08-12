readCSV`

```go
func readCSV(file string) ([][]string, error)
```

| Aspect | Description |
|--------|-------------|
| **Purpose** | Reads a CSV file located at *file* and returns all rows as a slice of string slices (`[][]string`). It is a small helper used by the spreadsheet‑generation code to ingest raw test results. |
| **Inputs** | `file` – the absolute or relative path to the CSV file that contains the results data. |
| **Outputs** | *data* – `[][]string`, each inner slice represents one row of the CSV (including header).<br>*err* – non‑nil if any I/O, permission, or parsing error occurs. |
| **Key dependencies** | • `os.Open` – opens the file for reading.<br>• `defer f.Close()` – guarantees the file descriptor is released even on errors.<br>• `csv.NewReader(f)` – creates a CSV reader that respects standard RFC‑4180 delimiters.<br>• `r.ReadAll()` – consumes the entire file and returns all records. |
| **Side effects** | *None beyond reading.* The function does not modify any global state, only opens/closes the file. |
| **Error handling** | Errors from `Open`, `ReadAll` or `Close` are propagated directly to the caller; no retry logic is performed here. |

### How it fits the package

The `resultsspreadsheet` command processes raw test result files and generates an Excel workbook for reporting.  
* `readCSV` is invoked by higher‑level functions that need to transform CSV input into a data structure suitable for populating spreadsheet sheets (e.g., `conclusionSheetHeaders`, `rawResultsSheetHeaders`).  
* By keeping the file‑reading logic isolated, the rest of the code can focus on mapping rows to sheet columns without worrying about I/O concerns.

### Usage example

```go
rows, err := readCSV("/tmp/results.csv")
if err != nil {
    log.Fatalf("cannot load results: %v", err)
}
for _, r := range rows {
    // r[0] is the first column of that row
}
```

---

#### Mermaid diagram (optional)

```mermaid
flowchart TD
  A[Caller] -->|calls readCSV(file)| B[readCSV]
  B -->|os.Open| C[FileHandle]
  C -->|csv.NewReader| D[CSV Reader]
  D -->|ReadAll| E[[Rows]]
  E -->|return rows, nil| B
  B -->|defer Close| C
```

*This diagram shows the flow of data from the caller through file opening, CSV parsing, and return.*
