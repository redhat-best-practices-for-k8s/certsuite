Tester.Run` – Core Test Execution Entry Point

## Overview
`Tester.Run` is the public entry point for executing all *hugepages* tests bundled in this package.  
It orchestrates three sub‑tests:

1. **Systemd‑managed hugepages** (`TestNodeHugepagesWithMcSystemd`)
2. **Kernel‑argument configured hugepages** (`TestNodeHugepagesWithKernelArgs`)

The method returns an `error` that aggregates any failures from the sub‑tests.

## Signature
```go
func (t Tester) Run() error
```
* **Receiver**: `Tester` – a struct holding state for this test run.  
  *Its fields are not shown here, but they typically include configuration and logging helpers.*
* **Returns**: `error` – non‑nil if any sub‑test fails.

## Flow Diagram (Mermaid)

```mermaid
flowchart TD
    A[Start] --> B{HasMcSystemdHugepagesUnits()}
    B -- true --> C[TestNodeHugepagesWithMcSystemd()]
    B -- false --> D[Log Info]
    C --> E{Error?}
    E -- yes --> F[Return Error]
    E -- no --> G
    G --> H{TestNodeHugepagesWithKernelArgs() ?}
    H -- error --> I[Return Error]
    H -- success --> J[Success]
```

## Detailed Steps

| Step | Operation | Notes |
|------|-----------|-------|
| 1 | **Check Systemd support** (`HasMcSystemdHugepagesUnits`) | Determines if the node uses `systemd` to manage hugepages. |
| 2 | **Log status** (`Info`) | If systemd is present, log that the test will run; otherwise skip it and note skipping. |
| 3 | **Run Systemd test** (`TestNodeHugepagesWithMcSystemd`) | Verifies that hugepages are correctly configured via systemd units. |
| 4 | **Handle error** (`Errorf`) | If this sub‑test fails, log the error and return it immediately. |
| 5 | **Log status** (`Info`) | Announces start of kernel‑argument test. |
| 6 | **Run Kernel‑args test** (`TestNodeHugepagesWithKernelArgs`) | Checks that hugepage size/quantity are correctly passed via boot arguments. |
| 7 | **Handle error** (`Errorf`) | On failure, log and return the error. |
| 8 | **Return nil** | All tests succeeded. |

## Key Dependencies

- `HasMcSystemdHugepagesUnits()` – Detects systemd configuration.
- `TestNodeHugepagesWithMcSystemd()` – Validates systemd‑managed hugepage units.
- `TestNodeHugepagesWithKernelArgs()` – Validates kernel boot arguments for hugepages.
- Logging helpers: `Info`, `Errorf`.

## Side Effects

- Writes to the test logger via `Info` and `Errorf`.
- No state mutation on the `Tester` receiver (pure read‑only run).
- May exit early if a sub‑test fails, propagating that error.

## Package Context

The **hugepages** package implements platform‑level checks for Kubernetes nodes.  
`Run` is invoked by higher‑level test harnesses to validate hugepage support under different configuration mechanisms (systemd vs kernel args). It ensures the node complies with best practices before proceeding with workload tests that rely on hugepages.

---
