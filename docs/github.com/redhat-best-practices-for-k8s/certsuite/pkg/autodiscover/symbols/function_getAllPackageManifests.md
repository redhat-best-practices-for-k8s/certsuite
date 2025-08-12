getAllPackageManifests`

| Aspect | Detail |
|--------|--------|
| **Signature** | `func(olmpkgclient.PackageManifestInterface) ([]*olmpkgv1.PackageManifest)` |
| **Exported?** | No – internal helper used only within the `autodiscover` package. |

### Purpose
Collects every *PackageManifest* resource that exists in a Kubernetes cluster and returns them as a slice of pointers.  
A PackageManifest is an Open Shift‐specific CRD that describes operator packages; this function is typically called when the discovery logic needs to inspect all available operators.

### Inputs
| Parameter | Type | Role |
|-----------|------|------|
| `pkgClient` | `olmpkgclient.PackageManifestInterface` | A typed client capable of listing PackageManifests via the OLM (Operator Lifecycle Manager) API. |

> **Note**: The function expects that `pkgClient` has already been created with a valid kubeconfig and context.

### Outputs
| Return value | Type | Meaning |
|--------------|------|---------|
| `[]*olmpkgv1.PackageManifest` | Slice of pointers | All PackageManifests found in the cluster. If an error occurs, an empty slice is returned. |

> The caller should inspect the length or contents to determine if any manifests were retrieved.

### Key Operations
1. **List Request**  
   Calls `pkgClient.List(ctx, listOptions)` to fetch all manifests.  
   - Uses `context.TODO()` – a placeholder that may later be replaced with a proper context.
2. **Error Handling**  
   On error, logs the issue via `log.Error(err)` (not shown in snippet but typical) and returns an empty slice.  
3. **Aggregation**  
   Iterates over the returned list (`pkgList.Items`) and appends each item to a local slice.

### Dependencies
- **`olmpkgclient.PackageManifestInterface`** – Interface from `github.com/operator-framework/operator-lifecycle-manager/pkg/api/client`. Provides the `List` method.
- **`olmpkgv1.PackageManifest`** – CRD type representing an OLM package manifest.
- **Context package (`context`)** – Used to create a context for the list call.
- **Logging** – The function calls `log.Error`, which is likely from the `github.com/sirupsen/logrus` or similar logger bundled in this repo.

### Side‑Effects
- No global state is modified; the function only reads from the cluster via the provided client.
- Errors are logged but not returned, making the caller responsible for checking if the result slice is empty.

### Role in `autodiscover`
The autodiscovery subsystem needs to know which operators are present in order to:
1. Detect operator‑specific resources (e.g., CSVs).
2. Determine whether certain features or policies should be applied.
3. Populate internal data structures that drive certificate management decisions.

`getAllPackageManifests` is a small but essential building block: it turns the cluster’s OLM state into an in‑memory list that other discovery helpers can consume.

---

#### Suggested Mermaid Flow

```mermaid
flowchart TD
    A[Call getAllPackageManifests(pkgClient)] --> B[List PackageManifests via pkgClient]
    B --> C{Error?}
    C -- Yes --> D[Log error & return []]
    C -- No --> E[Iterate over Items]
    E --> F[Append each Item to slice]
    F --> G[Return []*PackageManifest]
```

---
