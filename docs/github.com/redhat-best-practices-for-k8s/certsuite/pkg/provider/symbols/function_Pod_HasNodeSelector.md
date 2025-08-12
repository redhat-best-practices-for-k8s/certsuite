Pod.HasNodeSelector`

| | |
|-|-|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Receiver** | `p Pod` – the pod instance on which the method is called. |
| **Signature** | `func (p Pod) HasNodeSelector() bool` |

### Purpose
Determines whether a pod definition contains any node‑selector constraints.  
In OpenShift/Kubernetes, a *node selector* is a map of key/value pairs that restricts a pod to run only on nodes whose labels match the selector. The `HasNodeSelector` method answers: **“Does this pod have at least one node selector?”**

### Inputs / Outputs
- **Input** – implicit: the receiver (`p`) holds the pod spec, including its `Spec.NodeSelector` field.
- **Output** – a boolean:
  - `true` if `len(p.Spec.NodeSelector) > 0`
  - `false` otherwise

No external parameters are required; the method works purely on the internal state of the `Pod`.

### Key Dependencies
- Uses Go’s built‑in `len()` function to check the size of the selector map.
- Relies on the `Spec` field of the `Pod` struct (defined elsewhere in the package) which contains a `NodeSelector map[string]string`.

No global variables or other functions are accessed.

### Side Effects
None. The method is read‑only; it only inspects state and returns a value.

### How It Fits the Package
The `provider` package models OpenShift/Kubernetes objects for testing purposes.  
- Pods may be created with or without node selectors, affecting scheduling behaviour.
- Tests that validate cluster configuration often need to know whether a pod is constrained to specific nodes (e.g., master vs worker).
- `HasNodeSelector` provides a concise way for test logic to branch on this property, improving readability and reducing duplicated code.

```mermaid
graph TD
    Pod -->|contains| Spec
    Spec -->|has| NodeSelector(map)
    NodeSelector -->|size>0?| HasNodeSelector()
```

By exposing this helper, the package keeps pod‑related checks encapsulated and encourages consistent use across test suites.
