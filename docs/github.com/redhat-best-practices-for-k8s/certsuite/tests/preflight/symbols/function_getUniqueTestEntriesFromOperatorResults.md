getUniqueTestEntriesFromOperatorResults`

*File:* `tests/preflight/suite.go` – line 243  
*Package:* `preflight`

## Purpose
Collects all pre‑flight tests that have been executed by a set of operators and returns them as a map keyed by the test’s unique identifier.  
The function is used internally during the execution of the pre‑flight suite to deduplicate results when multiple operators report the same test.

> **Why it matters** – The suite may run many operator instances in parallel; each instance reports its own `PreflightTest` objects.  `getUniqueTestEntriesFromOperatorResults` guarantees that only one copy of each test is kept, simplifying later reporting and validation steps.

## Signature
```go
func getUniqueTestEntriesFromOperatorResults(ops []*provider.Operator) map[string]provider.PreflightTest
```

| Parameter | Type                               | Description |
|-----------|------------------------------------|-------------|
| `ops`     | `[]*provider.Operator`             | Slice of pointers to `Operator`. Each operator exposes a field (`Tests`) that holds the tests it has executed. |

### Return value
- `map[string]provider.PreflightTest`:  
  *Key*: test ID string (unique per pre‑flight test).  
  *Value*: the corresponding `PreflightTest` struct instance.

## Dependencies & Key Calls

| Dependency | Role |
|------------|------|
| `make(map[string]provider.PreflightTest)` | Initializes the result map. |
| `operator.Tests` | Source of tests for each operator (not explicitly shown in the snippet but inferred from usage). |

No global variables or other functions are accessed, so the function is pure and side‑effect free.

## Algorithm Overview

```text
1. Create an empty map: uniqueTests.
2. For each operator in ops:
     a. Iterate over operator.Tests (slice of PreflightTest).
     b. Use test.ID as key; store the test value in uniqueTests.
        - If the same ID appears again, it simply overwrites the previous entry,
          ensuring uniqueness.
3. Return uniqueTests.
```

## How It Fits the Package

- **Package context** – The `preflight` package orchestrates testing of Kubernetes operators before they are promoted to production.  
- **Interaction with other parts** – After all operators finish, the suite calls this function to gather a deduplicated set of tests, which is then used by reporting utilities (e.g., generating JSON/YAML summaries) and by validation logic that checks for test failures.
- **Design rationale** – By keeping `getUniqueTestEntriesFromOperatorResults` local to the package and unexported, it signals that this helper is an implementation detail of the suite’s internal workflow.

## Summary

`getUniqueTestEntriesFromOperatorResults` aggregates pre‑flight tests from multiple operator instances into a unique map keyed by test ID. It performs no side effects, depends only on `make`, and is a core step in preparing results for reporting within the `preflight` testing framework.
