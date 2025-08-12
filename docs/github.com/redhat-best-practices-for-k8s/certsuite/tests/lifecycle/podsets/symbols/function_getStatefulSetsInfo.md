getStatefulSetsInfo`

| Aspect | Detail |
|--------|--------|
| **Package** | `podsets` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets`) |
| **Visibility** | Unexported (used only within this package) |
| **Signature** | `func getStatefulSetsInfo(ss []*provider.StatefulSet) []string` |
| **Purpose** | Convert a slice of Kubernetes StatefulSet objects into a simple list of strings that uniquely identify each set by its namespace and name. The resulting format is `"namespace: name"`. This helper is used by test utilities that need to report or log which StatefulSets are being inspected, e.g., when waiting for scaling operations or readiness checks. |
| **Inputs** | `ss`: a slice of pointers to `provider.StatefulSet` (the type comes from the certsuite provider package; it represents a Kubernetes StatefulSet). |
| **Outputs** | A new slice of strings, one per input StatefulSet, containing `"namespace: name"`. The order matches the original slice. |
| **Key dependencies** | * `fmt.Sprintf` – formats each string.<br>* `append` – builds the result slice. |
| **Side‑effects** | None – purely functional. No global state is read or modified. |
| **Integration** | Called by other helper functions in this package (e.g., those that wait for scaling to complete) to produce human‑readable identifiers for logging, assertion messages, or test output. Because the function lives in a test package, it’s not part of the public API but facilitates clear diagnostics during test runs. |

### Example usage

```go
// Assume stsList is [](*provider.StatefulSet)
info := getStatefulSetsInfo(stsList)
// info might be: ["default: mysql", "prod: redis"]
log.Println("Checking stateful sets:", strings.Join(info, ", "))
```

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Input slice of *provider.StatefulSet] --> B{for each}
    B --> C[fmt.Sprintf("%s: %s", ns, name)]
    C --> D[append to result]
    D --> E[Return []string]
```

This helper keeps the test logic concise and separates formatting concerns from business logic.
