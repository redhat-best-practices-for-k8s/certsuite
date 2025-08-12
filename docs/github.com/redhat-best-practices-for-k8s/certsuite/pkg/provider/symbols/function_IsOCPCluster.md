IsOCPCluster`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`  
**Signature:**  

```go
func IsOCPCluster() bool
```

### Purpose

Determines whether the current Kubernetes cluster is an **OpenShift Container Platform (OCP)** installation. The function returns `true` when the cluster exhibits characteristics that are unique to OCP, and `false` otherwise.

> *Note:* The exact detection logic is not visible in the provided snippet; it likely inspects cluster‑wide resources or labels that are only present on OpenShift (e.g., presence of the `openshift.io` namespace, specific deployment names, or the `node-role.kubernetes.io/infra` label).  

### Inputs / Outputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| *none*    | —      | The function relies solely on global state (e.g., cached cluster information) and does not accept external arguments. |

| Return value | Type  | Meaning |
|--------------|-------|---------|
| `bool`       | `true` | Cluster is identified as OpenShift. |
|              | `false`| Cluster is *not* identified as OpenShift (likely vanilla Kubernetes). |

### Key Dependencies

- **Global state:** The function may read from package‑level variables such as:
  - `env`: holds environment configuration that might contain a flag like `IS_OCP`.
  - `loaded`: indicates whether cluster metadata has been fetched.
- **External resources:** It could query the API server for OCP‑specific objects (e.g., `Route`, `BuildConfig`) or inspect node labels such as those in `MasterLabels`/`WorkerLabels`.

### Side Effects

None. The function is read‑only and does not mutate any state.

### Context within the Package

The `provider` package provides abstractions over the underlying cluster (OpenShift vs vanilla Kubernetes).  
`IsOCPCluster()` is used by other components to:

- Enable or disable OCP‑specific checks.
- Adjust label handling for master/worker nodes (`MasterLabels`, `WorkerLabels`).
- Configure provider behavior based on whether the cluster supports OpenShift features.

Because it returns a boolean, callers can gate logic with simple conditionals such as:

```go
if provider.IsOCPCluster() {
    // run OCP‑only tests
} else {
    // fall back to generic Kubernetes checks
}
```

---

**Summary:** `IsOCPCluster` is a helper that inspects the current cluster’s characteristics and reports whether it is an OpenShift installation. It plays a central role in routing provider logic between OpenShift‑specific and generic Kubernetes paths.
