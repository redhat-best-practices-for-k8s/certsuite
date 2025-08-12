TcResultsSummary` – Test‑Case Result Aggregation

| Element | Description |
|---------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/testcases` |
| **File**   | `testcases.go:10` |

### Purpose
`TcResultsSummary` is a lightweight container that holds the total counts of test‑case outcomes for a single comparison run. It is used by the *claim* subsystem to report how many tests passed, failed, or were skipped when evaluating a claim against a set of certificate checks.

### Fields

| Field   | Type | Meaning |
|---------|------|---------|
| `Passed`  | `int` | Number of test cases that succeeded. |
| `Failed`  | `int` | Number of test cases that failed. |
| `Skipped` | `int` | Number of test cases that were not executed (e.g., due to missing data). |

### Typical Usage Flow

```go
// A map produced by the comparison engine: tcName -> "pass"/"fail"/"skip"
results := compareEngine.Run(...)

// Convert raw results into a summary struct
summary := getTestCasesResultsSummary(results)
```

1. **Result Map** – `compareEngine` returns a `map[string]string` where each key is a test‑case name and the value is one of `"pass"`, `"fail"`, or `"skip"`.
2. **Conversion** – The helper function `getTestCasesResultsSummary` iterates over this map, increments the appropriate field in a new `TcResultsSummary`, and returns it.
3. **Reporting** – The returned struct can be serialized to JSON, logged, or included in higher‑level claim reports.

### Key Dependencies

- **Helper Function**: `getTestCasesResultsSummary(map[string]string) TcResultsSummary` (located at line 70).  
  This function encapsulates the logic for counting results; no other external packages are required.
- **No Global State** – The struct is self‑contained and has no side effects.

### Side Effects
None. `TcResultsSummary` simply holds data; all manipulation occurs in pure functions that return new instances.

### Diagram (optional)

```mermaid
graph LR
    compareEngine --> resultsMap[map[string]string]
    resultsMap -->|passed/failed/skipped| getTestCasesResultsSummary
    getTestCasesResultsSummary --> summary[TcResultsSummary]
```

### Summary
`TcResultsSummary` is a small, immutable snapshot of test outcomes used throughout the claim comparison logic to provide concise metrics on certificate compliance checks. It serves as the bridge between raw per‑test results and human‑readable reports.
