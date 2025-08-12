testCrds` – Observability CRD Status Checker

## Overview
`testCrds` is a private helper used by the **observability** test suite to validate that every Custom Resource Definition (CRD) deployed in the test environment exposes a `status` sub‑resource.  
The function runs as part of the `ChecksDB` framework, which orchestrates multiple checks and collects their results.

## Signature
```go
func (*checksdb.Check, *provider.TestEnvironment)
```
* **Receiver** – a pointer to a `Check` from the `checksdb` package; it is used only for logging.
* **Parameter** – a pointer to `provider.TestEnvironment`, which represents the runtime Kubernetes cluster being tested.

The function does not return a value; it records its result directly into the passed‑in `Check`.

## Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | The check object that will receive status updates. |
| `env` | `*provider.TestEnvironment` | Provides access to the Kubernetes client used for querying CRDs. |

No other external inputs are required.

## Execution Flow

1. **Logging** – An informational log marks the start of the test.
2. **CRD Enumeration**  
   * The function queries all installed CRDs via `env.ClientSet.ApiextensionsV1().CustomResourceDefinitions()`.
3. **Status Sub‑resource Verification**  
   For each CRD, it inspects the `Spec.PreserveUnknownFields` flag and the presence of a `status` sub‑resource (`Subresources.Status != nil`).  
   * If the status sub‑resource is missing, a report object is created with:
     * `Name`: CRD name
     * `Result`: `Failed`
4. **Reporting** – All failures are aggregated into the check’s result set using `SetResult`.
5. **Logging** – A final log records completion of the test.

## Key Dependencies

| Dependency | Why it matters |
|------------|----------------|
| `env.ClientSet` | Enables interaction with the Kubernetes API to list CRDs. |
| `checksdb.Check` | Provides logging and result aggregation facilities. |
| `NewReportObject`, `AddField`, `SetResult` | Helper functions for building structured test reports. |

These dependencies are all local to the test package; no external services or global state are modified.

## Side Effects

* **State mutation** – The function updates the passed‑in `Check` with failure records, but otherwise leaves the environment unchanged.
* **Logging** – Emits informational and error logs that appear in the test output.

No global variables are read or written.

## Package Context

The `observability` package contains end‑to‑end tests for certificate management workloads.  
`testCrds` is invoked during the observability check phase to ensure CRDs have proper status sub‑resources, which is essential for operator health reporting and reconciliation loops.

---

### Suggested Mermaid Diagram
```mermaid
flowchart TD
  A[Start] --> B[Log start]
  B --> C[List CRDs via env.ClientSet]
  C --> D{CRD has status?}
  D -- Yes --> E[Continue]
  D -- No --> F[Create report object (Failed)]
  F --> G[Add to Check result]
  E --> H[End]
```
This diagram illustrates the decision path for each CRD and how failures are recorded.
