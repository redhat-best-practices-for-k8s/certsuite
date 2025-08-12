FindDeploymentByNameByNamespace`

### Purpose
Retrieves a single Kubernetes Deployment object from a cluster by its **name** and **namespace** using the client-go AppsV1 API.

> This helper is used throughout the autodiscover package to resolve the actual Deployment resource that backs various higher‑level resources (e.g., CSVs, CRDs). It abstracts away the `client.AppsV1Interface` plumbing so callers can simply ask for a deployment by its identifying strings.

### Signature
```go
func FindDeploymentByNameByNamespace(
    client appv1client.AppsV1Interface,
    namespace string,
    name string,
) (*appsv1.Deployment, error)
```

| Parameter | Type                         | Description |
|-----------|------------------------------|-------------|
| `client`  | `AppsV1Interface` (from `k8s.io/client-go/kubernetes/typed/apps/v1`) | The typed client used to query the API server. |
| `namespace` | `string` | Namespace of the target Deployment. |
| `name`      | `string` | Name of the target Deployment. |

### Return values
- `*appsv1.Deployment`: Pointer to the deployment object if found.
- `error`: Non‑nil when:
  - The client call fails (network, auth, etc.).
  - The deployment does not exist in the requested namespace.

### Key implementation steps

```go
func FindDeploymentByNameByNamespace(
    client appv1client.AppsV1Interface,
    namespace string,
    name string,
) (*appsv1.Deployment, error) {
    // The AppsV1Interface exposes a Deployments(namespace).Get(...) method.
    dep, err := client.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
    }
    return dep, nil
}
```

- **Context** – Uses `context.TODO()` because the caller is expected to handle cancellation/timeouts elsewhere.
- **Error wrapping** – Errors are wrapped with a message that includes the fully‑qualified deployment key for easier debugging.

### Dependencies & side effects
| Dependency | Role |
|------------|------|
| `client.Deployments(namespace).Get` | Performs an HTTP GET on `/apis/apps/v1/namespaces/{namespace}/deployments/{name}`. |
| `context.TODO()` | Provides a context; no cancellation or timeout logic is added here. |
| `fmt.Errorf` | Wraps errors for richer diagnostics. |

No global state is read or modified; the function is pure with respect to package globals.

### Usage in the package
- **CSV resolution** – When an Operator CSV references a Deployment, the autodiscover flow calls this helper to fetch that Deployment and then inspects its pod template.
- **Label extraction** – After obtaining the Deployment, other helpers read labels/annotations from `dep.Spec.Template.ObjectMeta` to determine the target namespace or probe configurations.

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Caller] --> B{FindDeploymentByNameByNamespace}
    B --> C[client.Deployments(namespace).Get]
    C --> D[API Server]
    D --> E{Success/Failure}
    E -- success --> F[*appsv1.Deployment]
    E -- failure --> G[Error]
```

This diagram illustrates the flow from caller to API server and back.

---

**Bottom line:**  
`FindDeploymentByNameByNamespace` is a lightweight wrapper around the Kubernetes client‑go deployment getter, providing consistent error handling for the autodiscover logic that relies on Deployment objects.
