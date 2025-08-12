getRoleBindings`

```go
func getRoleBindings(rbacv1typed.RbacV1Interface) ([]rbacv1.RoleBinding, error)
```

| Aspect | Description |
|--------|-------------|
| **Purpose** | Retrieve *all* `RoleBinding` objects that exist in the Kubernetes cluster. The function is used by the autodiscover logic to understand which subjects (users, groups, service accounts) have what permissions, and to subsequently decide whether a given pod or operator should be monitored for TLS traffic. |
| **Inputs** | `rbacv1typed.RbacV1Interface` – an interface that exposes the standard RBAC client set from the Kubernetes API (`client-go`). The caller typically passes `k8sClientset.RbacV1()` which is a typed client capable of calling methods such as `RoleBindings().List`. |
| **Outputs** | - A slice of `rbacv1.RoleBinding` objects. <br>- An error if the list operation fails or any other internal issue occurs. |
| **Key Dependencies** | * `rbacv1typed.RbacV1Interface` (from `k8s.io/client-go/kubernetes/typed/rbac/v1`) <br>* The underlying `List` method on the client’s `RoleBindings()` interface.<br>* Standard Kubernetes error handling (`Error`). |
| **Side‑Effects** | None. The function performs a read operation only; it does not modify any cluster state or local data structures. |
| **Integration into Package** | <p>The autodiscover package orchestrates the discovery of resources (operators, pods, networking objects) to determine what certificates need rotation. `getRoleBindings` is called during the initial collection phase when the package builds a global view of RBAC rules that might affect how TLS is injected or intercepted. By returning every role binding, downstream logic can inspect bindings for specific annotations, labels, or subjects and adjust its behavior accordingly.</p> |

### Typical Usage Flow

1. **Client Setup** – A caller creates an `rbacv1typed.RbacV1Interface` via the Kubernetes client set.
2. **Call `getRoleBindings`** – The function returns all role bindings; errors are propagated to the caller.
3. **Post‑processing** – Other parts of autodiscover iterate over the slice, looking for particular patterns (e.g., role binding names that reference an operator or a service account used by certsuite).

### Mermaid Diagram (suggested)

```mermaid
flowchart TD
    A[Caller] --> B[getRoleBindings(rbacv1typed.RbacV1Interface)]
    B --> C{Error?}
    C -- No --> D[Return []rbacv1.RoleBinding]
    C -- Yes --> E[Propagate error]
```

> **Note**: The function currently contains a `TODO` comment indicating that future improvements might add filtering or pagination. As it stands, it performs a simple list operation without any namespace scoping or label selectors.
