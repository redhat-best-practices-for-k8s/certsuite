ResultObjectsToString`

**Location**

`github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper/testhelper.go:711`

```go
func ResultObjectsToString(
    objects []*ReportObject,
    nonCompliantObjects []*ReportObject,
) (string, error)
```

### Purpose

Converts two slices of `*ReportObject` into a single JSON string that can be embedded in a test report.  
The function is used by the test harness to serialize compliance results for output or further processing.

- **`objects`** – slice containing *compliant* result objects.  
- **`nonCompliantObjects`** – slice containing *non‑compliant* result objects.

Both slices are marshaled into a JSON object with two fields:
```json
{
  "objects": [...],
  "non_compliant_objects": [...]
}
```
The function returns the JSON string or an error if marshalling fails.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `objects` | `[]*ReportObject` | Compliant results. |
| `nonCompliantObjects` | `[]*ReportObject` | Non‑compliant results. |

> **Note**: `ReportObject` is defined elsewhere in the package; it represents a single test result.

### Return Values

| Name | Type | Description |
|------|------|-------------|
| `string` | The JSON representation of both slices. |
| `error` | Non‑nil if `json.Marshal` fails. |

### Key Dependencies & Calls

- **`encoding/json.Marshal`** – serializes the combined structure into JSON.
- **`fmt.Errorf`** – wraps any marshalling error with a descriptive message.
- **`string` conversion** – converts the byte slice returned by `Marshal` to a Go string.

No external packages beyond the standard library are used.  
The function does not modify its inputs; it is purely functional.

### Side Effects

- None (pure function).  
- The only observable effect is the returned JSON string or error.

### Integration in the Package

`ResultObjectsToString` lives in the `testhelper` package, which provides utilities for running and reporting compliance tests.  
Typical usage flow:

1. Test logic collects compliant and non‑compliant `ReportObject`s.
2. These slices are passed to `ResultObjectsToString`.
3. The resulting JSON string is logged, written to a file, or embedded in an API response.

This function centralizes the serialization logic so that all test outputs have a consistent format.  

### Example

```go
compliant := []*ReportObject{...}
nonCompliant := []*ReportObject{...}

jsonStr, err := ResultObjectsToString(compliant, nonCompliant)
if err != nil {
    log.Fatalf("failed to serialize results: %v", err)
}
fmt.Println(jsonStr)
```

### Summary

`ResultObjectsToString` is a small, well‑defined helper that packages compliance results into JSON for downstream consumption. It relies only on standard library calls, produces no side effects, and fits cleanly into the `testhelper` package’s responsibility of preparing test output.
