createSingleWorkloadRawResultsSheet`

| Item | Details |
|------|---------|
| **Signature** | `func(*sheets.Sheet, string) (*sheets.Sheet, error)` |
| **Visibility** | unexported (internal helper) |

### Purpose
Creates a new Google‑Sheets sheet that contains the raw test‑case results for *one* workload extracted from a larger raw‑results sheet that may contain multiple workloads.

The returned sheet:
1. Keeps all columns present in the original `rawResultsSheet`.
2. Adds two **extra columns** at the end:
   * **Owner/TechLead Conclusion** – a free‑text field where the partner/user can record the name of the workload owner who should drive the fix.
   * **Next Step Actions** – a free‑text field for follow‑up actions to resolve the test case.

These two columns are added **after** all existing headers, preserving the original data layout while providing additional context for issue resolution.

> **Note:** The caller must guarantee that `rawResultsSheet` contains at least one row of data; otherwise the function will return an error.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `rawResultsSheet` | `*sheets.Sheet` | Sheet containing raw results for potentially multiple workloads. |
| `workloadName` | `string` | The workload name that should be added to the header of the new sheet, e.g., `"MyWorkload"` |

### Returns
| Value | Type | Description |
|-------|------|-------------|
| `*sheets.Sheet` | `*sheets.Sheet` | Newly created sheet containing only rows for the specified workload and the two additional columns. |
| `error` | `error` | Non‑nil if:
- The original sheet has no headers.
- Header indices cannot be resolved (e.g., missing required column names).
- Any internal helper returns an error.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetHeadersFromSheet` | Retrieves the header row of `rawResultsSheet`. |
| `GetHeaderIndicesByColumnNames` | Maps each expected column name to its index in the original sheet. |
| `stringToPointer` | Utility that converts a string to a `*string`, used for the new columns’ default values. |
| `append` (built‑in) | Adds rows/columns to slices. |
| `Errorf` (`fmt.Errorf`) | Creates error messages when prerequisites fail. |

### How it Works (Step‑by‑step)

1. **Header extraction**  
   ```go
   headerRow := GetHeadersFromSheet(rawResultsSheet)
   ```
   The function expects at least one row; otherwise an error is returned.

2. **Find indices of required columns**  
   It calls `GetHeaderIndicesByColumnNames` with the set of column names that exist in every raw‑results sheet (`workloadNameRawResultsCol`, `operatorVersionRawResultsCol`, etc.). If any are missing, an error is produced.

3. **Build new header row**  
   ```go
   newHeaders := append(headerRow,
                        "Owner/TechLead Conclusion",
                        "Next Step Actions")
   ```
   Two new string literals are appended to the existing header slice.

4. **Create the new sheet**  
   The function constructs a new `*sheets.Sheet` with:
   * `Name`: `"SingleWorkloadResults"` (from the exported constant).
   * `Rows`: starting with the newly built header row.
   * `Options`: any options inherited from the original sheet.

5. **Return**  
   The new sheet is returned along with a `nil` error if all steps succeeded.

### Where It Fits in the Package
*`createSingleWorkloadRawResultsSheet`* is used by higher‑level orchestration code that:
1. Reads a raw results spreadsheet from Google Sheets.
2. Splits it into per‑workload sheets (using this helper).
3. Writes those new sheets back to the target Google Sheet for review.

Because the function is internal, it encapsulates all logic needed to preserve header integrity while extending the sheet with user‑action columns. It relies on constants defined in `const.go` for column names and sheet titles, ensuring consistency across the package.
