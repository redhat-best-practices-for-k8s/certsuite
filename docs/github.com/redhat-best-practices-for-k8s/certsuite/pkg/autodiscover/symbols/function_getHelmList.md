getHelmList`

```go
func getHelmList(cfg *rest.Config, namespaces []string) map[string][]*release.Release
```

### Purpose  
`getHelmList` retrieves all Helm releases that are **deployed** in the given Kubernetes namespaces.  
It is used by the autodiscover logic to identify operators that were installed via Helm so that they can be processed later (e.g., for certificate discovery).

---

### Parameters  

| Name | Type | Description |
|------|------|-------------|
| `cfg` | `*rest.Config` | The REST client configuration used to create a Helm client. It typically comes from the Kubernetes client‑config (`kubeconfig`). |
| `namespaces` | `[]string` | A slice of namespace names in which to search for Helm releases. Empty or nil means “all namespaces”. |

---

### Return Value  

```go
map[string][]*release.Release
```

* **Key** – Namespace name (string).  
* **Value** – Slice of pointers to `helm.sh/helm/v3/pkg/release.Release` objects that are in the *Deployed* state for that namespace.  

If no releases exist for a namespace, that key will either be omitted or map to an empty slice.

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `NewClientFromRestConf(cfg)` | Builds a Helm client from the provided Kubernetes REST config. This client knows how to query the Helm release store (the ConfigMap/Secret backend). |
| `ListDeployedReleases(client, namespaces)` | Core helper that performs the actual listing of releases for the supplied namespaces and filters by status `Deployed`. |
| `panic(err)` | If creating the Helm client fails, the function panics. This is intentional because autodiscover expects a working client; failure indicates a mis‑configured environment. |

---

### Side Effects

* **Panics** on error while creating the Helm client.
* No mutation of global state or external resources.

The function is read‑only: it only queries existing data and returns a new map.

---

### How It Fits the Package  

`autodiscover` orchestrates detection of various operator deployments (Helm, OLM, Operators Hub, etc.).  
`getHelmList` is the low‑level routine that supplies the Helm‑specific part of this detection:

1. The package first collects a list of namespaces to inspect.
2. It calls `getHelmList` to get all deployed releases per namespace.
3. The returned map feeds into higher‑level logic (`processHelmRelease`, etc.) that interprets each release’s chart metadata and decides whether it is an operator the suite should interact with.

Because Helm operators are a common installation path on OpenShift/K8s clusters, this helper centralises all Helm client setup and querying logic in one place.  

---

### Suggested Mermaid Diagram (for package flow)

```mermaid
graph TD
    A[autodiscover] --> B[getHelmList]
    B --> C[helm.NewClientFromRestConf]
    B --> D[ListDeployedReleases]
    D --> E{release.Status == Deployed}
    E --> F[return map[string][]*Release]
```

This diagram shows the flow from the package entry point to the Helm client creation, release listing, filtering by status, and returning the result.
