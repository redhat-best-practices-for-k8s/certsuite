FindStatefulsetByNameByNamespace`

### Purpose
Retrieves a single Kubernetes **StatefulSet** object identified by its *name* and the *namespace* it resides in.  
The function is part of the `autodiscover` package, which provides helper utilities for discovering Kubernetes resources needed by CertSuite (e.g., to locate operator pods or network plugins). This lookup is used when the test harness needs to inspect the state of a StatefulSet that backs an application component.

### Signature
```go
func FindStatefulsetByNameByNamespace(
    client appv1client.AppsV1Interface,
    name string,
    namespace string,
) (*appsv1.StatefulSet, error)
```

| Parameter | Type                 | Description                               |
|-----------|----------------------|-------------------------------------------|
| `client`  | `AppsV1Interface`    | Kubernetes client for the *apps/v1* API group. |
| `name`    | `string`             | The name of the StatefulSet to fetch.     |
| `namespace` | `string`           | Namespace where the StatefulSet is expected. |

### Return Values
- `*appsv1.StatefulSet`: Pointer to the retrieved StatefulSet object. Returns `nil` if not found or on error.
- `error`:  
  * `nil` when lookup succeeds.  
  * Any error returned by the underlying client call, wrapped with context using `Error()`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `client.StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})` | Performs the actual API call to fetch the StatefulSet. |
| `TODO` (comment placeholder) | Indicates that additional logic may be inserted in the future (e.g., context handling or retries). |
| `Error()` from the package’s logging helpers | Wraps raw errors with a human‑readable message. |

### Side Effects
- No mutation of global state; purely read‑only.
- Performs an API call to the Kubernetes control plane; network I/O is involved.

### How It Fits in the Package
`autodiscover` contains several *finder* utilities that abstract away direct client calls for common resources (Deployments, DaemonSets, Pods, etc.).  
`FindStatefulsetByNameByNamespace` follows this pattern, enabling other parts of CertSuite to:

1. Locate a StatefulSet by name/namespace without duplicating client logic.
2. Handle errors in a consistent way across the package.

The function is exported (`FindStatefulsetByNameByNamespace`) so that it can be reused in tests or other packages that need to query StatefulSets during discovery or verification steps.
