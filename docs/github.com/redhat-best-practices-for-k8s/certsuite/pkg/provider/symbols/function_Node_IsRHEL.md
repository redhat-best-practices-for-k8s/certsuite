Node.IsRHEL`

```go
func (n Node) IsRHEL() bool
```

#### Purpose  
Determines whether the Kubernetes node represented by `Node` is running a Red‑Hat Enterprise Linux (RHEL) operating system.

#### How it works  
1. **Trim whitespace** – The method first removes any leading or trailing spaces from the value of the platform label stored in the `Node`.  
2. **Containment test** – It then checks whether this cleaned string contains one of the substrings that identify RHEL distributions (e.g., `"rhel"`, `"centos"`).  
3. **Result** – If a match is found, it returns `true`; otherwise it returns `false`.

The logic relies on two standard library helpers:

| Helper | Role |
|--------|------|
| `strings.TrimSpace` | Cleans the raw label value |
| `strings.Contains`   | Performs a substring search |

#### Inputs & Outputs  

- **Input**: The receiver `n Node`.  
  - The node’s platform information is read from an internal field (typically a map of labels).  
- **Output**: `bool` – `true` if the node’s OS is identified as RHEL‑based, `false` otherwise.

#### Dependencies  

| Dependency | Where it comes from |
|------------|---------------------|
| `Contains`, `TrimSpace` | Go's standard `strings` package (implicitly imported) |
| Node structure | Defined in the same package (`pkg/provider/nodes.go`) |

#### Side effects  
None. The function is pure: it only reads data from the receiver and returns a value.

#### Context within the package  

The `provider` package models the state of an OpenShift/Kubernetes cluster.  
- **Node roles** are defined via `MasterLabels` and `WorkerLabels`.  
- **Operating‑system detection** (like `IsRHEL`) is used by higher‑level logic to decide which configuration or test suite should be applied—for example, certain tests may only run on RHEL nodes.

Thus, `Node.IsRHEL` is a small but essential helper that allows the rest of the provider codebase to branch its behavior based on the underlying OS.
