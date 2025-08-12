Node.IsRTKernel() bool`

### Purpose
`IsRTKernel` determines whether the node on which it is called runs a **real‑time (RT) kernel**.

Real‑time kernels are typically used in performance‑critical workloads where deterministic latency is required.  
The method inspects the node’s kernel version string exposed by the Kubernetes API (`Node.Status.NodeInfo.KernelVersion`) and checks for known RT suffixes.

### Inputs & Outputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `receiver` | `Node` (value receiver) | The node instance whose kernel is examined. |

**Return value**

* `bool` – `true` if the kernel string contains any of the RT indicators, otherwise `false`.

### Implementation details
```go
func (n Node) IsRTKernel() bool {
    // Fetch raw kernel version string from the node's status.
    k := n.Status.NodeInfo.KernelVersion

    // Trim leading/trailing whitespace to avoid false negatives.
    k = strings.TrimSpace(k)

    // Check for any RT‑kernel identifier. The list is kept in a slice
    // (not shown here) and each element is searched with strings.Contains.
    for _, rt := range rtKernelIdentifiers {
        if strings.Contains(k, rt) {
            return true
        }
    }
    return false
}
```

* `strings.TrimSpace` – removes surrounding spaces that might be present in the kernel string.
* `strings.Contains` – used to detect substrings such as `"rt"`, `"preempt-rt"`, or other RT‑kernel tags.

The slice `rtKernelIdentifiers` is defined elsewhere in the package and typically contains values like:
```go
var rtKernelIdentifiers = []string{
    "rt",
    "preempt-rt",
}
```

### Dependencies
* **Standard library**: `strings.TrimSpace`, `strings.Contains`.
* **Node struct**: requires that `Status.NodeInfo.KernelVersion` is populated (normally provided by the Kubernetes API server).

No external packages are imported directly in this function.

### Side‑effects
None. The method only reads data from the node and performs string comparisons; it does not modify state or perform I/O.

### Package context
The `provider` package contains logic for evaluating cluster health and configuration.  
Nodes are central to many checks, such as verifying kernel features, resource limits, and role labels (`MasterLabels`, `WorkerLabels`).  

`IsRTKernel` is used by tests that need to differentiate between standard and RT kernels—for example, ensuring that workloads requiring deterministic scheduling run on appropriate nodes or validating that the cluster’s kernel configuration matches expectations.

---

#### Mermaid diagram (optional)

```mermaid
flowchart TD
    Node -->|Status.NodeInfo.KernelVersion| KernelString
    KernelString --> TrimSpace[TrimSpace]
    TrimSpace --> Contains[Contains(rtKernelIdentifiers)]
    Contains --> IsRT[Return bool]
```

This visualises the flow from node data to the boolean result.
