PolicyIsRT`

| Item | Detail |
|------|--------|
| **Signature** | `func PolicyIsRT(policy string) bool` |
| **Exported?** | Yes |
| **Location** | `pkg/scheduling/scheduling.go:147` |

## Purpose

`PolicyIsRT` determines whether a given CPU scheduling policy name corresponds to the **Real‑Time (RT)** family of policies.  
In the context of the *certsuite* project, this helper is used when configuring or validating container runtimes that must run with RT guarantees (e.g., `SCHED_FIFO`, `SCHED_RR`).

## Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `policy`  | `string` | The name of a CPU scheduling policy as it appears in user‑land tools or configuration files. |

The function accepts any string; it performs a case‑insensitive match against known RT policy names.

## Outputs

| Return | Type   | Description |
|--------|--------|-------------|
| `bool` | `true` if `policy` is an RT policy (`SCHED_FIFO`, `SCHED_RR`, or variants), otherwise `false`. |

The function never panics and always returns a boolean value.

## Key Dependencies

* **None** – The implementation uses only the Go standard library (string comparison, case conversion).
* It relies on the convention that RT policies are represented by the names `"fifo"`, `"rr"` or their uppercase equivalents.  
  These names are hard‑coded in the function body; no external constants or configuration files are involved.

## Side Effects

The function is pure: it does not read or modify any global state, file system, network, or process attributes.  
It simply returns a value based on its input.

## How It Fits Into the Package

`PolicyIsRT` lives in the `scheduling` package, which provides utilities for:

* Interacting with container runtime scheduling mechanisms (`GetProcessCPUSchedulingFn`, `CrcClientExecCommandContainerNSEnter`).
* Defining common constants representing CPU scheduling policies (e.g., `SharedCPUScheduling`, `ExclusiveCPUScheduling`).
* Validating or converting policy names supplied by users or configuration files.

When higher‑level code needs to branch on whether a requested scheduling policy should be treated as a real‑time policy, it calls `PolicyIsRT`.  
For example:

```go
if scheduling.PolicyIsRT(requestedPolicy) {
    // Apply RT‑specific configuration or validation.
}
```

Thus, `PolicyIsRT` serves as a small, reusable helper that abstracts the logic of identifying RT policies across the project.
