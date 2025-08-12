GetPodTopOwner`

| | |
|---|---|
| **Package** | `podhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper`) |
| **Exported** | Yes |

### Purpose
`GetPodTopOwner` walks the Kubernetes owner reference chain for a set of pods and returns, for each pod, the *top‑level* owning object (e.g., Deployment, StatefulSet, DaemonSet).  
The result is a map keyed by pod name with a `TopOwner` value that contains the kind, name, UID and namespace of that top owner.

### Signature
```go
func GetPodTopOwner(podName string, owners []metav1.OwnerReference) (map[string]TopOwner, error)
```

| Parameter | Type | Description |
|---|---|---|
| `podName` | `string` | The name of the pod for which we want to resolve its owner. |
| `owners` | `[]metav1.OwnerReference` | Direct owner references attached to the pod (e.g., from `metadata.ownerReferences`). |

| Return | Type | Description |
|---|---|---|
| `map[string]TopOwner` | map keyed by pod name | For each pod, the resolved top‑level owner. |
| `error` | error | Non‑nil if any step of the resolution fails (e.g., API call errors). |

### Key Dependencies

| Dependency | Role |
|---|---|
| `make(map[string]TopOwner)` | Initializes the result map. |
| `followOwnerReferences` | Recursively walks an owner chain until a top‑level object is reached. |
| `GetClientsHolder()` | Provides a Kubernetes clientset used by `followOwnerReferences`. |
| `Errorf` | Wraps and returns errors from underlying calls. |

### High‑level Flow

```mermaid
flowchart TD
    A[Start: podName + owners] --> B{Owners empty?}
    B -- Yes --> C[Return empty map]
    B -- No --> D[Call followOwnerReferences(podName, owners)]
    D --> E{Success?}
    E -- Yes --> F[Store TopOwner in result map]
    E -- No --> G[Return error via Errorf]
```

1. **Input validation** – If the `owners` slice is empty, an empty map is returned immediately (no owner to resolve).  
2. **Resolution** – Calls `followOwnerReferences`, which uses a Kubernetes client (obtained from `GetClientsHolder`) to walk owner references until it reaches an object that has no owner (the top level).  
3. **Result assembly** – The resolved `TopOwner` is stored in the map under the original pod name.  
4. **Error handling** – Any error from the resolution step is wrapped with `Errorf` and returned.

### Side Effects & Notes

* The function performs read‑only API calls; it does not modify any cluster objects.  
* It depends on a correctly configured Kubernetes client (via `GetClientsHolder`).  
* If multiple pods are processed by repeated calls, the overhead of creating a new map each time is negligible compared to the API round‑trips.

### How It Fits the Package

`podhelper` provides utilities for introspecting pod ownership.  
`GetPodTopOwner` is the entry point used by higher‑level diagnostics or reporting tools that need to associate pods with their controlling controllers (Deployments, Jobs, etc.).  It encapsulates the owner‑reference traversal logic and presents a clean map of pod → top‑owner for callers to consume.
