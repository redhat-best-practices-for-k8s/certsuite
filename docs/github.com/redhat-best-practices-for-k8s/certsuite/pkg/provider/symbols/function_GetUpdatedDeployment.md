GetUpdatedDeployment`

**Package:** `provider`  
**Location:** `pkg/provider/deployments.go:68`  

### Purpose
`GetUpdatedDeployment` is a convenience wrapper that returns the *current* state of a Deployment object in a Kubernetes cluster, given its name and namespace. The function does not create or modify anything; it simply fetches the latest deployment definition from the API server.

The method is used by higher‑level logic that needs to compare an expected deployment spec against what is actually running – for example, during validation tests or when reconciling desired state with reality.

### Signature
```go
func GetUpdatedDeployment(
    client appv1client.AppsV1Interface,
    name string,
    namespace string,
) (*appsv1.Deployment, error)
```

| Parameter | Type                         | Description |
|-----------|------------------------------|-------------|
| `client`  | `AppsV1Interface`            | A typed Kubernetes client that can access the Apps/v1 API group (where Deployments live). |
| `name`    | `string`                     | The name of the Deployment to fetch. |
| `namespace` | `string`                 | The namespace containing the Deployment. |

### Return values
- `*appsv1.Deployment`: a pointer to the latest Deployment object retrieved from the cluster.
- `error`: non‑nil if the deployment cannot be found or the API call fails.

### Implementation Details
```go
func GetUpdatedDeployment(client appv1client.AppsV1Interface, name string, namespace string) (*appsv1.Deployment, error) {
    return FindDeploymentByNameByNamespace(client, name, namespace)
}
```

- The function simply forwards its arguments to `FindDeploymentByNameByNamespace`, which performs the actual API call (`client.Deployments(namespace).Get(ctx, name, metav1.GetOptions{})`).
- No additional logic or side‑effects are present; it is a thin wrapper for readability and potential future extension.

### Dependencies
| Dependency | Role |
|------------|------|
| `FindDeploymentByNameByNamespace` | The core lookup routine that interacts with the Kubernetes API. |

### Side Effects
None – purely read‑only.

### Context within the Package
The `provider` package contains utilities for inspecting and manipulating Kubernetes objects (nodes, pods, containers, etc.).  
`GetUpdatedDeployment` is part of the deployment‑related helpers; other functions in this file fetch or manipulate Deployment resources. It provides a clear, named entry point that callers can use when they need the *current* state of a Deployment without exposing the underlying client logic.

---

**Usage Example**

```go
client := kubernetes.NewForConfigOrDie(cfg).AppsV1()
deploy, err := provider.GetUpdatedDeployment(client, "nginx-deploy", "default")
if err != nil {
    log.Fatalf("failed to get deployment: %v", err)
}
fmt.Printf("Deployment %s has %d replicas\n",
    deploy.Name, *deploy.Spec.Replicas)
```

This snippet demonstrates how a test or controller might retrieve the latest deployment configuration and use it for validation.
