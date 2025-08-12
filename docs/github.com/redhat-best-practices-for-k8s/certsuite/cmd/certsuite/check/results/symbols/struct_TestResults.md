TestResults`

> **Location**  
> `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/results` – `results.go:35`

## Purpose
`TestResults` is a lightweight wrapper that aggregates all test case results produced by the certsuite command line tool. It embeds the `TestCaseList` type, which holds the individual test cases and their outcomes. By exposing the embedded fields directly, consumers can treat a `TestResults` value as if it were a plain list of test cases while still keeping the semantic meaning that these are *results* rather than raw data.

## Composition

| Field | Type          | Notes |
|-------|---------------|-------|
| `embedded:TestCaseList` | `TestCaseList` | The only field; no additional metadata is stored here. |

The embedding means that all methods and fields of `TestCaseList` are promoted to `TestResults`. Typical usage:

```go
var results TestResults

// populate via unmarshalling or direct manipulation
results = append(results, TestCase{...})

// iterate over cases
for _, tc := range results {
    fmt.Println(tc.Status)
}
```

## Key Dependencies

| Dependency | Role |
|------------|------|
| `TestCaseList` (in the same package) | Provides the slice type and any helper methods for manipulating test cases. |

No external packages are referenced directly by this struct; it is purely a data container.

## Side Effects & Invariants

* **Immutability** – The struct itself has no exported setters or methods that modify its state beyond what `TestCaseList` offers. Consumers must be aware that adding/removing test cases changes the underlying slice.
* **No Validation** – There are no constructors or validation logic; any `TestResults` instance is considered valid as long as it contains zero or more `TestCase` entries.

## How It Fits the Package

The `results` package focuses on representing, serializing, and reporting test outcomes.  
- `TestCaseList` holds individual cases.  
- `TestResults` simply gives those cases a higher‑level semantic wrapper that can be used by other parts of the tool (e.g., output formatting, aggregation logic) without exposing the raw list type.

In short, `TestResults` is a thin façade over `TestCaseList`, enabling clearer intent and future extensibility (such as adding metadata or summary fields later).
