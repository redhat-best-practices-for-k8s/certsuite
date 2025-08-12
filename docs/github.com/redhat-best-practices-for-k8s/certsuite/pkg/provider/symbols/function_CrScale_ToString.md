CrScale.ToString()` – String Representation of a Scale Object

| Element | Details |
|---------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver type** | `CrScale` (defined in `scale_object.go`) |
| **Signature** | `func (c CrScale) ToString() string` |
| **Exported?** | Yes |

### Purpose
Converts a `CrScale` instance into a human‑readable description that can be used for logging, diagnostics or test output.  
A *scale object* represents the desired number of replicas for an application component (Deployment, ReplicaSet, etc.). The string contains:

1. **Kind** – the Kubernetes resource type (`Deployment`, `StatefulSet`, …).  
2. **Name** – the metadata name of that resource.  
3. **Desired replicas** – how many pods should run.  
4. **Current replicas** – the number actually running (if available).

The method is intentionally lightweight and side‑effect free; it only formats data.

### Inputs
- `c CrScale` – the receiver contains at least:
  - `Kind string`
  - `Name string`
  - `DesiredReplicas int32`
  - Optionally, `CurrentReplicas *int32` (pointer to allow nil when not known).

No external arguments are required.

### Output
- Returns a single line `string` such as  
  ```
  Deployment my-app: desired=3 current=2
  ```

If the `CurrentReplicas` field is `nil`, the output omits it:
```
Deployment my-app: desired=3
```

### Key Dependencies
| Dependency | How it’s used |
|------------|---------------|
| `fmt.Sprintf` | Formats the string. No other packages are called. |

### Side‑Effects
- None. The method reads only the receiver fields and returns a new string; it does not modify state or interact with external systems.

### Package Context
The `provider` package implements high‑level test logic for Certsuite, including interaction with Kubernetes objects.  
`CrScale` is used throughout the provider to:

* Track expected vs actual replica counts.
* Generate status messages for tests that validate scaling behavior (e.g., after a rollout or during a node drain).
* Serialize scale information into logs and result files.

By providing a concise `ToString()` method, test writers can embed readable descriptions in assertion failures or telemetry without duplicating formatting logic.
