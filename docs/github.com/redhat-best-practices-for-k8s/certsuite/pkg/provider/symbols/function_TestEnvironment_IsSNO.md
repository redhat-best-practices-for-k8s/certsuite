TestEnvironment.IsSNO`

```go
func (e TestEnvironment) IsSNO() bool
```

### Purpose  
`IsSNO` determines whether the test environment represents a **Single‑Node OpenShift** (SNO) cluster.  
In SNO, there is exactly one node that plays both master and worker roles; no separate dedicated control plane or worker nodes exist.

The function simply checks the number of nodes in the `TestEnvironment`. If only one node exists, it returns `true`; otherwise it returns `false`.

### Parameters & Return Value
| Name | Type | Description |
|------|------|-------------|
| none | – | – |
| **Returns** | `bool` | `true` if the environment contains a single node (SNO), otherwise `false`. |

### Key Operations

1. **Length Check**  
   ```go
   return len(e.Nodes) == 1
   ```
   * `e.Nodes` is a slice of `NodeInfo` objects that represent all nodes discovered in the cluster.

2. **No side‑effects** – The method only reads from the receiver; it does not modify state or perform external actions.

### Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `len` (built‑in) | Counts elements of `e.Nodes`. |

The function has no direct use of global variables, other functions, or types beyond those contained in the `TestEnvironment` struct.

### Context within the Package

- **Location**: `/pkg/provider/provider.go` – part of the core provider logic that abstracts cluster details for tests.  
- **Relation to Other Code**:  
  - The result of `IsSNO()` may influence test selection or configuration elsewhere in the suite (e.g., skipping worker‑only tests).  
  - It relies on `TestEnvironment.Nodes`, which is populated during environment discovery, typically by inspecting node labels (`MasterLabels` / `WorkerLabels`) or via Kubernetes API calls.

### Usage Example

```go
env := provider.NewTestEnvironment(...)
if env.IsSNO() {
    fmt.Println("Running tests in Single‑Node OpenShift mode")
}
```

---

**Mermaid diagram suggestion**

```mermaid
flowchart TD
  A[TestEnvironment] --> B{IsSNO()}
  B -- len==1 --> C[Return true]
  B -- else --> D[Return false]
```

This simple flow captures the core logic of `IsSNO`.
