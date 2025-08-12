testPreflightOperators`

```go
func testPreflightOperators(group *checksdb.ChecksGroup, env *provider.TestEnvironment)
```

## Purpose

`testPreflightOperators` is a helper used in the pre‑flight test suite to run a set of **operator** checks on a given Kubernetes environment and record their results.  
It:

1. Executes the operator tests that belong to `group`.
2. Aggregates the results into a single `PreflightResults` object.
3. Stores those results back into the database via `SetPreflightResults`.

The function is intentionally **unexported**; it is called from higher‑level test orchestration code in the same package.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `group` | `*checksdb.ChecksGroup` | The group of checks that should be executed.  Each check represents a pre‑flight operator test (e.g., verifying CNF‑cert tests). |
| `env` | `*provider.TestEnvironment` | A reference to the current test environment, containing the kube client and any configuration needed for executing the checks. |

## Workflow

```mermaid
flowchart TD
    A[Start] --> B{Run operator tests}
    B --> C[Collect results]
    C --> D[SetPreflightResults(group, results)]
    D --> E[Log summary]
```

1. **Running Operator Tests**  
   The function calls the internal helper `getUniqueTestEntriesFromOperatorResults` to transform raw operator test output into a slice of `checksdb.TestEntry`. This step removes duplicate entries and normalises data.

2. **Storing Results**  
   It passes the collected entries to `SetPreflightResults`, which writes them into the database under the provided check group. If this call fails, the test suite aborts with `t.Fatalf`.

3. **Logging**  
   The function logs the number of unique tests found and a summary of each result via `t.Info`. This provides visibility during test execution.

4. **Generating CNF‑cert Test**  
   Finally, for each operator test that is relevant to CNF certification, it generates a CNF‑cert test configuration by calling `generatePreflightOperatorCnfCertTest`. These configurations are then available for downstream tests.

## Dependencies

| Dependency | Role |
|------------|------|
| `SetPreflightResults` | Persists the aggregated results into the database. |
| `Fatal`, `Info` (from `testing.T`) | Test logging and failure handling. |
| `len` | Counts number of unique test entries for reporting. |
| `getUniqueTestEntriesFromOperatorResults` | Normalises operator test output. |
| `generatePreflightOperatorCnfCertTest` | Creates CNF‑cert specific test configurations from operator results. |

## Side Effects

* **Database Write** – `SetPreflightResults` mutates the checks database by adding or updating entries for the supplied group.
* **Logging** – Emits informational messages to the test harness; if an error occurs, it terminates the current test with a fatal log.

No other global state is mutated. The function relies on the passed-in `group` and `env`; any changes are local to these objects.

## Package Context

The `preflight` package orchestrates all pre‑flight tests for CertSuite.  
This helper sits in `suite.go`, alongside:

* `beforeEachFn` – a setup hook executed before each test.
* `env` – the shared `TestEnvironment`.

Together, they form the backbone of the pre‑flight test runner that validates operator behavior and CNF certification compliance.
