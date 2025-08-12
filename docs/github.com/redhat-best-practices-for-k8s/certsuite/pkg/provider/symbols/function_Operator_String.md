Operator.String` – String representation of an operator

| Item | Detail |
|------|--------|
| **Receiver** | `Operator` (value receiver) |
| **Signature** | `func() string` |
| **File/Line** | `pkg/provider/operators.go:72` |
| **Dependencies** | Uses `fmt.Sprintf`. No other functions or globals are referenced. |

### Purpose

The method returns a human‑readable description of an operator instance, primarily for logging and debugging purposes.  
An `Operator` represents a workload that may be running in the cluster (e.g., a Deployment, DaemonSet, etc.). The string output is used by test reporters to identify which operator was involved when a check fails.

### How it works

```go
func (o Operator) String() string {
    return fmt.Sprintf("%s (%s)", o.Name, o.Namespace)
}
```

1. **`o.Name`** – the Kubernetes object name of the operator (e.g., `"csi-node-driver-operator"`).  
2. **`o.Namespace`** – the namespace where the operator is deployed (usually `"openshift-storage"` or `"kube-system"`).

The method concatenates these two fields in the format:

```
<name> (<namespace>)
```

No side effects occur; the function merely formats and returns a string.

### Relationship to the rest of the package

- **`Operator` struct**: defined elsewhere in `operators.go`. It contains metadata about a running operator, such as its type (Deployment/DaemonSet), status, etc.  
- **Logging & Reporting**: Test suites that iterate over operators use this method to print which operator is being examined or has failed validation.  
- **No external state**: The function is pure and does not modify package globals like `MasterLabels`, `WorkerLabels`, or environment variables.

### Usage example

```go
op := Operator{Name:"csi-node-driver-operator", Namespace:"openshift-storage"}
fmt.Println(op.String())
// → "csi-node-driver-operator (openshift-storage)"
```

This concise representation keeps logs readable while still identifying the operator’s deployment scope.
