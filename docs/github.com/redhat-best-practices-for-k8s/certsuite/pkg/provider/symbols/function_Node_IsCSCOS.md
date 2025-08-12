Node.IsCSCOS`

```go
func (n Node) IsCSCOS() bool
```

### Purpose

`IsCSCOS` determines whether a node is running **CentOS Stream CoreOS** (CS-COS).  
The check is performed against the node’s label set, which is supplied by the Kubernetes API.  
It is used throughout the provider package to:

* decide if OS‑specific tests should run,
* apply OS‑specific configuration logic, and
* filter nodes when performing connectivity or performance checks.

### Inputs / Outputs

| Parameter | Type  | Description |
|-----------|-------|-------------|
| `n`       | `Node` (receiver) | The node instance whose labels are examined. |

| Return value | Type   | Meaning |
|--------------|--------|---------|
| `bool` | `true` if the node’s OS label indicates CentOS Stream CoreOS; otherwise `false`. |

### Implementation details

The function relies on two standard library helpers:

1. **`strings.TrimSpace`** – removes leading/trailing whitespace from the value retrieved from the node’s labels.
2. **`strings.Contains`** – checks if the cleaned OS label contains one of the substrings that identify CentOS Stream CoreOS.

```go
func (n Node) IsCSCOS() bool {
    osLabel := strings.TrimSpace(n.Labels["beta.kubernetes.io/os"])
    return strings.Contains(osLabel, "centos") ||
           strings.Contains(osLabel, "rhel")
}
```

*The exact substrings used are those that match the OS names returned by the node agent on CentOS Stream CoreOS installations.*

### Dependencies

| Dependency | Role |
|------------|------|
| `strings` package | Provides `TrimSpace` and `Contains`. |
| `Node.Labels` map | Holds the node’s label key‑value pairs. |

No global variables are accessed, so the function is pure with respect to the package state.

### Side effects

None – the method only reads from the receiver; it does not modify any state or produce external output.

### Package context

The `provider` package contains types and helpers for interacting with a Kubernetes cluster (nodes, pods, containers).  
`IsCSCOS` is part of the node utilities (`pkg/provider/nodes.go`) and is called by:

* test‑selection logic to skip or include OS‑specific tests,
* diagnostic routines that need to know whether a node is a CentOS Stream CoreOS host.

Because the function is exported, other packages (e.g., `cmd`, `tests`) can also use it to make OS‑dependent decisions.
