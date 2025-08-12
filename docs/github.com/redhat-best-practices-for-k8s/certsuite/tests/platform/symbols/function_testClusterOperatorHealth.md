testClusterOperatorHealth`

```go
func testClusterOperatorHealth(check *checksdb.Check, env *provider.TestEnvironment)
```

| Item | Description |
|------|-------------|
| **Purpose** | Validates that all required OpenShift cluster operators are in the *Available* state before a certsuite run proceeds. It records any missing or unhealthy operators into the test report. |
| **Inputs** | `check` – the check definition that owns this test (used for reporting).<br> `env` – a test‑environment context providing access to cluster state and logging facilities. |
| **Outputs / Side‑effects** | *No return value.* The function writes diagnostic information directly into the provided `*checksdb.Check` via helper functions (`NewClusterOperatorReportObject`, `SetResult`). It also logs progress with `LogInfo`. No global state is mutated. |
| **Key dependencies** | - `IsClusterOperatorAvailable(env, operatorName)` – queries the cluster for a given operator’s status.<br>- `LogInfo(msg string)` – simple logger used throughout the suite.<br>- Report‑building helpers: `NewClusterOperatorReportObject`, `SetResult`. |
| **How it fits the package** | In the *platform* test suite, this function is invoked during the pre‑test “before each” hook (see `beforeEachFn`). It ensures that the OpenShift environment is ready for certificate tests. The function lives in `suite.go` and forms part of a collection of health checks that guard against mis‑configured or partially installed cluster operators. |

### Flow (in prose)

1. **Logging** – Announces the start of the operator health check via `LogInfo`.  
2. **Operator verification** – For each critical OpenShift operator (e.g., *ClusterVersion*, *Ingress*, *Network*), it calls `IsClusterOperatorAvailable` to see if that operator reports status `Available`.  
3. **Report construction** – For every operator, a report object is created with `NewClusterOperatorReportObject`, appended to the check’s result list, and its status set accordingly (`SetResult`).  
4. **Completion** – Once all operators are processed, the function returns, leaving the check populated for later assertion by the test harness.

> **Note:** The exact list of operators and any thresholds (e.g., timeout) are defined elsewhere in the suite; this function merely orchestrates the checks and populates the report.
