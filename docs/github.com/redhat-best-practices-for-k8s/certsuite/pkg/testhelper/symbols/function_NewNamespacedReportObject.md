NewNamespacedReportObject`

```go
func NewNamespacedReportObject(reason string, typ string, compliant bool, namespace string) *ReportObject
```

### Purpose  
Creates a new **`ReportObject`** that represents the result of a test or check performed in a specific Kubernetes namespace.  
It builds on the generic `NewReportObject` helper and then injects the namespace into the report so downstream consumers can filter or group results by namespace.

---

### Parameters  

| Name      | Type   | Description |
|-----------|--------|-------------|
| `reason`  | `string` | Human‑readable explanation of why the test was performed (e.g., `"Image not signed"`). |
| `typ`     | `string` | Category/type identifier for the report (often a constant from this package such as `ImageName`, `PodType`, etc.). |
| `compliant` | `bool`  | Indicates whether the target passed (`true`) or failed (`false`). |
| `namespace` | `string` | The Kubernetes namespace to attach to the report. |

---

### Return Value  

* `*ReportObject` – a pointer to the newly constructed report object, already populated with all common fields plus an additional `"Namespace"` field.

---

### Key Dependencies

| Function Called | Role |
|-----------------|------|
| `NewReportObject(reason, typ, compliant)` | Generates the base report object. |
| `AddField(fieldName, value)` | Adds the `"Namespace"` key/value pair to that object. |

Both functions are defined in the same package (`testhelper`) and operate purely on data structures; no external state is modified.

---

### Side‑Effects & Constraints  

* **No global state mutation** – only local variables and the returned object are affected.
* The function expects a valid namespace string; an empty string will still be stored but may lead to ambiguous results downstream.
* Errors are not returned; any internal error would surface as a panic from `NewReportObject` or `AddField`.

---

### Usage Example

```go
// Create a report for a pod that failed in the "dev" namespace
r := NewNamespacedReportObject(
    "Pod missing required label",
    testhelper.PodType,
    false,
    "dev",
)

// r now contains fields like Reason, Type, Compliant, and Namespace.
```

---

### Relationship to the Package  

`NewNamespacedReportObject` is a convenience wrapper used throughout the `testhelper` package when tests need to report results that are scoped to a particular namespace. It keeps the public API simple: callers provide only the domain‑specific data (reason, type, compliance, and namespace) while the helper handles object construction and field injection consistently across all test cases.
