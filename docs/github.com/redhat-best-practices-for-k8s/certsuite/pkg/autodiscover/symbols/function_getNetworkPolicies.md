getNetworkPolicies`

**Location**  
`pkg/autodiscover/autodiscover_networkpolicies.go:27`

### Purpose
Collects all Kubernetes NetworkPolicy objects in the cluster and returns them as a slice.

The function is used by autodiscovery logic that needs to inspect existing network policies (e.g., to determine isolation rules for services or pods). It does **not** filter, modify, or delete any resources; it simply lists everything via the client passed in.

### Signature
```go
func getNetworkPolicies(networkingv1client.NetworkingV1Interface) ([]networkingv1.NetworkPolicy, error)
```
| Parameter | Type                                | Description |
|-----------|-------------------------------------|-------------|
| `networkingv1client` | `NetworkingV1Interface` | Client interface for the Networking‑v1 API group (from `k8s.io/client-go/kubernetes/typed/networking/v1`). |

**Returns**

* `[]networkingv1.NetworkPolicy` – all network policies found in the cluster.
* `error` – non‑nil if the list operation fails.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `List` (method on `NetworkPolicies`) | Performs a `GET /apis/networking.k8s.io/v1/networkpolicies` request. |
| `NetworkPolicies` (client method) | Provides access to the network policies resource endpoint. |
| `TODO` function call | Placeholder; currently does nothing, but indicates future work or debugging hooks. |

### Implementation Flow
```go
func getNetworkPolicies(client NetworkingV1Interface) ([]networkingv1.NetworkPolicy, error) {
    // 1. Call client.NetworkPolicies().List(ctx, options)
    // 2. Return the Items slice and any error.
}
```
* A context with `context.TODO()` is used – no cancellation or timeout is provided yet.
* The function simply returns the raw items; callers are responsible for filtering or processing.

### Side Effects
None. The function performs a read‑only API call and does not modify cluster state.

### Package Context
`autodiscover` dynamically discovers Kubernetes resources (CRDs, deployments, network policies, etc.) to configure certificate discovery rules.  
`getNetworkPolicies` is part of the internal helper set that abstracts away direct client calls, keeping higher‑level logic cleaner. It lives in `autodiscover_networkpolicies.go`, mirroring other “_network” helpers such as `getSriovResources`.

---

#### Suggested Mermaid Diagram (optional)

```mermaid
flowchart TD
  A[Caller] -->|calls|getNetworkPolicies
  getNetworkPolicies --> B[NetworkingV1Client]
  B --> C{NetworkPolicies}
  C --> D[List]
  D --> E[Returns []NetworkPolicy, error]
```

This visualizes the single API interaction performed by the function.
