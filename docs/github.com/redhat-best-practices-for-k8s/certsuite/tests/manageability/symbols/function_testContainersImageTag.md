testContainersImageTag` – Compliance Check for Container Image Tags  

**File:** `suite.go` (line 71)  
**Package:** `manageability`  

---

## Purpose
Validates that every container in the test environment specifies an explicit image tag.  
If a container’s image reference lacks a tag, it is considered **non‑compliant**; otherwise it is **compliant**.

The function populates the provided `*checksdb.Check` with a result status and detailed report objects for each container that passes or fails the test.

---

## Signature
```go
func testContainersImageTag(c *checksdb.Check, env *provider.TestEnvironment)
```
| Parameter | Type | Description |
|-----------|------|-------------|
| `c` | `*checksdb.Check` | The check record to be updated with result and report data. |
| `env` | `*provider.TestEnvironment` | Test environment that contains the list of containers under inspection. |

---

## Dependencies
| Dependency | Role |
|------------|------|
| `LogDebug`, `LogError`, `LogInfo` | Structured logging for debugging, error reporting, and informational messages. |
| `IsTagEmpty(img string) bool` | Helper that returns true if the supplied image reference has no tag. |
| `NewContainerReportObject(name, namespace, status string) *checksdb.ContainerReportObject` | Creates a report object describing an individual container’s compliance status. |
| `c.SetResult(status checksdb.CheckStatus)` | Stores the final status of the check (PASS/FAIL). |

---

## Algorithm Overview

1. **Iterate over all containers** in `env`.  
2. For each container:
   * If `IsTagEmpty(container.Image)` returns `true`, create a *non‑compliant* report object and append it to `c.ReportObjects`.
   * Otherwise, create a *compliant* report object and append it.
3. After processing all containers:
   * If any non‑compliant objects exist → set check result to **FAIL**.
   * Else → set check result to **PASS**.

---

## Side Effects
- The function mutates the supplied `checksdb.Check` by adding report objects and setting its status.  
- It emits log messages at debug, info, or error levels depending on encountered conditions.

---

## How it fits the package

The `manageability` package contains a suite of compliance checks that run against a Kubernetes environment.  
Each check is implemented as an unexported function like `testContainersImageTag`, invoked by higher‑level orchestration logic (e.g., a test runner).  

This particular check ensures best practices for container image specification, which is critical for reproducibility and security in deployments. The resulting report objects can be consumed by UI layers or exported to compliance dashboards.
