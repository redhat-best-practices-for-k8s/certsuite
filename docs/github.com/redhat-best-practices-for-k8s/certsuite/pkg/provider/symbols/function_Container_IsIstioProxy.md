## `func (c Container) IsIstioProxy() bool`

| Aspect | Details |
|--------|---------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Receiver type** | `Container` – a lightweight wrapper around a Kubernetes pod container (see `containers.go`). |
| **Signature** | `func (c Container) IsIstioProxy() bool` |
| **Exported** | ✅ |

---

### Purpose
Determines whether the receiver container is an Istio side‑car.  
In certsuite, certain tests are skipped or performed differently for Istio pods; this helper centralises that check.

### How it works
The method simply compares the container’s name against the constant `IstioProxyContainerName` defined in *pods.go*:

```go
const IstioProxyContainerName = "istio-proxy"
```

If the names match, the container is identified as an Istio proxy and the function returns `true`; otherwise it returns `false`.

### Inputs / Outputs
- **Input**: The method operates on a `Container` value (`c`).  
  No additional parameters are required.
- **Output**: A boolean indicating whether this container is the Istio side‑car.

### Key dependencies
| Dependency | Role |
|------------|------|
| `IstioProxyContainerName` (string) | The canonical name used to identify the Istio proxy container. |

No external services, network calls or mutable state are involved – the function is pure and thread‑safe.

### Side effects
None. It only reads from the receiver; no global state or side‑effects are modified.

### Integration in the package
- **Containers**: `Container` objects represent all containers discovered by certsuite’s provider logic.  
  `IsIstioProxy()` allows higher‑level code (e.g., test runners, pod evaluators) to treat Istio proxies specially.
- **Pods**: The pod‑listing logic in *pods.go* may use this method to filter out Istio side‑cars from the set of containers that should be inspected for TLS configuration or certificate validation.

### Usage example
```go
for _, container := range pod.Spec.Containers {
    if container.IsIstioProxy() {
        // Skip Istio proxy checks, log a message, etc.
        continue
    }
    // Run normal certificate checks on this container
}
```

---

#### Mermaid diagram (optional)

```mermaid
graph TD
  Pod -->|has| Container
  Container --isIstio--> IsIstioProxy()
  IsIstioProxy() --returns--> bool
```

This helper keeps the codebase consistent and makes it trivial to adjust the Istio side‑car name in a single place if upstream changes occur.
