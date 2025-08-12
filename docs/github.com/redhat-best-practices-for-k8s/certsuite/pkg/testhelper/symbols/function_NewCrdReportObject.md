NewCrdReportObject`

| Aspect | Detail |
|--------|--------|
| **Package** | `testhelper` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper) |
| **Exported** | Yes |
| **Signature** | `func NewCrdReportObject(name, version, reason string, compliant bool) *ReportObject` |

### Purpose
Creates a ready‑to‑use report object that represents the outcome of testing a Custom Resource Definition (CRD).  
The returned `*ReportObject` already contains the fields that downstream code expects for a CRD test:
- **Name** – the CRD name (`name`)
- **Version** – the API version (`version`)
- **Reason** – human‑readable explanation of compliance status (`reason`)
- **Compliant** – boolean flag indicating if the CRD passed all checks

This helper keeps the construction logic in one place so test writers can simply call it instead of manually populating a `ReportObject`.

### Parameters
| Name | Type | Meaning |
|------|------|---------|
| `name` | `string` | The fully‑qualified name of the CRD (e.g. `"mycrd.example.com"`). |
| `version` | `string` | API version of the CRD (`"v1alpha1"`, `"v1beta1"`, etc.). |
| `reason` | `string` | Text that explains why the test succeeded or failed. |
| `compliant` | `bool` | Indicates compliance: `true` for compliant, `false` otherwise. |

### Return Value
A pointer to a freshly allocated `ReportObject`.  
The object is fully populated with the provided fields and ready for inclusion in a test report.

### Key Dependencies & Side‑Effects
1. **`NewReportObject()`** – Used to create the base `ReportObject`. This function initializes all internal maps (fields, errors, etc.).
2. **`AddField()`** – Called three times to insert:
   - `"name"` → `name`
   - `"version"` → `version`
   - `"reason"` → `reason`
3. No global state is modified; the function is pure aside from creating a new struct instance.
4. The returned object may be mutated later by callers, but that mutation is outside this helper’s responsibility.

### Relationship to the Package
`NewCrdReportObject` sits among several “factory” helpers (e.g., `NewDeploymentReportObject`, `NewOperatorReportObject`) that standardise report construction for different Kubernetes resource types.  
It allows tests in the `pkg/testhelper` package to generate consistent CRD reports without repeating boilerplate code.

### Usage Example
```go
crdObj := testhelper.NewCrdReportObject(
    "mycrd.example.com",
    "v1alpha1",
    "CRD schema is valid and all required fields are present",
    true,
)
report.Add(crdObj)   // add to a larger report
```

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Call NewCrdReportObject] --> B(NewReportObject())
  B --> C{Create ReportObject}
  C --> D[AddField("name", name)]
  C --> E[AddField("version", version)]
  C --> F[AddField("reason", reason)]
  F --> G[Return *ReportObject]
```

This diagram illustrates the straightforward sequence of operations performed by `NewCrdReportObject`.
