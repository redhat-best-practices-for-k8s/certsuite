testHelmVersion`

`func(*checksdb.Check)()`

### Purpose
The function is a **private helper used by the certification test suite**.  
It validates that the Helm chart releases installed in the cluster match the
operator’s expectations and records any discrepancies as a *check result*.

### Inputs / Outputs

| Parameter | Type                 | Description |
|-----------|----------------------|-------------|
| `c`       | `*checksdb.Check`    | A pointer to the current test check.  The function updates this object with
information about Helm‑chart health and pod status, then calls `SetResult`
to mark the check as *pass* or *fail*. |

The function has no explicit return value; it mutates the supplied `Check`.

### Key Steps

1. **Acquire Kubernetes clients**  
   ```go
   client := GetClientsHolder()
   ```
   Retrieves a cached set of typed clients used throughout the test.

2. **List Helm releases**  
   ```go
   list, err := client.List(...)
   ```
   Enumerates all Helm release objects (the exact API type is not shown here).

3. **Validate existence and count**  
   * If no releases are found → `LogError` + `SetResult(fail)`.
   * If more than one release exists → logs a warning but still proceeds.

4. **Inspect each release’s pods**  
   For every Helm release, the function:
   - Calls `Pods()` via the core‑v1 client to fetch all associated pod objects.
   - Skips pods with status “Pending” or “Failed”.
   - Adds any non‑ready pods to a report.

5. **Report generation**  
   * Creates a `HelmChartReportObject` (via `NewHelmChartReportObject`) and
   attaches it to the check result.  
   * For each pod that fails readiness checks, creates a
     `PodReportObject` (`NewPodReportObject`) and appends it to the report.

6. **Finalize**  
   Calls `c.SetResult(...)` with the accumulated information.  
   The check is marked *pass* if all pods are ready; otherwise *fail*.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `GetClientsHolder` | Provides Kubernetes API clients. |
| `List`, `Pods`, `CoreV1` | Interact with the cluster to retrieve Helm releases and pod status. |
| `NewHelmChartReportObject`, `NewPodReportObject` | Build structured report objects for later consumption by the test harness. |
| `LogError`, `LogInfo` | Emit diagnostic information; side‑effect: writes to the test log. |
| `SetResult` | Mutates the supplied check object, marking pass/fail and attaching reports. |

### How It Fits Into the Package

The `certification` package orchestrates a suite of compliance checks against
a Kubernetes cluster.  
`testHelmVersion` is one of those checks; it runs after operators have been
validated (`skipIfNoOperatorsFn`) and Helm releases are present
(`skipIfNoHelmChartReleasesFn`).  
Its primary goal is to ensure that the operator’s Helm chart has produced a
healthy set of pods, thereby confirming correct deployment.

--- 

**Mermaid diagram (suggested)**

```mermaid
flowchart TD
    A[GetClientsHolder] --> B[List Helm Releases]
    B --> C{Releases Found?}
    C -- No --> D[LogError & SetResult(fail)]
    C -- Yes --> E[Iterate Releases]
    E --> F[Pods() via CoreV1]
    F --> G{Ready?}
    G -- Ready --> H[continue]
    G -- Not Ready --> I[Create PodReportObject]
    I --> J[Append to HelmChartReportObject]
    J --> K[SetResult(pass/fail)]
```

This function is a read‑only analyzer; it does not modify the cluster state,
only inspects and reports.
