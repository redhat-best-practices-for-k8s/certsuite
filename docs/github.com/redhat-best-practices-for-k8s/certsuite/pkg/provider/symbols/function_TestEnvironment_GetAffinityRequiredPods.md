TestEnvironment.GetAffinityRequiredPods`

### Purpose
`GetAffinityRequiredPods` is a convenience method on the **`TestEnvironment`** type that extracts all Pods in the current test environment which have an *affinity requirement* defined.  
Affinity requirements are used to control pod placement (e.g., node, pod or anti‑pod affinity).  The function therefore returns only those pods that will be scheduled with such constraints.

### Signature
```go
func (env TestEnvironment) GetAffinityRequiredPods() []*Pod
```

- **Receiver**: `TestEnvironment` – the struct that holds all resources of a test run (pods, nodes, deployments, etc.).  
- **Return value**: a slice of pointers to `Pod` objects (`[]*Pod`). The slice contains only pods for which the helper function `AffinityRequired` evaluates to `true`.

### Key dependencies
| Dependency | Role |
|------------|------|
| `env.Pods` | The complete set of pods that have been loaded into the test environment. |
| `AffinityRequired(pod *Pod) bool` | Predicate that checks whether a pod declares any affinity rules (node/pod/anti‑pod). This function is defined elsewhere in the package. |
| `append([]*Pod, *Pod)` | Standard Go slice operation used to accumulate matching pods. |

### Side effects
None – the method only reads from the environment and builds a new slice; it does not modify the underlying data structures.

### How it fits the package
The **`provider`** package implements a lightweight abstraction over Kubernetes objects that are needed for certsuite tests.  
- `TestEnvironment` is the central struct holding all loaded resources.  
- Various filter helpers (e.g., `GetAffinityRequiredPods`, `GetNodesWithHugePages`) allow callers to query specific subsets of those resources.  
- `GetAffinityRequiredPods` is typically used by test cases that need to validate placement policies or ensure that affinity‑enabled workloads are scheduled correctly.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[TestEnvironment] -->|contains| B[Pod list]
    B --> C{Has Affinity?}
    C -- yes --> D[Collect into result slice]
    C -- no  --> E[Skip]
    D --> F[Return []*Pod]
```

This diagram shows the flow from the test environment’s pod collection, through the affinity check, to the final slice returned.
