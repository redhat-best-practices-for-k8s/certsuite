Pod.String` – Human‑readable description of a pod

```go
func (p Pod) String() string
```

| Element | Details |
|---------|---------|
| **Purpose** | Provide a concise, human‑readable representation of a `Pod` value. The method is intended for debugging, logging and any place where the caller needs to see the pod’s key attributes in text form. |
| **Receiver** | `p Pod` – an instance of the `Pod` type defined in *pods.go*. No fields are mutated; the method is pure. |
| **Return value** | A single string that contains a formatted summary of the pod, constructed with `fmt.Sprintf`. The exact layout is not shown in the JSON snippet but typically includes at least: `<namespace>/<name> (status)` or similar. |
| **Key dependencies** | *Standard library* – uses `fmt.Sprintf` to build the string. No other packages or global variables are accessed. |
| **Side effects** | None; it only reads from the receiver and returns a value. |
| **Package context** | The `provider` package models Kubernetes resources (nodes, pods, containers, etc.). `Pod.String` is part of that model, enabling callers to log or display pod information without exposing internal struct layout. |

### How it fits in the package

* `Pod` objects are created and manipulated throughout the provider code when querying a cluster.  
* When diagnostics or test results need to reference a specific pod, calling `.String()` gives a consistent textual key that can be printed or used as a map key.  
* Because the method is exported (`String`) it also satisfies Go’s `fmt.Stringer` interface, allowing pods to be automatically formatted by functions such as `log.Printf("%v", pod)`.

---

#### Suggested Mermaid diagram (pod representation)

```mermaid
flowchart TD
    Pod -->|has fields| Namespace
    Pod -->|has fields| Name
    Pod -->|has fields| Status
    Pod -->|String()| "namespace/name (status)"
```

This diagram illustrates that `Pod.String` consumes the pod’s namespace, name and status to produce a single string.
