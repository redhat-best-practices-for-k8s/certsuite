GetDiffReport` – High‑level Diff Engine

| Item | Detail |
|------|--------|
| **Package** | `testcases` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/testcases`) |
| **Exported?** | Yes (`GetDiffReport`) |
| **Signature** | `func GetDiffReport(a, b claim.TestSuiteResults) *DiffReport` |
| **Purpose** | Produce a side‑by‑side comparison of two test‑suite result sets. It aggregates per‑test‑case outcomes into a tabular report that highlights successes, failures and missing entries. |

## Core Behaviour

1. **Map each suite to a lookup table**  
   - Calls `getTestCasesResultsMap` on both input suites (`a`, `b`).  
   - The helper returns `map[string]claim.TestCaseResult` where the key is the test case name.

2. **Build a merged set of all test‑case names**  
   - Uses `getMergedTestCasesNames(mapA, mapB)` to collect every unique test‑case name from both maps.  
   - Guarantees that each row in the resulting report contains an entry for every test case present in either suite.

3. **Populate the diff rows**  
   - For each merged name, a new `DiffRow` is appended to the report’s table:  
     ```go
     append(&report.Table.Rows, DiffRow{
         Name: name,
         Left: mapA[name], // may be zero value → “not found”
         Right: mapB[name],
     })
     ```
   - The row fields are of type `claim.TestCaseResult`. If a test case is absent in one suite, its field will hold the zero value; later formatting functions interpret this as *“not found”*.

4. **Summarise each side**  
   - After all rows are added, `getTestCasesResultsSummary` is called twice: once for `mapA`, once for `mapB`.  
   - The summaries count successes, failures and missing cases and are stored in the report’s `LeftSummary` / `RightSummary`.

5. **Return**  
   - A fully populated `*DiffReport` containing table rows and side summaries.

## Dependencies

| Function | Role |
|----------|------|
| `getTestCasesResultsMap` | Turns a `claim.TestSuiteResults` into a name→result map. |
| `getMergedTestCasesNames` | Computes the union of keys from two maps. |
| `append` (built‑in) | Adds a row to the report’s table slice. |
| `getTestCasesResultsSummary` | Generates per‑suite success/failure statistics. |

All helpers are internal to the same package and operate purely on data; no external I/O or side effects occur.

## Side Effects & Mutability

- The function **does not modify** its input parameters (`a`, `b`).  
- It creates a new `DiffReport` instance; callers receive ownership of this result.  
- No global state is accessed or altered.

## Integration in the Package

`GetDiffReport` is the public API that drives the comparison view presented by the CLI.  
Higher‑level code (e.g., command handlers) calls it after parsing two claim files, then passes the returned `*DiffReport` to a renderer that prints a table with colour coding.

```mermaid
graph LR
A[CLI parse] --> B{Two claim files}
B --> C[GetDiffReport(a,b)]
C --> D[Render table]
```

The function is central: it bridges raw test‑suite data and the user‑facing diff representation.
