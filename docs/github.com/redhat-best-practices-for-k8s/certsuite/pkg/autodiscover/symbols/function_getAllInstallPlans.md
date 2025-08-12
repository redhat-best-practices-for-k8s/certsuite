getAllInstallPlans`

```go
func (k OperatorsV1alpha1Interface) getAllInstallPlans() []*olmv1Alpha.InstallPlan
```

### Purpose
`getAllInstallPlans` is a helper that collects **every** `InstallPlan` resource present in the cluster for which the caller has read access.  
In the context of the *autodiscover* package this list is later used to determine which Operator Lifecycle Manager (OLM) operators are installed and their current status.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `k` | `v1alpha1.OperatorsV1alpha1Interface` | The client interface that provides access to OLM resources (`InstallPlans`, `CatalogSources`, etc.). It is typically a typed Kubernetes client created by the package’s discovery logic.

### Outputs
- **Slice of pointers**: `[]*olmv1Alpha.InstallPlan`
  - Each element represents one InstallPlan found in the cluster.
  - The slice may be empty if no install plans exist or an error occurs during listing.

### Key Steps & Dependencies

| Step | Operation | Dependency |
|------|-----------|------------|
| 1 | Call `k.InstallPlans("").List(...)` | `InstallPlans` method from the OLM client interface; `v1.ListOptions{}` (default). |
| 2 | Iterate over returned list (`installPlanList.Items`) | `olmv1Alpha.InstallPlan` struct. |
| 3 | Append each item’s pointer to a slice | Go built‑in `append`. |

The function is intentionally simple – it does **not** filter by namespace or label, and it ignores errors silently (it just logs them with `logrus.Error`). This design keeps the helper lightweight; higher‑level code can decide how to handle missing install plans.

### Side Effects
- Logs an error message if listing fails (`logrus.Error`).
- No mutation of cluster state; purely read‑only.

### How It Fits the Package
Within `autodiscover/autodiscover_operators.go`, this helper is called by other routines that need a global view of all operator install plans, such as:

```go
installPlans := getAllInstallPlans(olmClient)
for _, ip := range installPlans {
    // evaluate status, gather CSV information, etc.
}
```

Because `autodiscover` aims to automatically detect which operators are present and what certificates they expose, having a complete list of install plans is essential for correlating operator names with their corresponding ClusterServiceVersion (CSV) resources.

### Example Usage

```go
olmClient := // obtain v1alpha1.OperatorsV1alpha1Interface from kubeconfig
installPlans := getAllInstallPlans(olmClient)

for _, ip := range installPlans {
    fmt.Printf("Operator %s installed via plan %s\n", ip.Spec.CSV, ip.Name)
}
```

> **Note**: The function is unexported; it is intended for internal package use only.

--- 

*This documentation summarizes the role and mechanics of `getAllInstallPlans` within the autodiscover package.*
