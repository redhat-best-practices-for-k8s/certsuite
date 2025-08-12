NewClusterOperatorReportObject`

### Purpose
Creates a *cluster‑operator* section of the test report.

A cluster operator is a higher‑level component that manages one or more Kubernetes resources (CRDs, Deployments, etc.).  
The function produces a `*ReportObject` that represents a single cluster‑operator instance in the final JSON/YAML report.  
It sets common metadata fields and registers an empty list of compliance checks for this operator.

### Signature
```go
func NewClusterOperatorReportObject(name string, namespace string, compliant bool) *ReportObject
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | `string` | The name of the cluster‑operator. |
| `namespace` | `string` | The Kubernetes namespace in which the operator runs. |
| `compliant` | `bool` | Indicates whether the operator is expected to be compliant (`true`) or not (`false`). |

### Return value
- `*ReportObject`: a pointer to a fully‑initialized report object that can later be enriched with checks, resources, and metrics.

### Key Operations

1. **Instantiate a generic report object**  
   Calls `NewReportObject(name, namespace, compliant)` which creates a base object containing the provided metadata and an empty `Checks` slice.

2. **Add operator‑specific fields**  
   Uses `AddField(report, key, value)` to insert two attributes:
   - `Type`: set to `"ClusterOperator"` – this distinguishes it from other resource types in the report.
   - `Name`: a duplicate of the supplied name for quick lookup.

3. **Return** the fully populated object.

### Dependencies

| Dependency | Role |
|------------|------|
| `NewReportObject` | Builds the foundational structure (`ID`, `Namespace`, `Compliant`) that all report objects share. |
| `AddField` | Attaches custom key/value pairs to a report object. |

### Side Effects & Constraints
- **No I/O** – purely in‑memory construction.
- The returned object is immutable from the caller’s perspective; only methods on `ReportObject` can modify it further.
- The function assumes that the caller provides valid Kubernetes names and namespaces.

### Integration in the Package

The `testhelper` package generates test reports for various Kubernetes components.  
`NewClusterOperatorReportObject` is part of a family of constructor helpers (e.g., `NewDeploymentReportObject`, `NewServiceReportObject`) that standardise how each component type is represented.  
These objects are later aggregated into the final report structure (`ReportRoot`) and serialized to JSON/YAML for consumption by external tooling or dashboards.

---

**Mermaid diagram (optional)**

```mermaid
flowchart TD
  A[Call NewClusterOperatorReportObject] --> B{Create Base}
  B --> C(NewReportObject)
  C --> D[Set ID, Namespace, Compliant]
  D --> E[AddField(Type="ClusterOperator")]
  E --> F[AddField(Name=name)]
  F --> G[Return *ReportObject]
```
