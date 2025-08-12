NewDeploymentReportObject`

| Item | Description |
|------|-------------|
| **Signature** | `func NewDeploymentReportObject(namespace, deploymentName, reason string, compliant bool) *ReportObject` |
| **Exported** | Yes (`testhelper.NewDeploymentReportObject`) |
| **Location** | `pkg/testhelper/testhelper.go:333` |

### Purpose

Creates a fully‑populated `ReportObject` that represents the compliance status of a Kubernetes Deployment.  
The function is used by tests and reporting utilities to generate a single, reusable object containing:

* The namespace and name of the deployment under test.
* A human‑readable reason for (non)compliance.
* The overall compliance flag (`true` = compliant, `false` = non‑compliant).

### Parameters

| Parameter | Type   | Role |
|-----------|--------|------|
| `namespace` | `string` | Kubernetes namespace of the deployment. |
| `deploymentName` | `string` | Name of the Deployment resource. |
| `reason` | `string` | Explanation for the compliance status (e.g., “Missing liveness probe”). |
| `compliant` | `bool` | Flag indicating whether the deployment meets all checks. |

### Return Value

* `*ReportObject` – a pointer to a newly allocated `ReportObject`.  
  The object is guaranteed to be non‑nil and contains fields set via `AddField`.

### Key Steps & Dependencies

1. **Instantiate base report**  
   ```go
   ro := NewReportObject()
   ```
   * Calls the package‑level constructor that returns an empty `ReportObject` ready for field insertion.

2. **Attach deployment metadata**  
   Two calls to `AddField` populate:
   - `Namespace` (key: `"namespace"`, value from parameter).
   - `DeploymentName` (key: `"deployment_name"`, value from parameter).

3. **Record compliance data**  
   ```go
   ro.AddField(ReasonForCompliance, reason)
   ro.AddField("compliant", compliant)
   ```
   * `ReasonForCompliance` is a package constant string key (`"reason_for_compliance"`).
   * The `"compliant"` field uses the boolean value directly.

4. **Return** – the fully populated object is returned to the caller.

### Side Effects

* No global state is modified; all operations are on the newly created `ReportObject`.
* The function only performs memory allocations and string assignments.

### Package Context

`testhelper` provides a collection of helper functions and constants for generating structured compliance reports used in certsuite tests.  
`NewDeploymentReportObject` is one of several *factory* helpers that create report objects for specific Kubernetes resource types (Deployments, Pods, Services, etc.).  These helpers standardize field names via exported string constants defined elsewhere in the file.

### Suggested Diagram

```mermaid
flowchart TD
    A[Caller] --> B{NewDeploymentReportObject}
    B --> C[NewReportObject()]
    C --> D[AddField("namespace", namespace)]
    D --> E[AddField("deployment_name", deploymentName)]
    E --> F[AddField(ReasonForCompliance, reason)]
    F --> G[AddField("compliant", compliant)]
    G --> H[Return *ReportObject]
```

This diagram shows the linear creation and population of the `ReportObject` before it is returned.
