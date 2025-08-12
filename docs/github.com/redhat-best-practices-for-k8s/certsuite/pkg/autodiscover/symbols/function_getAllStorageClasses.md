getAllStorageClasses`

### Purpose
Retrieves all **Kubernetes StorageClass** objects that are available in the cluster.  
The function is used by the autodiscovery logic to discover PV/PVC related resources and later feed them into the test suite.

### Signature

```go
func getAllStorageClasses(storagev1typed.StorageV1Interface) ([]storagev1.StorageClass, error)
```

| Parameter | Type                                 | Description |
|-----------|--------------------------------------|-------------|
| `client`  | `storagev1typed.StorageV1Interface` | A typed client that provides access to the Storage API (e.g. obtained via `kubernetes.NewForConfig`). |

| Return | Type                        | Description |
|--------|-----------------------------|-------------|
| `[]storagev1.StorageClass` | Slice of all StorageClasses found in the cluster. |
| `error` | Non‑nil if the list operation fails or any unexpected error occurs. |

### Implementation Details

```go
func getAllStorageClasses(storageClient storagev1typed.StorageV1Interface) ([]storagev1.StorageClass, error) {
    // List all StorageClasses (namespace‑agnostic).
    scList, err := storageClient.StorageClasses().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("error listing storage classes: %w", err)
    }

    // The API returns a *storagev1.StorageClassList; convert it to a slice.
    var scs []storagev1.StorageClass
    for _, sc := range scList.Items {
        scs = append(scs, sc)
    }
    return scs, nil
}
```

* **API call** – `StorageClasses().List()` is the standard way to fetch all storage classes.  
* **Context** – a background context (`context.TODO()`) is used; callers cannot cancel it because this function is short‑lived.  
* **Error handling** – any error from the API is wrapped with a message and returned.

### Dependencies

| Dependency | Role |
|------------|------|
| `storagev1typed.StorageV1Interface` | Provides the typed client for storage operations. |
| `context.TODO()` | Supplies a context for the API call. |
| `metav1.ListOptions{}` | Empty options, meaning “list everything”. |
| `fmt.Errorf` | Wraps errors with contextual information. |

### Side Effects

* **No side effects** – the function is read‑only; it does not modify any resources or package state.

### Package Context

The `autodiscover` package orchestrates detection of cluster resources (nodes, pods, PV/PVC, network policies, etc.) to build a test environment.  
`getAllStorageClasses` fits into this pipeline by supplying the list of StorageClass objects that may influence how persistent volumes are provisioned or selected during tests.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Caller] -->|Passes client| B[getAllStorageClasses]
    B --> C{List storage classes}
    C --> D[storagev1.StorageClassList]
    D --> E[Return []StorageClass, nil]
    C --> F{Error?} -->|Yes| G[Return nil, error]
```

This diagram visualizes the function’s role in the autodiscovery workflow.
