## `IsWorkerNode` – Detecting a Worker Node

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver** | `n *Node` (the node being inspected) |
| **Signature** | `func (n Node) IsWorkerNode() bool` |
| **Visibility** | Exported – can be called from other packages. |

### Purpose

Determines whether a Kubernetes node is considered a *worker* node according to the label sets defined in this provider package.

> **Why it matters:**  
> Many tests and checks run only on worker nodes (e.g., pod placement, network policies). This helper abstracts the label‑based logic so callers need not duplicate it.

### How It Works

1. **Label Retrieval** – The function reads the node’s `Labels` map (`map[string]string`) which is part of the underlying Kubernetes API representation.
2. **Membership Check** –  
   * It iterates over each key in the global slice `WorkerLabels`.  
   * For every key, it checks if that label exists on the node using the helper `StringInSlice` (a simple existence check).
3. **Result** – Returns `true` if **any** of the worker‑label keys are present; otherwise returns `false`.

The function relies only on:
- The `Node` type’s `Labels` field.
- The exported global slice `WorkerLabels`.
- The helper `StringInSlice`, which simply tests membership in a slice.

### Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `WorkerLabels` | Contains the set of label keys that identify worker nodes. |
| `StringInSlice` | Utility to test if a string is present in a slice. |

The function has **no side effects** – it performs only read‑only operations on the node’s labels and global constants.

### Relationship to Other Package Items

- **`MasterLabels`**: The complementary set used by `Node.IsMasterNode()` (not shown here) for detecting master nodes.
- **Label definitions**: These are defined once in `provider.go`; all role checks use them, ensuring consistency across the codebase.
- **Node structure**: `Node` is a wrapper around Kubernetes `v1.Node`. `IsWorkerNode` is one of several helpers that interpret node metadata.

### Example Usage

```go
// Iterate over all nodes and print worker names
for _, n := range provider.GetAllNodes() {
    if n.IsWorkerNode() {
        fmt.Println("Worker:", n.Name)
    }
}
```

### Summary

`IsWorkerNode()` is a small, read‑only helper that centralizes the logic for determining whether a node belongs to the worker role. It uses the globally defined `WorkerLabels` slice and a generic string‑in‑slice check, producing a boolean result with no side effects.
