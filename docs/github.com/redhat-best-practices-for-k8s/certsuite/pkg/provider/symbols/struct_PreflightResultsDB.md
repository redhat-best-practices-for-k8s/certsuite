PreflightResultsDB`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **File / Line** | `provider.go:192` |
| **Exported** | ✅ (public API) |

### Purpose
`PreflightResultsDB` is a lightweight container that aggregates the outcome of all *pre‑flight* checks executed by CertSuite.  A pre‑flight check is a single test that validates a Kubernetes cluster’s configuration or behaviour before a workload is deployed.  The struct holds three slices, each containing `PreflightTest` values representing tests that ended with different statuses:

| Slice | Meaning |
|-------|---------|
| `Errors` | Tests that panicked or produced an unexpected error during execution. |
| `Failed` | Tests that executed successfully but returned a non‑zero status code (i.e., the cluster failed the test). |
| `Passed` | Tests that succeeded and returned zero exit status. |

These slices are populated by `GetPreflightResultsDB`, which is invoked by `Container.SetPreflightResults`.  The resulting map of container name → `PreflightResultsDB` is stored on the `Container` instance for later reporting.

### Key Dependencies
- **`PreflightTest`** – The type stored in each slice; it holds a test’s name, metadata, help text, and optionally an error.  
- **`GetPreflightResultsDB(*plibRuntime.Results)`** – Reads the raw results from `plbRuntime.Results`, extracts all pre‑flight tests, classifies them into the three slices above, and returns the populated struct.  
- **`Container.SetPreflightResults(map[string]PreflightResultsDB, *TestEnvironment) error`** – Calls `GetPreflightResultsDB` for each container’s results map and writes a human‑readable report to the test environment.

### Input / Output
| Function | Inputs | Outputs |
|----------|--------|---------|
| `GetPreflightResultsDB(*plibRuntime.Results)` | A pointer to a `plbRuntime.Results` object that contains all test run data. | A fully populated `PreflightResultsDB`. |
| `Container.SetPreflightResults(map[string]PreflightResultsDB, *TestEnvironment) error` | - `map[string]PreflightResultsDB`: pre‑flight results per container.<br>- `*TestEnvironment`: context for logging/reporting. | Writes the report; returns an error if writing fails. |

### Side Effects
- **Logging** – `SetPreflightResults` emits informational logs (via `Info`) about the number of tests in each category.
- **Report Generation** – The function writes a formatted pre‑flight results table into the test environment’s output stream.

### How It Fits the Package

```mermaid
graph TD;
    Container-->SetPreflightResults;
    SetPreflightResults-->|calls|GetPreflightResultsDB;
    GetPreflightResultsDB-->|returns|PreflightResultsDB;
    PreflightResultsDB-->|contains|PreflightTest[];
```

- The `provider` package orchestrates test execution across multiple containers.  
- After a container finishes running its pre‑flight tests, the runtime produces a `plbRuntime.Results`.  
- `GetPreflightResultsDB` converts this raw data into an easy‑to‑consume struct (`PreflightResultsDB`).  
- The container then hands that struct to `SetPreflightResults`, which logs and reports it.

### Usage Example

```go
// Assume we already have a map of container results.
resultsMap := map[string]PreflightResultsDB{
    "container1": GetPreflightResultsDB(container1Runtime),
    // ...
}

env := NewTestEnvironment()
if err := container.SetPreflightResults(resultsMap, env); err != nil {
    log.Fatalf("failed to set preflight results: %v", err)
}
```

### Summary
`PreflightResultsDB` is a minimal yet essential data holder that aggregates pre‑flight test outcomes into three categories.  It bridges the raw runtime output (`plbRuntime.Results`) and the human‑readable report produced by `Container.SetPreflightResults`, enabling clear visibility of cluster readiness before workloads are deployed.
