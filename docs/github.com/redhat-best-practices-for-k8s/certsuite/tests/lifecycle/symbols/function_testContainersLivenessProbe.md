testContainersLivenessProbe`

| Attribute | Value |
|-----------|-------|
| **Package** | `lifecycle` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle`) |
| **Visibility** | unexported (private) |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |
| **Location** | `suite.go:307` |

### Purpose
`testContainersLivenessProbe` validates that the containers in a Kubernetes workload expose a working liveness probe.  
The function is executed as part of a larger test suite for lifecycle‑related checks. It records two sub‑reports:

1. **All containers** – whether every container has a non‑empty `livenessProbe`.
2. **Containers without readiness** – among those that lack a readiness probe, whether each still has a liveness probe.

If any container fails the check, the overall result is set to *Failed* and an error message is logged; otherwise the result remains *Passed*.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `check` | `*checksdb.Check` | Holds metadata for the test case (ID, description, etc.) and will receive the final status via `SetResult`. |
| `env` | `*provider.TestEnvironment` | Provides access to the Kubernetes client, the namespace under test, and helper methods for querying workloads. |

### Workflow

1. **Log start** – `LogInfo` prints “testing liveness probe …”.
2. **Collect all containers**  
   * Calls `getAllContainersFromEnv(env)` (implementation hidden in this snippet).  
   * Builds a `ContainerReportObject` named `"All containers"` and appends it to the check’s report.
3. **Check every container** – iterates over each container; if any lacks a liveness probe, logs an error, marks the overall result as failed, and breaks.
4. **Collect containers without readiness**  
   * Calls `getContainersWithoutReadinessProbe(env)` (implementation hidden).  
   * Builds a second report object named `"Containers without readiness"`.
5. **Check those containers** – same loop logic as step 3 but for the filtered list.
6. **Set final result** – if no failures were detected, the check remains in its default `Passed` state; otherwise it has already been set to *Failed*.

### Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Logging framework for test progress and errors. |
| `NewContainerReportObject` | Creates a container‑level report entry attached to the main check. |
| `SetResult` | Updates the overall status of the test (`checksdb.ResultPassed/Failed`). |
| Helper functions (e.g., `getAllContainersFromEnv`) | Abstract away Kubernetes API interactions and filtering logic. |

### Side Effects

* Adds two container‑level report objects to the passed `check`.
* May set the check’s result to *Failed*.
* Emits log messages but does **not** modify any external state (e.g., cluster resources).

### Integration with Package

Within the `lifecycle` test package, this function is invoked by a higher‑level orchestrator that iterates over all registered checks. It complements other container health tests such as readiness probes and lifecycle hook verifications, collectively ensuring that workloads adhere to Kubernetes best practices.

> **Note**: The actual Kubernetes interactions (`getAllContainersFromEnv`, `getContainersWithoutReadinessProbe`) are defined elsewhere in the package; they query the test environment’s namespace for pods belonging to the workload under test.
