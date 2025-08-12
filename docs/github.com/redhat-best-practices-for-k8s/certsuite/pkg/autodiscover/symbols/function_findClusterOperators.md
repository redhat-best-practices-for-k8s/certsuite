findClusterOperators`

### Purpose
`findClusterOperators` retrieves the list of **ClusterOperator** objects that are installed on a target Kubernetes cluster.  
The function is used by the autodiscover logic to determine which operators are present so that downstream components can decide whether a particular operator’s configuration or status should be inspected.

> **Why this matters** – many features in Cert‑Suite depend on knowing whether an operator such as *OpenShift Cluster Version* or *Istio* is available.  This helper abstracts the API call and error handling, keeping higher‑level code clean.

### Signature
```go
func findClusterOperators(clientconfigv1.ClusterOperatorInterface) ([]configv1.ClusterOperator, error)
```

| Parameter | Type                                    | Description |
|-----------|----------------------------------------|-------------|
| `client`  | `clientconfigv1.ClusterOperatorInterface` | Kubernetes client that can list ClusterOperator resources. |

| Return | Type                                 | Description |
|--------|--------------------------------------|-------------|
| `[]configv1.ClusterOperator` | slice of all ClusterOperators found on the cluster. |
| `error` | any error encountered while listing; if the resource is not found, a *not‑found* sentinel may be returned by the caller. |

### Key Dependencies
- **Kubernetes client-go** – the interface is part of the OpenShift API (`clientconfigv1.ClusterOperatorInterface`).  
- **k8s.io/api/config/v1** – `configv1.ClusterOperator` type used in the return slice.  
- Logging utilities (`Debug`) from the package’s logger to record failures.

### Implementation Notes
```go
func findClusterOperators(client clientconfigv1.ClusterOperatorInterface) ([]configv1.ClusterOperator, error) {
    // Attempt to list all ClusterOperator objects.
    coList, err := client.List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        // If the API is missing (e.g., non‑OpenShift cluster), we treat it as a normal “no operators” case.
        if errors.IsNotFound(err) {
            log.Debug("ClusterOperator resource not found; likely not an OpenShift cluster")
            return nil, nil
        }
        return nil, err
    }

    // Convert the List object to a plain slice of ClusterOperators.
    var co []configv1.ClusterOperator
    for _, item := range coList.Items {
        co = append(co, *item.DeepCopy())
    }
    return co, nil
}
```

* The function uses `client.List` with an empty `ListOptions`.  
* It distinguishes a “resource not found” error from other failures: non‑OpenShift clusters will simply yield an empty list.  
* All other errors are propagated to the caller.

### Side Effects & Error Handling
- **No side effects** on the cluster; only reads data.  
- Logs a debug message when the resource is absent, aiding troubleshooting in mixed environments.  
- Returns `nil` slice with no error if the API endpoint does not exist (common for vanilla Kubernetes).

### Package Context
Within `autodiscover`, this helper sits under *ClusterOperator* discovery logic.  Other functions such as `findIstioOperators` or `findSRIoVResources` rely on a similar pattern but target different CRDs.  By centralising the list‑and‑error handling here, the package keeps its higher‑level autodiscover routines concise and focused on interpreting the operator data rather than plumbing the API.

---

**Mermaid diagram (suggested)**

```mermaid
flowchart TD
    A[Caller] -->|client| B(findClusterOperators)
    B --> C[List Kubernetes]
    C --> D{Success?}
    D -- yes --> E[Return []ClusterOperator]
    D -- no --> F{IsNotFound?}
    F -- yes --> G[Log debug, return nil,nil]
    F -- no --> H[Return error]
```

This visual helps developers understand the control flow and error handling strategy of `findClusterOperators`.
