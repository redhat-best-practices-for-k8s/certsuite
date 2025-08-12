Deployment`

`Deployment` is a thin wrapper around Kubernetes’ native **AppsV1 Deployment** object.  
It lives in the *provider* package and is used wherever the test harness needs to inspect or format deployment objects without pulling in the full client‑side logic.

| Aspect | Detail |
|--------|--------|
| **Purpose** | Provide a convenient, domain‑specific view of a Deployment that exposes only the helpers needed by CertSuite (readiness check & string representation). |
| **Composition** | The struct embeds `*appsv1.Deployment` from the official Kubernetes API (`k8s.io/api/apps/v1`). This gives it direct access to all fields of a standard Deployment while keeping the type name short. |
| **Fields** | *None explicitly defined* – only the embedded pointer. |
| **Key Methods** | • `IsDeploymentReady() bool`<br>• `ToString() string` |

### Method Details

#### `func (d Deployment) IsDeploymentReady() bool`

* **Input**: none (uses the embedded Deployment).  
* **Output**: `true` if the deployment’s status indicates that all desired replicas are available; otherwise `false`.  
  The implementation simply checks `d.Status.AvailableReplicas == d.Spec.Replicas`.  
* **Side‑effects**: none.  
* **Dependencies**: uses only fields from the embedded Deployment – no external packages or globals.

#### `func (d Deployment) ToString() string`

* **Input**: none.  
* **Output**: a human‑readable representation of the deployment’s key attributes, formatted as:
  ```
  <kind> <name> (<namespace>) - replicas:<desired>/<available>
  ```
  where `<kind>` is `Deployment`, `<name>` and `<namespace>` come from the embedded object's metadata, and replica counts come from spec & status.  
* **Side‑effects**: none.  
* **Dependencies**: calls `fmt.Sprintf` (standard library). No other packages.

### Relationship to Package

The *provider* package orchestrates interactions with Kubernetes resources. `Deployment` is the minimal representation used by higher‑level test functions such as:

```go
GetUpdatedDeployment(client, name, ns)
```

which fetches a raw Deployment from the cluster and wraps it in this struct for easier handling.

### Suggested Diagram

```mermaid
classDiagram
    class provider.Deployment {
        <<struct>>
        +*appsv1.Deployment
        +IsDeploymentReady() bool
        +ToString() string
    }
    class k8s.io/api/apps/v1.Deployment {
        <<struct>>
    }

    provider.Deployment --> k8s.io/api/apps/v1.Deployment : embeds
```

### Summary

`provider.Deployment` is a read‑only wrapper that:

1. Exposes the underlying Kubernetes Deployment via embedding.  
2. Provides two utility methods used throughout CertSuite: readiness checking and pretty printing.  
3. Has no mutable state or side effects beyond reading its embedded fields.

This design keeps the rest of the provider package decoupled from direct use of the raw API structs while still giving tests convenient access to deployment information.
