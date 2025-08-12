GetTotalTests` – Overview

| Item | Detail |
|------|--------|
| **Package** | `checksdb` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb`) |
| **Signature** | `func GetTotalTests() int` |
| **Exported?** | Yes |
| **Purpose** | Return the current number of test groups stored in the checks database. |

## Functionality

```go
// GetTotalTests returns how many check groups are currently registered.
func GetTotalTests() int {
    return len(dbByGroup)
}
```

The function simply counts the entries in the package‑wide map `dbByGroup`.  
`dbByGroup` is defined as:

```go
var dbByGroup = make(map[string]*ChecksGroup)
```

Each key represents a group name (e.g., “network”, “storage”) and maps to a pointer to a `ChecksGroup`, which contains the individual checks belonging to that group. Therefore, the return value of `GetTotalTests` is **the number of distinct groups**, not the total count of individual checks.

> **Note**: If the caller expects the sum of all checks across all groups, this function would need to iterate over each `ChecksGroup` and aggregate their lengths.

## Dependencies

| Dependency | Role |
|------------|------|
| `len` (built‑in) | Computes the number of keys in `dbByGroup`. |
| `dbByGroup` (global map) | Holds all registered check groups. |

No external packages or complex logic are involved, so there are no side effects beyond reading a shared data structure.

## Usage Context

- **Metrics & Reporting** – The function is used by higher‑level reporting tools to display how many test suites are available in the current configuration.
- **Health Checks** – Some diagnostics may verify that at least one group exists; `GetTotalTests` provides an easy way to perform this check.

## Diagram (optional)

```mermaid
graph TD;
    A[Caller] --> B{GetTotalTests()};
    B --> C[len(dbByGroup)];
    C --> D[Number of keys];
```

This diagram illustrates that the caller obtains a count by delegating to Go’s `len` on the global map.

---

**Summary:**  
`GetTotalTests` is a lightweight accessor that returns the number of registered check groups in the `checksdb` package. It relies only on the global `dbByGroup` map and the built‑in `len` function, producing no side effects.
