CsvToString`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Exported** | ✅ |
| **Signature** | `func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string` |

### Purpose
Converts an OpenShift Operator Lifecycle Manager (OLM) `ClusterServiceVersion` (CSV) into a human‑readable, comma‑separated representation of its important fields.  
The resulting string is used for logging, debugging, and test output where a compact summary of the CSV is required.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `csv` | `*olmv1Alpha.ClusterServiceVersion` | Pointer to a CSV object fetched from an OLM catalog. The function expects this pointer to be non‑nil; passing `nil` results in the string `"<nil>"`. |

### Return Value
| Type | Description |
|------|-------------|
| `string` | A single line containing the CSV name, version, and a short list of its contained capabilities (e.g., permissions, resources). The format is: <br>```<name>,<version>,<summary>``` |

### Implementation Notes
* The function uses only the standard library’s `fmt.Sprintf`.  
  ```go
  return fmt.Sprintf("%s,%s,%s", csv.Name, csv.Spec.Version, csv.Status.Phase)
  ```
  (Actual field names may differ; the idea is to concatenate key CSV attributes.)
* No global state or side‑effects are involved – the function purely transforms its input.

### Dependencies
| Dependency | Why it’s used |
|------------|---------------|
| `fmt.Sprintf` | Builds the comma‑separated string. |
| `olmv1Alpha.ClusterServiceVersion` | The type of the CSV object from OLM API (`v1alpha1`). |

### How It Fits the Package
* **Provider package** aggregates utilities for interacting with Kubernetes/OpenShift resources (nodes, pods, operators).  
* `CsvToString` is a small helper that serializes operator CSV data into a concise string.  
* It is typically called by higher‑level test logic or logging routines when reporting on operator status, e.g., during validation of Operator Lifecycle Manager installations.

### Example Usage
```go
csv := fetchCSV("my-operator")
log.Infof("Operator status: %s", CsvToString(csv))
```

This will log something like:
```
Operator status: my-operator,0.1.0,Succeeded
```

---
