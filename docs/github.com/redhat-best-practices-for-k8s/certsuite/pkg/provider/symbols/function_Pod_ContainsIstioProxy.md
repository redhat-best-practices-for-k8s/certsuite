Pod.ContainsIstioProxy`

```go
func (p Pod) ContainsIstioProxy() bool
```

### Purpose  
`ContainsIstioProxy` checks whether a Kubernetes pod contains an Istio side‑car proxy container.  
In CertSuite, the presence of an Istio proxy influences how network connectivity tests are run – pods that carry the proxy may be treated differently because traffic is intercepted by Envoy.

### Receiver & Inputs  

| Element | Type | Description |
|---------|------|-------------|
| `p`     | `Pod` | The pod instance to inspect. The `Pod` type (defined in *pods.go*) holds metadata and a slice of containers (`Containers []Container`). |

No function arguments are required.

### Output  
Returns a single boolean:

- **`true`** – the pod contains at least one container whose name matches the constant `IstioProxyContainerName` (`"istio-proxy"`).
- **`false`** – no such container is found.

### Implementation Details  

The method iterates over `p.Containers`.  
For each container it compares `container.Name` with the exported string constant:

```go
const IstioProxyContainerName = "istio-proxy"
```

If a match is found, the loop breaks and `true` is returned immediately; otherwise the function returns `false`.

### Dependencies & Side‑Effects  

| Dependency | Role |
|------------|------|
| `IstioProxyContainerName` | Constant used for comparison. |
| `Pod.Container` type | Provides the `Name` field. |

No external state, global variables, or side effects are involved – the function is pure.

### Package Context  

`ContainsIstioProxy` belongs to the **provider** package of CertSuite, which orchestrates test execution against a Kubernetes cluster.  
The method is used by:

- Connectivity tests that need to know if a pod has an Istio proxy before deciding how to reach it.
- Any logic that filters or counts pods based on side‑car presence.

By keeping the check encapsulated in the `Pod` type, the rest of the package can treat pods uniformly without duplicating container‑name logic.
