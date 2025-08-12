isSingleNamespacedOperator`

| Aspect | Detail |
|--------|--------|
| **Package** | `operator` – test helpers for Certsuite’s operator tests. |
| **Visibility** | Unexported (used only within the package). |
| **Signature** | `func isSingleNamespacedOperator(namespace string, namespaces []string) bool` |

### Purpose
`isSingleNamespacedOperator` determines whether an operator under test is configured to run in a *single* namespace.  
In many tests we need to know if the operator’s scope is limited to one specific namespace (e.g., `cert-manager-operator`) or if it operates cluster‑wide. The helper simply checks the length of the slice that lists all namespaces where the operator is expected to be present.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `namespace` | `string` | The namespace name currently being inspected (unused in the current implementation but kept for future extensions). |
| `namespaces` | `[]string` | Slice containing all namespaces that should contain the operator. |

### Return Value
- `true` if `len(namespaces) == 1`, indicating a single‑namespace deployment.
- `false` otherwise.

### Dependencies & Side Effects
* **Dependencies** – Relies only on Go’s built‑in `len` function; no external packages or globals are accessed.  
* **Side effects** – None; the function is pure and has no observable impact on program state.

### Usage Context
The helper is invoked by test suites that need to adjust their expectations based on operator scope. For example:

```go
if !isSingleNamespacedOperator(ns, env.OperatorNamespaces) {
    // Run cluster‑wide validation logic
}
```

Here `env` refers to the package-level variable of type `provider.TestEnvironment`, which holds configuration such as `OperatorNamespaces`. The function is used in `tests/operator/helper.go` and indirectly influences assertions in various test files under `tests/operator`.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Call: isSingleNamespacedOperator] --> B{len(namespaces)}
    B -- 1 --> C["Return true\n(single‑namespace)"]
    B -- >1 or 0 --> D["Return false\n(cluster‑wide)"]
```

This visual summarizes the decision path within the function.
