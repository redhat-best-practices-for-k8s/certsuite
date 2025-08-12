## `GetPodsUsingSRIOV`

| Aspect | Detail |
|--------|--------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Receiver** | `TestEnvironment` – holds the current list of pods and other environment state. |
| **Signature** | `func (env TestEnvironment) GetPodsUsingSRIOV() ([]*Pod, error)` |
| **Exported** | Yes |

### Purpose

Collects all Kubernetes pods that are configured to use SR‑IOV virtual network interfaces.

The function iterates over the internal pod list (`TestEnvironment.Pods`) and filters those for which `IsUsingSRIOV` returns `true`.  
If any error occurs while checking a pod, the whole operation fails immediately with an annotated error.

### Inputs & Outputs

| Input | Type | Description |
|-------|------|-------------|
| `env` (receiver) | `TestEnvironment` | The environment snapshot that contains all pods to be inspected. |

| Output | Type | Description |
|--------|------|-------------|
| `[]*Pod` | Slice of pointers to `Pod` objects | Only the pods that actually use SR‑IOV. |
| `error` | `nil` on success, otherwise an error describing the first failure encountered. |

### Key Dependencies

1. **`IsUsingSRIOV(pod *Pod) (bool, error)`**  
   - Determines whether a single pod uses SR‑IOV.
2. **Standard library helpers**  
   - `append`: to build the result slice.  
   - `fmt.Errorf`: for error wrapping.

No global variables are accessed directly; the function relies solely on the receiver’s data.

### Side‑Effects

- No mutation of the environment or pod objects – purely read‑only.
- The returned slice is a new collection; modifying it does not affect the original `TestEnvironment.Pods`.

### How It Fits Into the Package

`GetPodsUsingSRIOV` is part of the **filters** subsystem (file: `filters.go`).  
The provider package exposes various query helpers that allow tests to focus on specific pod characteristics.  
This function is used by higher‑level test suites that need to assert SR‑IOV configuration, e.g.:

```go
pods, err := env.GetPodsUsingSRIOV()
if err != nil { /* handle error */ }
for _, p := range pods {
    // perform further checks
}
```

### Flow Diagram (Mermaid)

```mermaid
flowchart TD
  A[Start] --> B{env.Pods not empty?}
  B -- No --> C[Return empty slice, nil]
  B -- Yes --> D[Initialize result []Pod]
  E[For each pod in env.Pods] --> F[Call IsUsingSRIOV(pod)]
  F --> G{Error?}
  G -- Yes --> H[Wrap error with fmt.Errorf and return]
  G -- No --> I{Uses SR‑IOV?}
  I -- No --> J[Continue loop]
  I -- Yes --> K[Append pod to result]
  K --> J
  J --> L{Loop finished?}
  L -- Yes --> M[Return result, nil]
```

This diagram illustrates the linear iteration and early‑exit error handling.

---

**Summary**  
`GetPodsUsingSRIOV` is a lightweight read‑only helper that filters a test environment’s pod list for SR‑IOV usage. It relies on `IsUsingSRIOV`, reports errors immediately, and returns only the relevant pods to the caller.
