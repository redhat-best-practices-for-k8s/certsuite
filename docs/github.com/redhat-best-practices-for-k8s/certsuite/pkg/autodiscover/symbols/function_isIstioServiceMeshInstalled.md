isIstioServiceMeshInstalled`

| Aspect | Details |
|--------|---------|
| **Signature** | `func (appClient appv1client.AppsV1Interface, ns []string) bool` |
| **Visibility** | Unexported – used internally by the autodiscover package. |

### Purpose
Detects whether an Istio service mesh is present in a Kubernetes cluster that has *at least one* namespace from `ns`.  
The function checks for a specific deployment (`istioDeploymentName`) that should exist only when Istio is installed.

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `appClient` | `appv1client.AppsV1Interface` | A typed client for the *Apps* API group (contains Deployments, StatefulSets, etc.). |
| `ns` | `[]string` | List of namespace names to search. Istio’s control‑plane components are expected to live in one of these namespaces. |

### Return Value
- `true` if a deployment named `istioDeploymentName` is found in any of the supplied namespaces.
- `false` otherwise.

The function never panics; errors from the API client are logged and cause the check to return `false`.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `StringInSlice` | Checks whether a namespace string is already in the internal list of processed namespaces. |
| `appClient.Deployments(ns).Get(...)` | Retrieves a deployment object from the API server. |
| Logging helpers (`Info`, `Warn`, `Error`) | Emit diagnostic messages during execution. |
| `IsNotFound` | Determines whether an error returned by the API client is a “not found” error. |

### Side‑Effects

* Logs informational, warning, or error messages through the package’s logging infrastructure.
* No state is mutated in global variables.

### Flow Overview (Mermaid)

```mermaid
flowchart TD
  A[Start] --> B{Iterate over ns}
  B --> C{Deployments(ns).Get(istioDeploymentName)}
  C -->|Found| D[Info: Istio found]
  D --> E[Return true]
  C -->|Not Found| F[Warn: Deployment missing]
  F --> G[Continue loop]
  C -->|Other error| H[Error: API failure]
  H --> I[Return false]
  B --> J{Loop finished}
  J --> K[Info: No Istio found]
  K --> E
```

### How It Fits the Package

`autodiscover` dynamically determines which components are present in a cluster to adjust certificate discovery logic.  
Istio is a common service mesh that can alter traffic flow and TLS termination; knowing whether it’s installed allows the suite to apply appropriate tests or skip them.

This helper is called from higher‑level detection functions that iterate over known operator namespaces (e.g., Istio, OpenShift Service Mesh). By encapsulating the deployment lookup in this function, the package keeps its discovery logic modular and testable.
