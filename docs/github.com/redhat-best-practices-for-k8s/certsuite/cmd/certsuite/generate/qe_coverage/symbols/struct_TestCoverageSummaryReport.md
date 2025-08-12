TestCoverageSummaryReport`

The **`TestCoverageSummaryReport`** type is the central data structure that holds a high‑level view of how many test cases in a CertSuite run have Quality Engineering (QE) coverage information, and what that coverage looks like.

| Field | Type | Meaning |
|-------|------|---------|
| `CoverageByTestSuite` | `map[string]TestSuiteQeCoverage` | Maps each test suite name to its per‑suite QE coverage statistics.  The value type (`TestSuiteQeCoverage`) is defined elsewhere in the package and contains counters such as *total tests*, *tests with QE data*, and *coverage percentages* for that suite. |
| `TestCasesTotal` | `int` | Total number of test cases discovered across all suites. |
| `TestCasesWithQe` | `int` | Number of those test cases that actually have QE metadata attached (i.e., they appear in the input map to `GetQeCoverage`). |
| `TotalCoveragePercentage` | `float32` | Overall coverage percentage, calculated as `(TestCasesWithQe / TestCasesTotal) * 100`. |

## How It Is Produced

The package function **`GetQeCoverage`** builds a `TestCoverageSummaryReport` from the map that maps a test case identifier to its QE description.  
During construction it:

1. Iterates over all entries in the input map.
2. Accumulates per‑suite counters into `CoverageByTestSuite`.
3. Updates `TestCasesTotal` and `TestCasesWithQe`.
4. Calculates `TotalCoveragePercentage` using a series of `float32()` casts.

The function is exported so callers can feed it raw QE data and obtain a ready‑to‑use summary.

## How It Is Consumed

The private helper **`showQeCoverageForTestCaseName`** prints a human‑readable representation of a single test case’s coverage.  
It receives:

- A string with the test case name.
- The `TestCoverageSummaryReport` that contains all suite statistics (though this function only uses the summary to format its output).

The helper writes several lines using `fmt.Println/Printf`, including:

- The test case name and whether it has QE data.
- The count of total tests versus those with QE for the relevant suite.
- A list of missing coverage entries.

This visual aid is useful when debugging or reviewing coverage reports during CI runs.

## Integration With the Package

`TestCoverageSummaryReport` sits at the top level of the **qecoverage** package.  
All public API that produces or consumes QE coverage summaries relies on this struct:

- **Input** – `GetQeCoverage` accepts a map keyed by `claim.Identifier`.
- **Output** – callers receive a fully populated `TestCoverageSummaryReport`, which they can further inspect, marshal to JSON, or feed into other reporting tools.

Because the struct contains only primitive types and maps of simple structs, it is trivially serializable (e.g., with `encoding/json`) and easy to pass between goroutines if needed.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    subgraph Input
        rawMap[map[claim.Identifier]claim.TestCaseDescription]
    end
    subgraph Process
        GetQeCoverage["GetQeCoverage()"]
        TCSR[TestCoverageSummaryReport]
    end
    subgraph Output
        summary[TCSR]
        print[showQeCoverageForTestCaseName()]
    end

    rawMap --> GetQeCoverage --> TCSR --> print
```

This diagram shows the flow from raw QE data to the final printed coverage report.
