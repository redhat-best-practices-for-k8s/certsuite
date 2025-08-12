testPodHostPath`

*File:* `tests/accesscontrol/suite.go` – line 462  
*Package:* `accesscontrol`

---

## Purpose
`testPodHostPath` validates that no pod in the cluster exposes a **hostPath** volume when its `readOnly` flag is set to `true`.  
The check ensures that pods cannot read arbitrary host filesystem paths, which would violate least‑privilege principles.

> **Note:** The function name does not follow Go naming conventions for exported functions (`TestPodHostPath`). It is intentionally unexported because it is used only as a test helper inside the suite.

---

## Signature

```go
func testPodHostPath(check *checksdb.Check, env *provider.TestEnvironment) ()
```

| Parameter | Type                    | Description |
|-----------|------------------------|-------------|
| `check`   | `*checksdb.Check`      | The check definition that drives the test (contains metadata such as name and severity). |
| `env`     | `*provider.TestEnvironment` | Runtime context providing access to the Kubernetes client, logging facilities, and shared state. |

The function does **not** return a value; results are recorded via side‑effects on `check`.

---

## Key Steps

1. **Log start**  
   ```go
   LogInfo(env.Logger, "Running hostPath readOnly test")
   ```
   Provides visibility in the test output.

2. **Iterate over all Pods** (implementation details hidden in `env.Client`).  
   For each pod:
   * Inspect its volumes.
   * If a volume is of type `hostPath` and its `readOnly` flag is set to `true`, a *failure* report is generated.

3. **Report construction** – uses the helper `NewPodReportObject`:
   ```go
   podObj := NewPodReportObject()
   podObj.AddField("pod", pod.Name)
   podObj.SetType("hostPath")
   ```
   *When a violation is found:*  
   - Append the object to `check.FailedPods`.
   - Set `check.Result = checksdb.Failure`.

4. **Log errors** – if any step fails (e.g., unable to list pods), `LogError` records it and the function terminates early.

5. **Success case** – If no violations are found, `check.Result` remains `checksdb.Success`.

---

## Dependencies

| Dependency | Role |
|------------|------|
| `env.Logger` | Structured logger for diagnostics. |
| `NewPodReportObject()` | Factory that creates a report object with fields such as pod name and type. |
| `LogInfo`, `LogError` | Convenience wrappers around the logger. |
| `check.FailedPods` | Slice collecting all failing pod reports. |
| `checksdb.Check.Result` | Final status of the test (`Success`, `Failure`). |

The function does not modify global state except through the passed‑in `check`.  
It uses only read‑only data from `env`.

---

## Integration in the Test Suite

* `testPodHostPath` is invoked by the suite’s orchestrator (likely via a map of test functions keyed by check name).  
* It complements other pod‑related checks such as `testContainerSecurityContext`, ensuring comprehensive access‑control validation.  

The results are aggregated into the overall report for the *Access Control* test group.

---

## Summary

`testPodHostPath` is a focused helper that scans all pods, flags any hostPath volumes marked read‑only, and records failures in the supplied `checksdb.Check`.  
It relies on standard logging, report object construction, and the Kubernetes client provided by `provider.TestEnvironment`, making it a self‑contained unit of test logic within the access‑control package.
