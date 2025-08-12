getClusterRoleBindings`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover` |
| **Visibility** | Unexported (private to the package) |
| **Signature** | `func getClusterRoleBindings(rbac rbacv1typed.RbacV1Interface) ([]rbacv1.ClusterRoleBinding, error)` |

### Purpose
Retrieves *all* `ClusterRoleBinding` resources that exist in the target Kubernetes cluster.  
The function is used by higher‑level discovery logic to inspect which cluster‑wide permissions are granted and subsequently decide whether certain certificates or secrets should be managed.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `rbac` | `rbacv1typed.RbacV1Interface` | A typed client that provides access to the RBAC API group. The caller typically obtains this from a shared Kubernetes client (`clientset.RbacV1()`).

### Return Values
| Index | Type | Meaning |
|-------|------|---------|
| 0 | `[]rbacv1.ClusterRoleBinding` | Slice containing every cluster‑role binding returned by the API server. |
| 1 | `error` | Non‑nil if the List call fails (e.g., network issue, permission denied). In that case the slice is empty.

### Key Steps
1. **API Call** – Uses `rbac.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})`.  
   * `context.TODO()` indicates no special cancellation logic is required here; the call blocks until completion or error.
2. **Error Handling** – If the List request fails, it logs the error (`log.Error(err)`) and returns an empty slice with that error.

### Dependencies
- **k8s.io/client-go/kubernetes/typed/rbac/v1**: provides `RbacV1Interface` and `ClusterRoleBindingList`.
- **metav1.ListOptions**: default options (empty filter).
- **context.TODO**: placeholder context.
- **logrus/log** (via `log.Error`): side‑effect for diagnostics.

### Side Effects
- No mutation of global state; only performs a read from the API server.
- Emits an error log if the API call fails, which may surface in application logs.

### Relationship to the Package
`autodiscover` orchestrates discovery of various Kubernetes objects (CRDs, RBAC resources, etc.) to inform certificate management decisions.  
`getClusterRoleBindings` is a helper that encapsulates the standard pattern for listing cluster‑role bindings; other parts of the package call it when they need to inspect or filter role‑binding information.

---

#### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Caller] -->|rbac client| B[getClusterRoleBindings]
    B --> C{List API}
    C --> D[Success: []ClusterRoleBinding]
    C --> E[Error: log & return error]
```

This diagram visualizes the call flow from the caller to the RBAC List operation and the two possible outcomes.
