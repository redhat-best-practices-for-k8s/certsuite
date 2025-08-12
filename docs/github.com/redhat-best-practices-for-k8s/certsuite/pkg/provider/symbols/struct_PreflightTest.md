PreflightTest` – Lightweight Result Wrapper

**File:** `pkg/provider/provider.go:185`  
**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`

---

## Overview
`PreflightTest` is a *value type* used throughout the provider package to convey the outcome of an individual pre‑flight check performed on a Kubernetes cluster.  
It is intentionally minimal: it stores only human‑readable data that can be logged, displayed in reports, or surfaced via the CLI/API.

> **Why a struct instead of primitives?**  
> The tests are often executed concurrently and aggregated later; using a struct makes passing metadata (name, description, remediation) along with the error status straightforward without relying on external maps or global state.

---

## Fields

| Field | Type   | Purpose |
|-------|--------|---------|
| `Name`        | `string` | Identifier of the test (e.g., `"kubelet-ssl"`). Used as a key in result maps and for sorting. |
| `Description` | `string` | Human‑readable explanation of what the test verifies. Displayed in reports. |
| `Error`       | `error`  | The actual error returned by the underlying check. A `nil` value indicates success. |
| `Remediation` | `string` | Suggested action to fix a failure (often a command or configuration snippet). Empty if the test passed. |

All fields are exported so callers can construct and read results directly.

---

## Typical Usage Flow

```mermaid
flowchart TD
    A[Run preflight check] --> B{Success?}
    B -- Yes --> C[Return PreflightTest{Name, Description, nil, ""}]
    B -- No  --> D[Return PreflightTest{Name, Description, err, remediation}]
```

1. **Invocation** – Each provider implements a function that returns `PreflightTest`.  
2. **Execution** – The check performs its logic; if it fails, it creates an error and supplies remediation text.  
3. **Aggregation** – The test runner collects many `PreflightTest` instances into a slice or map for reporting.

---

## Key Dependencies & Side Effects

| Dependency | Effect |
|------------|--------|
| `error` interface | Allows any Go error type; callers usually cast to custom error types if needed. |
| No external packages in this struct – it is self‑contained. |

The struct itself has **no side effects**; all logic resides in the functions that generate it.

---

## Where It Appears

* **Provider implementations** (`pkg/provider/*`) create `PreflightTest` values for each check they expose.  
* **Result aggregation utilities** (e.g., in `cmd/cli/report.go`) iterate over slices of `PreflightTest`.  
* **Unit tests** assert that the fields are populated correctly.

---

## Summary

- *Purpose*: Encapsulate a single pre‑flight test result with metadata and remediation.  
- *Inputs*: None directly; created by provider functions.  
- *Outputs*: Returned to callers, then aggregated for reporting.  
- *Dependencies*: Only Go’s standard `error` type.  
- *Side Effects*: None – purely data holder.

This struct is the cornerstone of how CertSuite communicates test outcomes across the provider layer and into higher‑level reporting components.
