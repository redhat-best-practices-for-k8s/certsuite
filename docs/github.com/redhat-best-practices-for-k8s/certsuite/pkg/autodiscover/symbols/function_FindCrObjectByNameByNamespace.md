FindCrObjectByNameByNamespace`

> **Signature**  
> ```go
> func FindCrObjectByNameByNamespace(
>     getter scale.ScalesGetter,
>     namespace string,
>     name string,
>     resource schema.GroupResource,
> ) (*scalingv1.Scale, error)
> ```

### Purpose

`FindCrObjectByNameByNamespace` is a helper that retrieves the *Scale* sub‑resource for an arbitrary custom resource (CR) identified by its **name** and **namespace**.  
The function is used in the autodiscover logic when a pod set’s scaling behaviour must be inspected or manipulated without knowing the concrete CRD type at compile time.

### Inputs

| Parameter | Type                        | Description |
|-----------|-----------------------------|-------------|
| `getter`  | `scale.ScalesGetter`       | A client that implements the Kubernetes *scales* API interface (usually a typed client from controller-runtime). It exposes `.Scales()` which returns an object capable of fetching scale sub‑resources. |
| `namespace` | `string`                  | The namespace in which the target CR resides. |
| `name`      | `string`                   | The name of the CR instance whose Scale is required. |
| `resource`  | `schema.GroupResource`     | A pair `(Group, Resource)` that identifies the CRD kind (e.g., `"apps/v1"` and `"deployments"`). This allows the function to call the correct REST endpoint for scaling. |

### Outputs

* `*scalingv1.Scale` – The scale object returned by the API. It contains the current replica count and, optionally, a target replica field if the CR supports it.
* `error` – Any error that occurred during the lookup. Common errors include:
  * The Scale sub‑resource is not supported for the given resource (e.g., the CRD does not implement the scale interface).
  * Network or authentication failures when contacting the API server.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `getter.Scales()` | Provides a client that can issue `Get` requests against the scale sub‑resource. |
| `scalingv1.Scale` | The Kubernetes type representing the scaling status of a resource. |
| `TODO` | A placeholder used by the code generator for error handling; replaced with an actual error in the compiled binary. |
| `Error` | Utility to wrap or construct errors (likely from the `github.com/pkg/errors` package). |

### Side Effects

The function performs a read‑only GET request against the Kubernetes API server.  
No state is mutated on the client side, and no resources are created or deleted.

### How It Fits in the Package

* **Package**: `autodiscover` – responsible for automatically discovering pod sets, CRDs, and related scaling information across a cluster.
* **Role**: This helper abstracts away the details of how to obtain a Scale object for any CR. Other parts of the package (e.g., `autodiscover_podset.go`) call it when they need to understand or modify replica counts without hard‑coding specific resource types.
* **Flow Example**  
  ```go
  scale, err := FindCrObjectByNameByNamespace(client.Scales(), ns, crName, myResource)
  if err != nil { /* handle */ }
  replicas := scale.Spec.Replicas
  ```

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Caller] -->|Passes| B(FindCrObjectByNameByNamespace)
    B --> C[getter.Scales()]
    C --> D[GET /apis/<group>/v1/namespaces/<ns>/<resource>/<name>/scale]
    D --> E[scalingv1.Scale]
    E --> F[Caller]
```

This diagram illustrates the single API call that the function performs to retrieve scaling information for any CR.
