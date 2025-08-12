nameInDeploymentSkipList`

### Purpose
`nameInDeploymentSkipList` determines whether a given deployment should be **excluded** from scaling‑related lifecycle tests.  
The test suite contains a list of deployments that are known to cause flaky or unsupported behavior when the number of replicas is changed during the test run. This helper inspects that list and reports if the supplied deployment name matches any entry.

### Signature
```go
func(name string, namespace string, skipList []configuration.SkipScalingTestDeploymentsInfo) bool
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | `string` | The Kubernetes **deployment** name under test. |
| `namespace` | `string` | The namespace in which the deployment resides. |
| `skipList` | `[]configuration.SkipScalingTestDeploymentsInfo` | Slice containing objects that specify deployments to skip, each of which includes a `Name` and optional `Namespace`. |

### Return Value
- **`true`** – if the `(name, namespace)` pair matches an entry in `skipList`.
- **`false`** – otherwise.

The function performs a straightforward linear search; it does not modify any state.

### Key Dependencies
* **`configuration.SkipScalingTestDeploymentsInfo`** – the struct that defines each skip rule.  
  It is defined elsewhere in the test configuration package and contains at least the fields `Name` (string) and `Namespace` (string, optional).

No other globals or external packages are accessed.

### Side Effects
None. The function is pure: it only reads its arguments and returns a boolean.

### How It Fits Into the Package

In the `lifecycle` test suite (`suite.go`) scaling tests are run against many deployments.  
Before initiating a scaling operation, the test code calls `nameInDeploymentSkipList` to decide whether to skip that deployment:

```go
if nameInDeploymentSkipList(deployName, deployNamespace, env.SkipScalingDeployments) {
    t.Skipf("Skipping %s/%s per configuration", deployNamespace, deployName)
}
```

By centralising the skip logic in this helper, the suite remains easy to maintain and can be extended simply by updating the `skipList` supplied from test configuration.
