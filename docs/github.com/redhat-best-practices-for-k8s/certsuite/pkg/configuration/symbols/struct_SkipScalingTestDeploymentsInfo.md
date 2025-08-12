SkipScalingTestDeploymentsInfo`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration` |
| **Exported?** | Yes (`exported: true`) |
| **Purpose** | Holds the identifying information of a deployment that should be excluded from the scaling‑test suite. This prevents known problematic deployments (e.g., those with custom autoscaling logic or resource constraints) from causing false positives or test failures. |

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | The Kubernetes **deployment name** to skip. |
| `Namespace` | `string` | The namespace where the deployment resides. If omitted, tests may default to a standard namespace or treat it as global (implementation‑dependent). |

#### How It Is Used

1. **Configuration Load**  
   During package initialization, `configuration.go` reads a configuration file (typically JSON/YAML) that contains an array of these structs. The loader unmarshals each entry into a `SkipScalingTestDeploymentsInfo`.

2. **Runtime Filtering**  
   The scaling test harness iterates over all deployments in the cluster and checks each against the skip list:
   ```go
   for _, dep := range allDeployments {
       if shouldSkip(dep, skipList) { continue }
       runScaleTest(dep)
   }
   ```
   The helper `shouldSkip` compares both `Name` and `Namespace`.

3. **Side Effects**  
   - No direct mutation of cluster state; purely advisory.
   - Logging may emit a notice when a deployment is skipped to aid debugging.

#### Dependencies

- **Configuration Loader** – relies on the package’s YAML/JSON unmarshalling utilities (`encoding/json`, `gopkg.in/yaml.v2`).
- **Test Harness** – uses this struct via the global skip list exposed by `configuration.Load()` or similar accessor functions.
- **Logging** – optional integration with a logger to record skipped deployments.

#### Placement in the Package

The struct lives in `configuration.go`, which also defines:

```go
type Config struct {
    SkipScalingTestDeployments []SkipScalingTestDeploymentsInfo `json:"skip_scaling_test_deployments"`
}
```

Thus, it is part of the overall configuration schema that governs test execution behavior.

#### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Configuration File] -->|JSON/YAML| B[config.Load()]
    B --> C{Parse into Config}
    C --> D[SkipScalingTestDeploymentsInfo List]
    D --> E[Test Harness]
    E --> F[Iterate Deployments]
    F --> G{Is deployment in skip list?}
    G -- Yes --> H[Skip Test]
    G -- No --> I[Run Scaling Test]
```

This diagram illustrates the data flow from configuration to test execution, highlighting where `SkipScalingTestDeploymentsInfo` fits.
