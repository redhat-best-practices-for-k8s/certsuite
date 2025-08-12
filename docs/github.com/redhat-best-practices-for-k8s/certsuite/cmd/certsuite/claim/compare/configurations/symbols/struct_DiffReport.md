DiffReport` – Summary of Configuration Differences

| Element | Description |
|---------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/configurations` |
| **Purpose** | Encapsulates the result of comparing two Kubernetes configuration objects (`*claim.Configurations`). It stores a diff object and a count of abnormal events that were detected during comparison. |

### Fields

| Field | Type | Notes |
|-------|------|-------|
| `Config` | `*diff.Diffs` | Holds the raw diffs produced by the external **diff** package (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/diff`). This is a structured representation of all changes (additions, deletions, modifications) between two configuration trees. |
| `AbnormalEvents` | `AbnormalEventsCount` | A custom type that tracks how many abnormal events were found while comparing the configurations. The type is defined elsewhere in the package and typically contains counters for different event categories (e.g., missing fields, unexpected values). |

> **Why both?**  
> `Config` gives a low‑level view of *what* changed; `AbnormalEvents` offers a high‑level summary useful for reporting or decision logic.

### Methods

#### `func (dr DiffReport) String() string`

- **Exported** – can be used by callers to obtain a human‑readable representation.
- **Implementation**: Delegates to the `String()` methods of its embedded fields (`Config` and `AbnormalEvents`).  
  - It first calls `dr.Config.String()` to serialize the diff tree.  
  - Then it appends the string from `dr.AbnormalEvents.String()`.  
- **Side effects**: None; pure value conversion.

### Construction

The primary constructor in this package is:

```go
func GetDiffReport(old, new *claim.Configurations) *DiffReport
```

1. Calls `Compare(old, new)` – a helper that walks the two configuration trees and returns a `*diff.Diffs` object.  
2. Counts abnormal events by inspecting the returned diff (via `len()` on relevant slices).  
3. Returns a fully populated `DiffReport`.

> **Dependencies**  
> - `compare.go` (provides `Compare`).  
> - `diff` package for diff data structures.  
> - `AbnormalEventsCount` type defined in the same package.

### Usage Pattern

```go
oldCfg := loadConfig(...)
newCfg := loadConfig(...)

report := configurations.GetDiffReport(oldCfg, newCfg)
fmt.Println(report.String()) // human‑friendly output
if report.AbnormalEvents.HasCritical() {
    // take remedial action
}
```

The struct serves as the bridge between raw diff data and higher‑level decision logic in CertSuite’s claim comparison tooling. It is intentionally read‑only after creation, ensuring that callers cannot mutate internal state inadvertently.
