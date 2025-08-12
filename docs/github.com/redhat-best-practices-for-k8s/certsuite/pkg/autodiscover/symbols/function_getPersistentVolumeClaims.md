getPersistentVolumeClaims`

| Aspect | Details |
|--------|---------|
| **Signature** | `func (client corev1client.CoreV1Interface) ([]corev1.PersistentVolumeClaim, error)` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover` |
| **Exported?** | No – it is an internal helper used by the autodiscover package. |

### Purpose
Retrieve all PersistentVolumeClaims (PVCs) that exist in the Kubernetes cluster for which a `CoreV1Interface` client can list resources. The function is called from higher‑level discovery logic to gather information about storage usage, which may influence certificate placement or workload profiling.

### Parameters
| Name | Type | Role |
|------|------|------|
| `client` | `corev1client.CoreV1Interface` | A typed Kubernetes client that exposes the `CoreV1()` API group. It must be initialized with a valid kubeconfig and context; otherwise, calls to `List` will fail. |

### Return Values
| Index | Type | Meaning |
|-------|------|---------|
| 0 | `[]corev1.PersistentVolumeClaim` | Slice of PVC objects retrieved from the cluster. If no PVCs exist or an error occurs, this slice may be empty. |
| 1 | `error` | Non‑nil if the API call fails (e.g., network issue, permission denied). The caller is expected to handle or surface this error upstream. |

### Key Dependencies
* **Kubernetes Client** – Uses `client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})`.  
  * The empty string (`""`) denotes the “all namespaces” scope.
* **Context** – The function internally creates a background context (`context.Background()`). No external cancellation is considered.  
* **TODO placeholder** – There is an intentional `TODO` comment in the source, indicating that future enhancements (e.g., filtering or pagination) may be added.

### Side Effects
* None beyond API calls; no state is mutated.
* The function performs a network request to the Kubernetes API server each time it is invoked.

### Integration into the Package
The autodiscover package orchestrates discovery of various cluster resources (CSV, operator deployments, PVCs, etc.) to build a runtime view. `getPersistentVolumeClaims` is one step in that pipeline:
1. Called by higher‑level discovery logic (`discoverClusterResources` or similar).
2. The returned PVC list feeds into certificate recommendation algorithms that may consider storage classes or node affinity.

### Suggested Diagram
```mermaid
graph LR
  A[Client Call] --> B[getPersistentVolumeClaims]
  B --> C[List API call to /api/v1/persistentvolumeclaims]
  C --> D[Return []PVC, error]
```

---

**Note:** As the function is unexported and contains a `TODO`, its current behavior is minimal. Future revisions may add filtering by namespace or labels, pagination handling, or integration with custom resource discovery logic.
