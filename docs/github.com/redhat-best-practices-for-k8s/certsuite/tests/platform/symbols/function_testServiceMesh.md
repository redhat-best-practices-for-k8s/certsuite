testServiceMesh`

| | |
|---|---|
| **Package** | `platform` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/platform`) |
| **Visibility** | unexported (used only inside the test suite) |
| **Signature** | `func (*checksdb.Check, *provider.TestEnvironment)` |

## Purpose

`testServiceMesh` is a helper invoked by the platform test suite to verify that the Service Mesh
(typically Istio) has been correctly installed in the cluster under test.  
It performs a minimal sanity check:

1. Confirms the presence of an Istio‑sidecar proxy on every pod.
2. Records a report object for each pod that passes or fails this verification.

The function does **not** return a value; instead, it records its findings directly into the supplied
`*checksdb.Check` instance by calling `SetResult`.

## Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `c` | `*checksdb.Check` | The test result container to which the outcome of this check will be written. |
| `env` | `*provider.TestEnvironment` | Holds context about the current test environment (e.g., the Kubernetes client, namespace). |

> **Note:** Both parameters are required; the function panics if either is `nil`.

## Workflow

1. **Log start** – `LogInfo("Starting Service Mesh check")`.
2. **Determine proxy type** – Calls `IsIstioProxy()` to decide which pod label selector to use.
3. **Iterate over pods**  
   * For each pod that should contain a sidecar:
     * Create a new report object via `NewPodReportObject(env, pod.Name)`.
     * If the pod has an Istio proxy (`IsIstioProxy()` returns true), log success; otherwise log error.
     * Append the report to an internal slice (`append(pods, podReport)`).
4. **Set overall result** – After all pods are processed, `c.SetResult(true/false)` is called based on whether any pod failed the check.

## Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Structured logging for test progress and failures. |
| `IsIstioProxy()` | Determines if Istio is expected; influences pod selection logic. |
| `NewPodReportObject(env, podName)` | Builds a per‑pod report structure to be stored in the check result. |
| `append` | Aggregates individual pod reports into a slice. |
| `SetResult` (method on `*checksdb.Check`) | Stores the final pass/fail status of the Service Mesh test. |

## Side Effects

- **Logging**: Emits informational or error messages to the test log.
- **State mutation**: Updates the supplied `Check` instance with a result flag and pod reports.

No global state is modified; the function relies solely on its arguments and the package‑level logging helpers.

## How It Fits the Package

The `platform` test suite orchestrates end‑to‑end validation of a Kubernetes cluster’s compliance with Red Hat best practices.  
`testServiceMesh` is one of several internal helper functions that each perform a distinct check (e.g., checking for RBAC policies, validating network policy enforcement).  
By feeding its outcome into the `checksdb.Check` object, it integrates seamlessly with the test framework’s reporting and aggregation mechanisms.

---

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Start] --> B{IsIstioProxy?}
    B -- Yes --> C[Select Istio pods]
    B -- No  --> D[Select non‑Istio pods]
    C & D --> E{Pod has sidecar?}
    E -- Yes --> F[Log success, create report]
    E -- No  --> G[Log error, create report]
    F & G --> H[Append to slice]
    H --> I[Loop next pod]
    I --> J[All pods processed?]
    J -- No --> E
    J -- Yes --> K[SetResult(true/false)]
    K --> L[End]
```

This diagram visualizes the decision flow within `testServiceMesh`.
