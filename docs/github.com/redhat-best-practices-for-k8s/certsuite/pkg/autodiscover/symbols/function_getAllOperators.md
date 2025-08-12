getAllOperators`

```go
func getAllOperators(client v1alpha1.OperatorsV1alpha1Interface) ([]*olmv1Alpha.ClusterServiceVersion, error)
```

### Purpose  
`getAllOperators` retrieves **all** Operator Lifecycle Manager (OLM) `ClusterServiceVersion` objects that are visible to the supplied OLM client.  
It is used by the autodiscover logic to build a list of available operators and their CSVs for further analysis (e.g., label extraction, namespace resolution).

### Parameters
| Name   | Type                                         | Description |
|--------|----------------------------------------------|-------------|
| `client` | `v1alpha1.OperatorsV1alpha1Interface` | OLM client interface that exposes the `ClusterServiceVersions()` method for listing CSVs. |

> **Note**: The function is unexported; callers must use higher‑level helpers in this package.

### Return Values
| Value | Type                                 | Description |
|-------|--------------------------------------|-------------|
| `[]*olmv1Alpha.ClusterServiceVersion` | slice of pointers to `ClusterServiceVersion` | All CSVs found by the client. |
| `error` | `error` | Non‑nil if the list operation fails. |

### Key Dependencies
- **OLM Client** – Provides `ClusterServiceVersions().List(ctx, opts)` to fetch CSV objects.
- **Context & Options** – Uses a background context and default `metav1.ListOptions{}` (no filtering).
- **Logging** – Emits an informational log when the list succeeds or an error log on failure.

### Side Effects
- No mutation of global state.
- Logs messages via the package logger (`log.Info` / `log.Error`).

### Flow Overview
```mermaid
flowchart TD
    A[Call getAllOperators(client)] --> B{List CSVs}
    B -->|Success| C[Return slice, nil error]
    B -->|Failure| D[Log error, return nil, err]
```

1. Calls `client.ClusterServiceVersions().List(context.Background(), metav1.ListOptions{})`.
2. If the call errors, logs and returns the error.
3. On success, extracts the `.Items` field from the returned list and converts it to a slice of pointers.
4. Returns that slice with a nil error.

### How It Fits in `autodiscover`
- The package orchestrates discovery of operators, CRDs, and custom resources across namespaces.
- `getAllOperators` is a low‑level helper used by higher‑level functions (e.g., `DiscoverOperators`) to obtain the full set of CSVs before filtering or processing them further.

--- 

**Unknown**: Any additional internal state changes outside this function are not observable from its signature.
