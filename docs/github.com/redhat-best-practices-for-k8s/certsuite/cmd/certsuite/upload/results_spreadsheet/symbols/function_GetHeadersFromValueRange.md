GetHeadersFromValueRange`

```go
func (*sheets.ValueRange) GetHeadersFromValueRange() []string
```

### Purpose
`GetHeadersFromValueRange` extracts the header row from a Google Sheets **ValueRange** object and returns it as a slice of strings.  
The function is used by other parts of the package to:

1. Identify column names in a sheet.
2. Map data values to the corresponding columns when writing results.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `vr` | `*sheets.ValueRange` | The ValueRange returned from a Sheets read request. It contains a 2‑D slice (`Rows`) of cell values, where the first row is expected to hold header names.

### Return
- `[]string`: A one‑dimensional slice containing the string representation of each header value in order.

### Implementation details

```go
func (vr *sheets.ValueRange) GetHeadersFromValueRange() []string {
    var headers []string
    if len(vr.Values) == 0 { // no rows present
        return headers
    }
    firstRow := vr.Values[0]
    for _, cell := range firstRow {
        switch v := cell.(type) {
        case string:
            headers = append(headers, v)
        default:
            headers = append(headers, fmt.Sprint(v))
        }
    }
    return headers
}
```

* The function iterates over the first row (`vr.Values[0]`).
* It handles values that are already `string` and any other type by calling `fmt.Sprint`.
* No external state is mutated; the operation is pure.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `append`   | Builds the result slice. |
| `Sprint` (from `fmt`) | Converts non‑string cell values to a string representation. |

The function has **no side effects**: it only reads from the supplied `ValueRange`. It does not touch any global variables or modify the input.

### Context in the Package

* The package `resultsspreadsheet` orchestrates uploads of test results to Google Sheets.
* `GetHeadersFromValueRange` is a utility used by functions that:
  * Validate that required columns exist before writing data.
  * Build column‑index maps for later use when populating cells.

Because the header extraction logic is trivial yet repeated, it lives in `sheet_utils.go` so that other modules can import and reuse it without duplicating code. The function remains exported to be usable by tests or external callers if needed.
