DiffReport` – Summary of Test‑Case Differences  

The **`DiffReport`** struct is the central data structure used by the
*compare* command in CertSuite to expose a human‑readable summary of how two
claim files differ.

| Field | Type | Meaning |
|-------|------|---------|
| `Claim1ResultsSummary` | `TcResultsSummary` | Aggregated counts (passed, skipped, failed) for the first claim file. |
| `Claim2ResultsSummary` | `TcResultsSummary` | Same aggregation for the second claim file. |
| `DifferentTestCasesResults` | `int` | Number of test cases that appear in both claims but have differing results or are missing from one side. |
| `TestCases` | `[]TcResultDifference` | Detailed per‑test‑case information: name, result in each claim and an indicator of difference. |

> **Purpose**  
> The struct is produced by `GetDiffReport`, which merges two
> `claim.TestSuiteResults` objects (one from each claim file) into a single,
> consumable object.  It is then rendered to the user via its `String()`
> method, producing two tables: a summary table and a detailed diff table.

## Key Functions that Touch `DiffReport`

| Function | Role |
|----------|------|
| **`GetDiffReport(claim.TestSuiteResults, claim.TestSuiteResults)`** | Builds a `DiffReport` from two result sets. It calls helper functions to create maps of test‑case results, merge the names, and compute per‑summary statistics. |
| **`DiffReport.String()`** | Implements the `fmt.Stringer` interface.  Formats the summary tables as plain text using `Sprintf`. No side effects beyond string construction. |

### How `GetDiffReport` Works

1. **Map Creation** – `getTestCasesResultsMap` turns each claim’s raw results into a map keyed by test‑case name for O(1) look‑ups.
2. **Name Merging** – `getMergedTestCasesNames` returns the union of all names from both claims, ensuring every case is considered.
3. **Per‑Case Analysis** – For each merged name, the helper compares results; if they differ or one is missing, it increments `DifferentTestCasesResults` and appends a `TcResultDifference` to `TestCases`.
4. **Summary Calculation** – `getTestCasesResultsSummary` aggregates pass/skipped/failed counts separately for each claim.

The resulting `DiffReport` contains all information needed by the CLI output
layer or any downstream tooling that may want to programmatically consume
the diff data.

### Rendering with `String()`

- Builds a two‑section string:
  - **Test Cases Summary Table** – columns: status, count in CLAIM‑1, count in CLAIM‑2.
  - **Different Test Cases Table** – columns: test‑case name, result in each claim.
- Uses repeated `fmt.Sprintf` calls; the output is deterministic and
  free of external state.

## Dependencies & Side Effects

| Dependency | Notes |
|------------|-------|
| `claim.TestSuiteResults`, `TcResultsSummary`, `TcResultDifference` | Defined elsewhere in the same package; they carry raw test‑case data. |
| Standard library (`fmt`) | Only for string formatting; no I/O or global state changes. |

**Side effects:** None beyond returning a fully populated struct and generating
a formatted string.  All operations are pure.

## Place in the Package

`DiffReport` sits at the heart of the *compare* command:
- **Input:** Two `claim.TestSuiteResults` (one per claim file).
- **Processing:** `GetDiffReport` merges, compares, and summarizes.
- **Output:** A string representation for CLI display or a struct that can be serialized.

```mermaid
graph TD;
    ClaimA[Claim File A] -->|results| GetDiffReport;
    ClaimB[Claim File B] -->|results| GetDiffReport;
    GetDiffReport --> DiffReport;
    DiffReport --> String();
```

This concise representation makes it straightforward for developers to
understand how test‑case differences are computed and presented.
