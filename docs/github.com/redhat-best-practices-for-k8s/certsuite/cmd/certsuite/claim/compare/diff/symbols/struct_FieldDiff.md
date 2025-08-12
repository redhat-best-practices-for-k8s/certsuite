FieldDiff` – Data structure for representing a single field difference

| Element | Description |
|---------|-------------|
| **Package** | `diff` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff`) |
| **Purpose** | Holds the minimal information required to describe that a particular field in two claim files differs. It is used as an element of the slice returned by the package’s `Compare` function, allowing callers to inspect which fields differ and what their values were. |

## Fields

| Field | Type | Meaning |
|-------|------|---------|
| `FieldPath` | `string` | Dot‑separated path that uniquely identifies the field within the claim structure (e.g., `"metadata.name"`). The same string is used for both claims, making it easy to correlate the values. |
| `Claim1Value` | `interface{}` | Value extracted from the first claim file at `FieldPath`. It can be any JSON‑serialisable type: primitive, array, map, etc. |
| `Claim2Value` | `interface{}` | Value extracted from the second claim file at the same path. |

> **Note** – The struct is *pure data*: it has no methods and does not perform any logic. All behaviour comes from the surrounding package that constructs and consumes these structs.

## How It Is Used

1. **Diff Calculation**  
   The `diff` package walks both claim objects recursively, comparing each field. Whenever a mismatch is found, a new `FieldDiff` instance is created:

   ```go
   diff := FieldDiff{
       FieldPath:   currentPath,
       Claim1Value: valFromClaim1,
       Claim2Value: valFromClaim2,
   }
   diffs = append(diffs, diff)
   ```

2. **Result Aggregation**  
   The slice of `FieldDiff` values is returned to the caller (e.g., a CLI or API endpoint). Callers can then format, filter, or count differences.

3. **Reporting / Logging**  
   Because it contains only the path and raw values, it can be marshalled into JSON for machine‑readable reports or printed in human‑friendly tables.

## Dependencies & Side Effects

- **Dependencies:** None – `FieldDiff` uses only built‑in types (`string`, `interface{}`).
- **Side effects:** None – it is an immutable data holder. The only side effect comes from the code that creates instances (which may allocate memory).

## Placement in the Package

The `diff` package’s public API typically exposes a function like:

```go
func Compare(claim1, claim2 interface{}) ([]FieldDiff, error)
```

`FieldDiff` is the return type used to convey each difference. It sits at the top level of the package and can be imported by other packages (e.g., a CLI or test harness) without pulling in any heavy dependencies.

---

### Mermaid diagram (optional)

```mermaid
classDiagram
    class FieldDiff {
        +string FieldPath
        +interface{} Claim1Value
        +interface{} Claim2Value
    }

    class CompareResult {
        +[]FieldDiff Diffs
    }

    CompareResult --> FieldDiff : contains
```

This diagram illustrates that a comparison result aggregates many `FieldDiff` instances, each describing one differing field.
