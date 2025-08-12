ownedByClusterWideOperator`

| Aspect | Detail |
|--------|--------|
| **Package** | `accesscontrol` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol`) |
| **Signature** | `func ownedByClusterWideOperator(topOwners map[string]podhelper.TopOwner, env *provider.TestEnvironment) (string, bool)` |
| **Exported?** | No – internal helper used by the test suite. |

### Purpose

`ownedByClusterWideOperator` inspects a set of *top‑owners* (the highest‑level owners in an object’s ownership chain) to determine whether any of them is a **Cluster Service Version (CSV)** that was installed by a **cluster‑wide operator**.

If such a CSV exists, the function returns its name and a `true` flag.  
Otherwise it signals absence with an empty string and `false`.

This helper is used in tests that verify access‑control rules for resources owned by operators that run cluster‑wide (i.e., not namespaced).

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `topOwners` | `map[string]podhelper.TopOwner` | A map keyed by object UID to a `TopOwner`. Each entry represents the highest‑level owner of some resource under test. |
| `env` | `*provider.TestEnvironment` | Test environment that contains information about installed operators and CSVs (e.g., list of cluster‑wide operators). It is passed because the helper must query this data via `isCSVAndClusterWide`. |

### Return Values

| Index | Type | Meaning |
|-------|------|---------|
| 0 | `string` | The name of the matching CSV, if any. Empty string otherwise. |
| 1 | `bool` | `true` when a cluster‑wide operator’s CSV is found among the top owners; `false` otherwise. |

### Key Dependencies

* **`isCSVAndClusterWide`** – a helper function (not shown) that checks whether a given `TopOwner` represents a CSV installed by a cluster‑wide operator.  
  The current function delegates to it in a loop over all provided top owners.

* **`podhelper.TopOwner`** – the struct holding ownership metadata (name, kind, namespace).  
  It is used only for inspection; no modifications are made.

### Side Effects

The function performs read‑only checks and does not modify `topOwners`, `env`, or any global state.  
It merely evaluates conditions and returns results.

### How it fits the package

Within the `accesscontrol` test suite, many tests need to determine whether a resource is managed by a cluster‑wide operator in order to apply appropriate policy checks (e.g., ensuring that such operators do not expose privileged resources).  
`ownedByClusterWideOperator` provides a concise way to make this determination for any set of top owners discovered during the test run.

### Example Usage

```go
// In a Ginkgo test:
topOwners := getTopOwnersFromPods(pods)
if csvName, ok := ownedByClusterWideOperator(topOwners, env); ok {
    // The resource is owned by cluster‑wide operator CSV named `csvName`.
    // Apply specific assertions...
}
```

### Mermaid Diagram (optional)

```mermaid
graph TD
  subgraph Test Suite
    A[Collect Pods] --> B[Determine Top Owners]
    B --> C{ownedByClusterWideOperator()}
  end

  subgraph Helper
    C --> D{isCSVAndClusterWide() ?}
    D -->|Yes| E[Return CSV name & true]
    D -->|No| F[Continue loop / Return "", false]
  end
```

This function is a small, but essential building block for validating that cluster‑wide operators are correctly isolated in the test scenarios.
