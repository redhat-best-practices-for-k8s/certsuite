testDeploymentScaling`

```go
func testDeploymentScaling(env *provider.TestEnvironment,
                            timeout time.Duration,
                            check *checksdb.Check) func()
```

### Purpose

`testDeploymentScaling` is a **lifecycle‑check helper** used by the CertSuite tests to validate that a Kubernetes Deployment can be scaled safely.  
The function returns another function (a closure) that will be executed by the test harness at run time.  The returned routine performs the following steps:

1. Marks the check as needing refresh (`SetNeedsRefresh`).
2. Logs diagnostic information about the deployment under test.
3. Validates ownership and owner references of the Deployment’s pods.
4. Checks whether the Deployment is in a skip list or is not managed by CertSuite; if so, it records a skipped result.
5. Retrieves any associated HPA (Horizontal Pod Autoscaler) for the Deployment and, if present, runs `TestScaleHpaDeployment`.
6. If no HPA exists, falls back to `TestScaleDeployment` which performs a manual scale‑up/down test.
7. Records success or failure in the check report (`SetResult`).

The routine does **not** modify the environment except for the check’s state; it only reads from Kubernetes and logs via the provided logger.

### Inputs

| Parameter | Type                     | Description |
|-----------|--------------------------|-------------|
| `env`     | `*provider.TestEnvironment` | Test context that exposes a K8s client, logger, and report objects. |
| `timeout` | `time.Duration`          | Maximum time allowed for scaling operations. |
| `check`   | `*checksdb.Check`        | The check record to update with results and logs. |

### Output

A **no‑argument closure** (`func()`) that, when invoked, executes the scaling test logic described above.

### Key Dependencies

- **Provider / Environment**
  - `env.Client` – Kubernetes API client.
  - `env.Logger` – Structured logger for diagnostics.
  - `env.Report` – Object used to create and store report entries (`NewDeploymentReportObject`, `SetResult`).

- **Helper Functions**
  - `SetNeedsRefresh(env, check)` – Marks the check as needing a refresh before execution.
  - `LogInfo / LogError` – Logging utilities.
  - `ToString(obj)` – Serialises Kubernetes objects for logging.
  - `IsManaged(check, obj)` – Determines if CertSuite manages the resource.
  - `CheckOwnerReference`, `GetOwnerReferences` – Validate pod ownership.
  - `nameInDeploymentSkipList(name string) bool` – Checks against a skip list.
  - `GetResourceHPA(env, deploymentName)` – Retrieves any HPA tied to the Deployment.
  - `TestScaleHpaDeployment(...)` and `TestScaleDeployment(...)` – Perform actual scaling tests.

- **Constants**
  - `intrusiveTcSkippedReason`, `localStorage`, etc. (used in logging and skip logic).

### Side Effects

- **State Changes**  
  The function may update the check record with a result (`SetResult`) and logs entries via `env.Logger`. No Kubernetes resources are created or deleted.

- **Global Impact**  
  None beyond the local check context; it operates purely on the passed environment.

### Integration in Package

The `lifecycle` package contains end‑to‑end tests for various lifecycle scenarios (deployment, statefulset, pod set).  
`testDeploymentScaling` is invoked by test cases that iterate over all Deployments discovered in a cluster.  It encapsulates the logic for scaling verification so that each deployment can be tested consistently without duplicating code across multiple test functions.

```go
for _, dep := range deployments {
    env.Checks.AddCheck(checksdb.NewDeploymentScaling(dep))
    env.Run(testDeploymentScaling(env, timeout, check))
}
```

Thus, `testDeploymentScaling` serves as the core routine that drives deployment‑scaling validation in CertSuite’s lifecycle tests.
