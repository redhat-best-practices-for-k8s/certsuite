NewHelmChartReportObject`

| Item | Details |
|------|---------|
| **Signature** | `func(namespace string, chartName string, reason string, compliant bool) *ReportObject` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper` |

### Purpose
Creates a ready‑to‑use `ReportObject` that represents the result of a Helm‑chart related test.  
The object is pre‑populated with the fields required by the reporting infrastructure:

* The **namespace** in which the chart was deployed.
* The **name** of the Helm chart under test.
* A textual **reason** explaining why the test passed or failed.
* A boolean **compliance flag** indicating success (`true`) or failure (`false`).

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `namespace` | `string` | Kubernetes namespace where the chart is installed. |
| `chartName` | `string` | Identifier of the Helm chart being evaluated. |
| `reason`    | `string` | Human‑readable explanation for the test outcome. |
| `compliant` | `bool`   | Test result: `true` = compliant, `false` = non‑compliant. |

### Output
* A pointer to a freshly allocated `ReportObject`.  
  The returned object has:
  * `Type` set to `"Helm"` (via the constant `HelmType`).
  * `ReasonForCompliance` or `ReasonForNonCompliance` populated with `reason`.
  * `Compliant` flag set accordingly.

### Key Dependencies
1. **`NewReportObject(namespace, chartName)`** – creates a base `ReportObject` and sets its namespace and name.
2. **`AddField(fieldName, value)`** – appends custom fields to the report:
   * Adds `"HelmChart"` field with the chart name.
   * Adds `"ReasonForCompliance"` or `"ReasonForNonCompliance"` depending on `compliant`.

No other global variables or types are touched.

### Side‑Effects
* None beyond creating and returning a new struct instance.  
  The function does not modify any package‑level state or external resources.

### Package Context
`testhelper` supplies utilities for constructing test reports that later feed into CertSuite’s compliance engine.  
Helm charts form one of the supported resource types, so `NewHelmChartReportObject` provides a convenient entry point for tests that validate chart installation and configuration.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Call NewHelmChartReportObject] --> B{Create base ReportObject}
  B --> C[Set Namespace & Name]
  A --> D{Add Helm‑specific fields}
  D --> E[AddField("HelmChart", chartName)]
  D --> F{compliant?}
  F -- true --> G[AddField("ReasonForCompliance", reason)]
  F -- false --> H[AddField("ReasonForNonCompliance", reason)]
  A --> I[Return *ReportObject]
```

This diagram illustrates the sequence of operations performed by `NewHelmChartReportObject`.
