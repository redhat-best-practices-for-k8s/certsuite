GetAllOperatorGroups`

> **Location**: `pkg/provider/operators.go` (line 350)  
> **Signature**: `func GetAllOperatorGroups() ([]*olmv1.OperatorGroup, error)`  

### Purpose
Collects every *OperatorGroup* object that exists in the Kubernetes cluster.  
An OperatorGroup is a CRD from the OpenShift‑Managed‑Lifecycle‑Manager (OLM) API; it defines which namespaces an operator can watch.  
This helper centralises discovery of all OperatorGroups so other parts of CertSuite can reason about operator scope and placement.

### Inputs
None – the function obtains its own data from the Kubernetes client configured in the package’s global `ClientsHolder`.

### Outputs
- **`[]*olmv1.OperatorGroup`** – a slice containing pointers to each OperatorGroup found.  
  The slice may be empty if no groups exist or if an error occurs.
- **`error`** – non‑nil when any of the following happen:
  * failure to retrieve the client set (`GetClientsHolder`)
  * Kubernetes API call fails
  * Unexpected list result (e.g., nil pointer)

### Key Steps & Dependencies

| Step | Description | Dependency |
|------|-------------|------------|
| 1. **Client Retrieval** | `GetClientsHolder()` fetches a cached `Clients` struct that holds the OLM client (`OperatorsV1()`). | `GetClientsHolder`, `OperatorsV1` |
| 2. **List OperatorGroups** | Calls `client.OperatorGroups().List(ctx, metav1.ListOptions{})`. The list request uses the default context and no namespace filter (cluster‑wide). | `List`, `OperatorGroups` |
| 3. **Error Handling** | If the API returns *NotFound*, the function logs a warning but treats it as an empty result rather than failure. Any other error is wrapped with context and returned. | `IsNotFound`, `Warn` |
| 4. **Result Aggregation** | Iterates over the returned list, appending each item to a slice of pointers. The length of the final slice is logged for diagnostics. | `len`, `append` |

### Side Effects
- Logs informational or warning messages via the package’s logger (`Warn`).  
- Does **not** modify any cluster state.

### Package Context
`GetAllOperatorGroups` lives in the `provider` package, which provides a thin abstraction over Kubernetes/OLM clients.  
Other provider helpers (e.g., node inspection, pod analysis) rely on this function to understand operator scopes when validating workloads or network policies.  

#### Mermaid Flow (optional)

```mermaid
flowchart TD
    A[GetClientsHolder] --> B[OperatorsV1()]
    B --> C[OperatorGroups()]
    C --> D[List(...)]
    D -- success --> E{items}
    E --> F[append to slice]
    D -- NotFound --> G[Warn & return empty slice]
    D -- other error --> H[return wrapped error]
```

---

**Bottom line:**  
`GetAllOperatorGroups` is a read‑only, cluster‑wide discovery helper that returns every OperatorGroup CRD present, handling missing resources gracefully and logging its progress for debugging.
