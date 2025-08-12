getRoles`

**Location**

`pkg/autodiscover/autodiscover_rbac.go:52`

```go
func getRoles(rbacv1typed.RbacV1Interface) ([]rbacv1.Role, error)
```

### Purpose

Retrieves **all Role objects** that exist in the current Kubernetes cluster.  
The function is a thin wrapper around the client‑set’s RBAC v1 API and is used by the autodiscover logic to gather role information for later analysis (e.g., mapping permissions or generating diagnostics).

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `rbac` | `rbacv1typed.RbacV1Interface` | A typed Kubernetes client that exposes RBAC v1 methods. The caller typically passes the result of `kubernetes.NewForConfig(cfg).RbacV1()`.

### Returns

| Value | Type | Description |
|-------|------|-------------|
| `[]rbacv1.Role` | Slice of Role objects | All roles returned by the API call, or an empty slice if none exist. |
| `error` | Error | Non‑nil if the underlying client fails to list roles. The error is wrapped with contextual information (`"getRoles: %w"`).

### Implementation Details

```go
func getRoles(rbac rbacv1typed.RbacV1Interface) ([]rbacv1.Role, error) {
    // List all Roles in all namespaces (no namespace scope).
    roles, err := rbac.Roles("").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("getRoles: %w", err)
    }
    return roles.Items, nil
}
```

* The function calls `rbac.Roles("")` with an empty string to request cluster‑scoped listing (all namespaces).  
* It uses `context.TODO()` because no cancellation or timeout is needed for this short operation.  
* Errors from the API call are wrapped using `fmt.Errorf`, preserving the original error for debugging.

### Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `rbacv1typed.RbacV1Interface` | Provides access to RBAC resources. |
| `context.TODO()` | Context used for the API call (no cancellation). |
| `metav1.ListOptions{}` | Empty options → list all roles. |
| `fmt.Errorf` | Wraps errors with a clear message. |

No global variables are read or modified.

### How It Fits the Package

* The `autodiscover` package scans cluster resources to determine configuration and security posture.  
* `getRoles` is called by higher‑level discovery functions that need to inspect Role objects (e.g., mapping permissions to namespaces, checking for privileged roles).  
* By centralizing RBAC listing logic here, the rest of the code can remain agnostic about client construction and error handling.

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Caller] -->|rbacv1typed.RbacV1Interface| B[getRoles]
    B --> C{List Roles}
    C --> D[Return []rbacv1.Role, nil]
    C --> E[Error -> fmt.Errorf("getRoles: %w", err)]
```

This diagram shows the call path from a caller to `getRoles`, the internal list operation, and the two possible outcomes.
