getAtLeastOneSubscription`

```go
func getAtLeastOneSubscription(
    op *Operator,
    csv *olmv1Alpha.ClusterServiceVersion,
    subs []olmv1Alpha.Subscription,
    pkgManifests []*olmpkgv1.PackageManifest,
) bool
```

### Purpose  
Determines whether the given **Cluster Service Version (CSV)** can be considered valid for the test run by checking that **at least one** of its available subscriptions is satisfied by a package manifest in the cluster.  
If no subscription matches, the CSV is ignored and an error message is logged.

This helper is used by the operator discovery logic to filter out CSVs that cannot be installed because none of their required packages are present.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `op` | `*Operator` | The operator being inspected. Used only for logging (`op.log.Error`). |
| `csv` | `*olmv1Alpha.ClusterServiceVersion` | The CSV whose subscriptions are examined. |
| `subs` | `[]olmv1Alpha.Subscription` | All subscriptions listed in the CSV’s spec. |
| `pkgManifests` | `[]*olmpkgv1.PackageManifest` | Package manifests that have been loaded from the cluster. They represent available packages that could satisfy a subscription. |

### Return Value

- **`true`** – at least one subscription can be satisfied by a package manifest.
- **`false`** – no subscriptions match any package manifest; the CSV is considered unusable.

### Key Dependencies & Calls

| Called function | Role |
|-----------------|------|
| `getPackageManifestWithSubscription(pkgManifests, sub)` | Searches `pkgManifests` for a package that satisfies the given subscription. Returns the matching `*olmpkgv1.PackageManifest` and a boolean flag. |
| `Error()` (method of `op.log`) | Emits an error message when none of the subscriptions match any package. |

The function itself performs only pure logic: iterating over subscriptions, invoking the helper search, and logging if necessary. It does **not** modify any state outside its parameters.

### Side‑Effects

- Logs a single error line per CSV that fails to find a matching subscription.
- No mutation of global variables or passed objects.

### How it fits the package

`provider/operators.go` contains the logic for discovering and validating Operators in an OpenShift cluster.  
During discovery, each CSV is evaluated:

1. All its subscriptions are collected (`csv.Spec.InstallStrategy.StrategySpec.Spoke.OperatorGroupConfig.Subscriptions`).
2. `getAtLeastOneSubscription` checks whether any of those subscriptions can be satisfied by the current set of package manifests.
3. If **false**, the CSV (and thus the operator) is skipped for further tests.

Thus, this function acts as a gatekeeper that ensures only operators with at least one viable subscription proceed to test execution.
