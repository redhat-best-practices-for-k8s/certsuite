NewClusterVersionReportObject`

| Item | Detail |
|------|--------|
| **Signature** | `func NewClusterVersionReportObject(version string, aReason string, isCompliant bool) *ReportObject` |
| **Exported?** | Yes (capitalized name) |
| **Location** | `pkg/testhelper/testhelper.go:269` |

## Purpose

Creates a `ReportObject` that represents the compliance status of a Kubernetes/OpenShift cluster’s version.  
The function is part of the test‑reporting framework used by CertSuite to generate structured compliance reports.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `version` | `string` | The actual cluster version string (e.g., `"4.12.0"`). |
| `aReason` | `string` | Human‑readable reason for the compliance state (e.g., `"Supported by policy"`). |
| `isCompliant` | `bool` | Indicates whether the cluster version meets the required criteria (`true` = compliant, `false` = non‑compliant). |

## Return Value

* `*ReportObject` – a pointer to a newly allocated `ReportObject`.  
  The returned object is fully populated with fields relevant to a *cluster version* report.

## Implementation Details

1. **Create Base Object**  
   Calls the helper `NewReportObject()` (also defined in this package) which returns an empty, but properly initialized, `ReportObject`.

2. **Populate Fields**  
   The function then calls `AddField` on the object twice:
   * `"reason"` – set to `aReason`.
   * `"compliant"` – set to a string representation of `isCompliant` (`"true"` or `"false"`).

3. **Return**  
   The fully populated report object is returned.

No external state is modified; the function operates purely on its arguments and returns a new value.

## Dependencies

| Called Function | Purpose |
|-----------------|---------|
| `NewReportObject()` | Constructs an empty `ReportObject`. |
| `AddField(string, string)` | Adds a key/value pair to the report. |

Both helpers are defined in the same package (`testhelper`), so no external imports are required.

## Package Context

* **Package**: `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`
* The `testhelper` package provides utilities for building test reports and defining compliance criteria.  
  `NewClusterVersionReportObject` is one of several constructors that create specialized report objects (e.g., for operators, resources, etc.).  
* This function is used by the CertSuite test suite when evaluating whether a cluster’s version satisfies policy requirements.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[User] -->|Calls| B(NewClusterVersionReportObject)
    B --> C{Create empty ReportObject}
    C --> D[NewReportObject()]
    D --> E[AddField("reason", aReason)]
    E --> F[AddField("compliant", isCompliantStr)]
    F --> G[Return *ReportObject]
```

This diagram illustrates the flow from invocation to final report object creation.
