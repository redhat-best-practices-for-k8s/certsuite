GetQeCoverage` – Package **qecoverage**

| Item | Detail |
|------|--------|
| **File** | `qe_coverage.go` (line 78) |
| **Signature** | `func GetQeCoverage(map[claim.Identifier]claim.TestCaseDescription) TestCoverageSummaryReport` |
| **Exported** | ✅ |

---

## Purpose
`GetQeCoverage` aggregates quantitative coverage data from a set of test case descriptions.  
The function consumes a map that associates each claim identifier with its corresponding `TestCaseDescription`. From this, it calculates a summary report (`TestCoverageSummaryReport`) containing the total number of tests, passed tests, failed tests, and the overall pass‑rate expressed as a percentage.

---

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `testCases` | `map[claim.Identifier]claim.TestCaseDescription` | A lookup table where each key is a unique claim identifier. The value holds metadata about a single test case (e.g., name, status, etc.). |

---

## Return Value

| Type | Description |
|------|-------------|
| `TestCoverageSummaryReport` | A struct that holds aggregated coverage metrics: total tests, passed count, failed count, and pass‑rate. The exact fields are defined in the same package (see *Types* section below). |

---

## Key Operations
1. **Iterate** over all entries in `testCases`.  
2. For each entry, inspect its status (passed/failed) via a field inside `TestCaseDescription` (likely `Status`).  
3. Increment counters accordingly.  
4. After the loop, compute pass‑rate as:
   ```go
   passRate := float32(passed) / float32(total) * multiplier
   ```
   The `multiplier` constant (defined at line 14) is used to convert a fraction into a percentage value.  
5. Populate and return a `TestCoverageSummaryReport`.

---

## Dependencies

| Dependency | Role |
|------------|------|
| `claim.Identifier` & `claim.TestCaseDescription` | Input types, part of the *claims* package. |
| `multiplier` (constant) | Multiplies fraction by 100 to produce a percentage; defined in the same file. |
| Standard library functions: `append`, `float32` | Used for building slices and numeric conversion during calculation. |

---

## Side Effects & Mutability
- **Read‑only** – The function does not modify its input map or any global state.
- Returns a new report value; no external side effects.

---

## Usage Context

`GetQeCoverage` is invoked by the CLI command `qeCoverageReportCmd` (declared in the same file).  
After the coverage data is generated, the command serialises the returned `TestCoverageSummaryReport` to JSON/YAML for downstream consumption or display. This function therefore serves as the core logic behind the *“Generate QE Coverage Report”* feature of CertSuite.

---

## Types Referenced (inferred)

```go
type TestCoverageSummaryReport struct {
    TotalTests   int     `json:"total_tests"`
    Passed       int     `json:"passed"`
    Failed       int     `json:"failed"`
    PassRatePct  float32 `json:"pass_rate_pct"` // e.g., 95.3
}
```

*The actual struct may contain additional fields such as timestamps or detailed per‑claim breakdowns.*

---

## Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[CLI: qeCoverageReportCmd] --> B{Collect test cases}
    B --> C[GetQeCoverage(testCases)]
    C --> D[TestCoverageSummaryReport]
    D --> E[Output JSON/YAML]
```

---

**Bottom line:**  
`GetQeCoverage` is a pure, functional helper that turns raw test case metadata into a concise coverage summary. It operates on a map of claim identifiers, counts successes/failures, computes a pass‑rate percentage using the `multiplier` constant, and returns a structured report used by the CLI output command.
