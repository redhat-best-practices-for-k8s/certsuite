Node.IsControlPlaneNode`

```go
func (n Node) IsControlPlaneNode() bool
```

#### Purpose  
Determines whether a Kubernetes node belongs to the *control‑plane* (master) pool.

The function inspects the node’s labels and returns **true** if any of the known control‑plane
labels are present.  It is used by tests that need to skip or run checks only on worker nodes,
or when building a list of nodes for role‑specific diagnostics.

#### Inputs  
- `n Node` – the receiver; a node representation that exposes its labels as a map
  (typically `map[string]string`).  

No other parameters are required.

#### Outputs  
- `bool` – `true` if the node carries at least one label from the `MasterLabels`
  slice, otherwise `false`.

#### Key dependencies
| Dependency | Role |
|------------|------|
| `MasterLabels` (global) | Slice of strings that contain all known control‑plane labels. |
| `StringInSlice` (helper function) | Checks whether a string is present in a slice; used to test each label. |

> **Note**:  
> *The `WorkerLabels` global is unrelated to this method—it is used elsewhere to detect worker nodes.*

#### Side effects
None – the method only reads node labels and performs an immutable check.

#### How it fits the package

In the `provider` package, which models Kubernetes objects for CertSuite’s testing logic,
role detection is central.  
`IsControlPlaneNode` provides a reusable predicate that other components (e.g., test runners,
diagnostic collectors) invoke to filter or conditionally execute logic based on node role.

```mermaid
flowchart TD
    A[Node] -->|has labels| B[MasterLabels]
    B --> C{Any label present?}
    C -- yes --> D[IsControlPlaneNode() returns true]
    C -- no  --> E[returns false]
```

This concise check underpins many higher‑level decision points in the test suite, ensuring
that control‑plane specific tests are only applied where appropriate.
