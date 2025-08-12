NewPodReportObject`

| Item | Details |
|------|---------|
| **Package** | `testhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`) |
| **Signature** | `func NewPodReportObject(namespace, podName, reason string, compliant bool) *ReportObject` |
| **Exported** | ✅ |

### Purpose
Creates a `ReportObject` that represents the result of a compliance check performed on a single Kubernetes Pod.  
The returned object is already populated with the core identifying fields (`Namespace`, `PodName`) and the compliance verdict (`ReasonForCompliance/NonCompliance`, `Compliant`).  

This helper is used by test suites to generate consistent, structured output for each pod under evaluation.

### Parameters
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `namespace` | `string` | Namespace where the Pod resides. |
| `podName`    | `string` | Name of the Pod being evaluated. |
| `reason`     | `string` | Human‑readable explanation of why the pod is compliant or not. |
| `compliant`  | `bool`   | `true` if the pod meets all checks; otherwise `false`. |

### Return Value
* `*ReportObject` – a pointer to an instance that contains:
  * Core fields (`Namespace`, `PodName`)
  * The `ReasonForCompliance` or `ReasonForNonCompliance` field set depending on `compliant`
  * The `Compliant` boolean

### Implementation Flow
```go
func NewPodReportObject(namespace, podName, reason string, compliant bool) *ReportObject {
    // 1. Create a base ReportObject for the Pod.
    obj := NewReportObject(PodType, namespace, podName)

    // 2. Add the compliance reason field.
    if compliant {
        obj.AddField(ReasonForCompliance, reason)
    } else {
        obj.AddField(ReasonForNonCompliance, reason)
    }

    // 3. Return the fully populated object.
    return obj
}
```
* `NewReportObject` (internal helper) initializes a `ReportObject` with the type (`PodType`) and basic metadata.  
* `AddField` appends additional key/value pairs to the object’s internal map.

### Side Effects & Dependencies
| Dependency | Role |
|------------|------|
| `NewReportObject` | Constructs the base object; does not modify global state. |
| `AddField` | Mutates the returned `ReportObject`; no external side effects. |

No global variables are accessed or mutated within this function.

### Package Context
* The `testhelper` package provides a suite of utilities for generating structured test reports in CertSuite.
* `NewPodReportObject` is part of the “report creation” helpers, mirroring similar constructors for other resource types (`Deployment`, `Service`, etc.).
* It integrates with the report‑generation workflow used by the CLI and CI pipelines.

### Usage Example
```go
podReport := testhelper.NewPodReportObject(
    "default",
    "nginx-xyz",
    "All containers use non‑root user",
    true,
)
```

The resulting `podReport` can then be serialized to JSON/YAML or fed into further analysis tooling.
