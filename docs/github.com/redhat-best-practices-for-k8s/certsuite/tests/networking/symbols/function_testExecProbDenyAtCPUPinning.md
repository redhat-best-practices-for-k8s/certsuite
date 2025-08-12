testExecProbDenyAtCPUPinning`

**Location**

```
github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/suite.go:171
```

### Purpose

`testExecProbDenyAtCPUPinning` is an **internal test helper** used by the networking suite to verify that a pod’s *exec* probe configuration does not violate CPU‑pinning restrictions.  
It performs a quick sanity check:

1. Ensures the pod has an exec probe (`HasExecProbes`).  
2. Builds a report object for each offending probe and records the result as a failure.

The function is **not exported**; it is invoked only by the test framework when running the networking tests.

### Signature

```go
func (*checksdb.Check, []*provider.Pod)()
```

| Parameter | Type                 | Description |
|-----------|----------------------|-------------|
| `check`   | `*checksdb.Check`    | The check configuration that drives the test. (Unused in this helper but part of the common signature for all test helpers.) |
| `pods`    | `[]*provider.Pod`    | Slice of pods inspected by the test. |

The function returns **nothing**; it mutates the pod report objects inside the supplied slice.

### Key Operations

1. **Logging**
   ```go
   LogInfo("Checking for exec probes that might violate CPU‑pinning")
   ```
   Records a high‑level informational message in the test log.

2. **Probe Presence Check**
   ```go
   if !HasExecProbes(pods) { … }
   ```
   - If no exec probes are present, logs an error and exits early.
   - Otherwise proceeds to collect offending probes.

3. **Report Construction**
   ```go
   report := NewPodReportObject()
   report.SetResult(...)
   ```
   For each pod that contains a problematic exec probe, the helper creates a `PodReportObject` (via `NewPodReportObject`) and populates it with a failure result (`SetResult`).  
   The resulting reports are appended back to the original slice of pods.

4. **Error Handling**
   - Uses `LogError` when prerequisites (e.g., absence of exec probes) are not met.
   - No panic or recover; all errors are reported through logging and result objects.

### Dependencies

| Dependency | Role |
|------------|------|
| `HasExecProbes` | Detects whether any pod in the slice defines an exec probe. |
| `NewPodReportObject` | Instantiates a report container for a pod. |
| `SetResult` | Records the test outcome (pass/fail) on the report object. |
| Logging helpers (`LogInfo`, `LogError`) | Emit structured log messages. |

These helpers are defined elsewhere in the networking test package and operate on shared types such as `provider.Pod` and `checksdb.Check`.

### Side‑Effects

- **Mutates** the supplied `pods` slice by appending new report objects.
- Emits log entries via the package’s logging system.
- No external state changes (no file I/O, no network calls).

### How It Fits the Package

Within the `networking` test suite, each test helper follows a standard signature (`func(*checksdb.Check, []*provider.Pod)()`).  
`testExecProbDenyAtCPUPinning` is part of a series of checks that validate probe configurations against best‑practice rules.  
Its result feeds into the overall test report, allowing CI pipelines to surface CPU‑pinning violations early.

--- 

**Mermaid suggestion (optional)**

```mermaid
flowchart TD
  A[Pods] --> B{HasExecProbes?}
  B -- No --> C[LogError & Exit]
  B -- Yes --> D[Iterate Pods]
  D --> E[Create PodReportObject]
  E --> F[SetResult(Fail)]
  F --> G[Append to pods slice]
```
