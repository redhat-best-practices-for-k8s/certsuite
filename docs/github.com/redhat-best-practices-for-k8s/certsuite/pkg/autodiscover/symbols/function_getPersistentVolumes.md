getPersistentVolumes`

| Item | Details |
|------|---------|
| **Signature** | `func(corev1client.CoreV1Interface) ([]corev1.PersistentVolume, error)` |
| **Package** | `autodiscover` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover) |

### Purpose
Retrieves all PersistentVolumes (PVs) visible to the supplied Kubernetes client.  
The function is used by autodiscovery logic that needs to examine PV properties (e.g., storage class, labels) when generating certificates or configuration for workloads that consume persistent storage.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `client` | `corev1client.CoreV1Interface` | A Kubernetes client interface capable of performing Core v1 operations. It is typically created from a kubeconfig and used to talk to the cluster's API server.

### Return Values
| Index | Type | Description |
|-------|------|-------------|
| 0 | `[]corev1.PersistentVolume` | Slice containing all PV objects returned by the API call. If an error occurs, this slice will be nil. |
| 1 | `error` | Non‑nil if the list operation fails (e.g., network issue, permission denied). A nil value indicates success.

### Implementation Details
```go
func getPersistentVolumes(client corev1client.CoreV1Interface) ([]corev1.PersistentVolume, error) {
    // List all PersistentVolumes in the cluster.
    pvList, err := client.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("cannot list persistent volumes: %w", err)
    }

    // The TODO comment indicates future enhancement (e.g., filtering).
    // Currently the function simply returns the raw list.
    return pvList.Items, nil
}
```

- **API call** – `client.PersistentVolumes().List(...)` uses the CoreV1 client to fetch all PV objects.  
- **Context** – `context.TODO()` is used; callers may want to supply a cancellable context in future revisions.  
- **Error handling** – Errors from the API are wrapped with a descriptive message and returned.

### Dependencies
| Dependency | Role |
|------------|------|
| `corev1client.CoreV1Interface` | Provides access to Core v1 resources (here, PersistentVolumes). |
| `context.TODO()` | Placeholder context for the request. |
| `metav1.ListOptions{}` | Empty options; could be extended with label/field selectors later. |

### Side‑Effects
- **No state mutation** – The function only reads from the cluster; it does not modify any resources.
- **Network traffic** – Executes a GET request to `/api/v1/persistentvolumes` on the Kubernetes API server.

### How It Fits the Package
The `autodiscover` package gathers information about a running Kubernetes environment to determine which workloads and resources require certificates or configuration changes.  
`getPersistentVolumes` is one of several helper functions that collect cluster data (e.g., deployments, services, storage classes). The returned PV slice can be consumed by higher‑level logic that:

1. Identifies PVCs bound to these PVs.
2. Checks PV annotations/labels for compliance or configuration hints.
3. Generates or updates certificate resources based on PV characteristics.

By keeping this function simple and read‑only, the package ensures reliable discovery while allowing future enhancements (e.g., filtering by storage class) without breaking existing consumers.
