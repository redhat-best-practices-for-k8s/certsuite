TestEnvironment.GetMasterCount`

```go
func (e TestEnvironment) GetMasterCount() int
```

### Purpose  
`GetMasterCount` returns the number of **control‑plane** nodes that belong to the current test environment.  
It is used by tests that need to know how many master nodes are available for validation or configuration purposes.

### Inputs / Outputs  

| Parameter | Type | Description |
|-----------|------|-------------|
| `e TestEnvironment` (receiver) | struct | The test environment instance containing the list of all cluster nodes. |

**Return value**

- `int`: Count of nodes that satisfy the *control‑plane* role.

### Key Dependencies  

| Dependency | Role |
|------------|------|
| `IsControlPlaneNode` | Determines whether a given node is considered a master (control plane) by checking its labels against `MasterLabels`. This function is called for each node in the environment. |

No other external packages or global variables are accessed directly within this method.

### Side Effects  

- **None** – The function performs read‑only analysis on the receiver’s data and returns an integer; it does not modify any state.

### How It Fits the Package

The `provider` package models a Kubernetes test environment.  
Nodes in that environment carry role labels (see `MasterLabels` / `WorkerLabels`).  
`GetMasterCount` aggregates those nodes, enabling other parts of the suite (e.g., connectivity or configuration tests) to adapt their logic based on how many master nodes exist.

```mermaid
flowchart TD
  A[TestEnvironment] -->|contains| B[Node list]
  B --> C{for each node}
  C --> D{IsControlPlaneNode(node)}
  D -- true --> E[+1 count]
  D -- false --> F[skip]
  E & F --> G[count result]
```

*The function simply iterates over all nodes in the environment, increments a counter when `IsControlPlaneNode` returns true, and finally returns that counter.*
