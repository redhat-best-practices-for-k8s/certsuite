getPVCFromSlice`

```go
func getPVCFromSlice([]corev1.PersistentVolumeClaim, string) *corev1.PersistentVolumeClaim
```

| Element | Description |
|---------|-------------|
| **Purpose** | Locate a specific PersistentVolumeClaim (PVC) in an in‑memory slice and return a pointer to it.  Used by the volume lifecycle tests to inspect or manipulate PVCs that have been created during test setup. |
| **Parameters** | 1. `pvcList []corev1.PersistentVolumeClaim` – The slice of PVC objects obtained from the Kubernetes API (e.g., via `client.List`). <br>2. `name string` – The name of the PVC to find. |
| **Return value** | A pointer to the matching `*corev1.PersistentVolumeClaim`. If no PVC with the given name exists, `nil` is returned. |
| **Key dependencies** | * `k8s.io/api/core/v1`: Core Kubernetes types (`PersistentVolumeClaim`).<br>* No external packages beyond the core API are imported; the function is purely a linear search. |
| **Side effects** | None – the function only reads from its arguments and returns a reference to an element within the slice. It does not modify the slice or any global state. |
| **How it fits in the package** | The `volumes` test package exercises lifecycle operations on PVCs (creation, binding, deletion).  After performing API calls that return lists of PVCs, tests need a quick way to access an individual object by name.  `getPVCFromSlice` provides this helper without exposing any internal representation details or requiring repeated code across multiple test files. |

### Typical usage in tests

```go
pvcList, _ := client.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
targetPVC := getPVCFromSlice(pvcList.Items, "my-test-pvc")
require.NotNil(t, targetPVC)
```

### Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Call `getPVCFromSlice(pvcList, name)`] --> B{Iterate over pvcList}
    B -->|found?| C[Return pointer to matching PVC]
    B -->|not found| D[Return nil]
```

> **Note**: The function is unexported (`private`) and intended solely for internal test logic. It does not perform any error handling beyond the simple lookup, delegating responsibility to callers if they need richer diagnostics.
