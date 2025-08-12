getResourceQuotas`

| Aspect | Detail |
|--------|--------|
| **File** | `autodiscover_resources.go` (line 27) |
| **Visibility** | Unexported (`private`) – used only within the `autodiscover` package. |
| **Signature** | `func getResourceQuotas(corev1client.CoreV1Interface) ([]corev1.ResourceQuota, error)` |

### Purpose
Collect all `ResourceQuota` objects that exist in the Kubernetes cluster reachable through the supplied Core‑V1 client.  
A *resource quota* is a namespaced object that limits aggregate resource consumption (CPU, memory, etc.) for workloads within that namespace.

The function returns:
- A slice of `corev1.ResourceQuota` instances – one per discovered object.
- An error if the list operation fails.

### Inputs
| Parameter | Type | Meaning |
|-----------|------|---------|
| `corev1client.CoreV1Interface` | Kubernetes Core‑V1 client interface | Allows the function to perform API calls (`ResourceQuotas().List`) against the cluster. The caller typically passes a real client created from a kubeconfig, but any implementation of this interface (e.g., a fake client for tests) works.

### Output
| Return value | Type | Meaning |
|--------------|------|---------|
| `[]corev1.ResourceQuota` | Slice of ResourceQuota objects | All resource quotas that the API server returns. |
| `error` | Error or nil | Non‑nil if the list request fails (network issue, permission error, etc.). |

### Key Dependencies
- **Kubernetes Go client** (`k8s.io/api/core/v1`, `k8s.io/client-go/kubernetes/typed/core/v1`).  
  The function uses:
  - `ResourceQuotas(namespace string)` – to obtain a namespaced resource quota interface.
  - `List(ctx, opts)` – to fetch the list of quotas.  
- **Context** is implicitly used via the default background context in the client call (the client internally creates it).

### Side Effects
The function only performs a read‑only API call; it does not modify any cluster state or local data structures.

### Package Context
`autodiscover` is responsible for detecting various Kubernetes resources that may influence certificate generation and deployment.  
- `getResourceQuotas` feeds the rest of the package with information about namespace limits, which can affect how many pods or services the system can safely create.
- The function is invoked by higher‑level discovery logic (e.g., `discoverClusterResources`) to aggregate resource constraints across the cluster.

### Usage Flow
```go
clientset := kubernetes.NewForConfig(cfg)
quotas, err := autodiscover.getResourceQuotas(clientset.CoreV1())
if err != nil {
    // handle error
}
// quotas now holds all ResourceQuota objects in the cluster
```

> **Note**: The implementation currently contains a `TODO` comment, indicating that future enhancements (e.g., filtering by namespace or handling specific quota names) may be added. As it stands, it returns *all* resource quotas without discrimination.

---

#### Suggested Mermaid Diagram (Resource Quota Discovery Flow)

```mermaid
flowchart TD
    A[Caller] -->|Pass CoreV1Interface| B[getResourceQuotas]
    B --> C{List Request}
    C -->|Success| D[Return []corev1.ResourceQuota]
    C -->|Failure| E[Return error]
```

This diagram illustrates the single‑step data flow: a caller provides a client, `getResourceQuotas` performs an API list call, and returns either the results or an error.
