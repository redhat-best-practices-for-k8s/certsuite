PrintResultsTable`

| Aspect | Detail |
|--------|--------|
| **Location** | `github.com/redhat-best-practices-for-k8s/certsuite/internal/cli` – file `cli.go`, line 147 |
| **Signature** | `func PrintResultsTable(results map[string][]int) func()` |
| **Exported** | Yes (public API of the `cli` package) |

### Purpose
`PrintResultsTable` produces a *summary table* that displays, for each test suite, how many checks were:

- Passed (`CheckResultTagPass`)
- Failed (`CheckResultTagFail`)
- Skipped (`CheckResultTagSkip`)
- Running (`CheckResultTagRunning`)
- Aborted (`CheckResultTagAborted`)
- Encountered an error (`CheckResultTagError`)

The function is intended to be used **after** all tests have finished executing. It takes the aggregated result counts per suite and returns a *function* that, when called, prints the table to `stdout`. This design allows callers to defer printing until the end of the test run or embed it in a larger report.

### Parameters
| Name | Type | Meaning |
|------|------|---------|
| `results` | `map[string][]int` | Keys are suite names; values are slices of integers whose positions correspond to the six result tags above. The slice length is expected to be 6, one element per tag in the order shown in the constants list. |

### Return Value
A **zero‑argument function** (`func()`) that performs the actual printing when invoked. This returned closure captures `results` by value.

### Key Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| Standard library `fmt.Printf` / `Println` | Used for formatted console output. No external packages are imported. |
| Constants (`CheckResultTag*`) | Provide readable labels for each column in the table. |
| Global variables | None are accessed or modified by this function; it is pure apart from writing to stdout. |

### How It Works
1. **Header** – Prints a header line with the suite name and the six result tags, color‑coded using ANSI escape codes defined in the constants (`Red`, `Green`, etc.).
2. **Rows** – Iterates over each key/value pair in `results`:
   - Retrieves counts for each tag.
   - Formats them into a single row aligned under the header columns.
3. **Footer** – Prints an empty line to separate the table from subsequent output.

Because it only writes to standard output, calling the returned function has no side effects on program state beyond console visibility.

### Usage Pattern
```go
// After collecting results in `suiteResults`
defer PrintResultsTable(suiteResults)()
```

This defers printing until after all other deferred actions run, ensuring the table appears last in the CLI output.

### Diagram (optional)

```mermaid
graph LR
  A[Collect test results] --> B{Map<string, []int>}
  B --> C[PrintResultsTable(results)]
  C --> D[(Closure)]
  D --> E[fmt.Printf/Println]
```

The diagram shows the flow from result collection to the closure that performs printing.

--- 

**Summary:**  
`PrintResultsTable` is a lightweight helper that formats and prints a colored summary table of test outcomes for each suite. It accepts a pre‑aggregated map, returns a closure for deferred execution, and relies solely on standard library formatting functions and predefined ANSI color constants.
