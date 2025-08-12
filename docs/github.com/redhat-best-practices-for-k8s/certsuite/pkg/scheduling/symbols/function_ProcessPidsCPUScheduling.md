ProcessPidsCPUScheduling`

### Overview
`ProcessPidsCPUScheduling` evaluates the CPU scheduling settings of all processes that belong to a given container.  
It queries each process for its scheduling policy and priority, maps those values onto the human‑readable strings defined in this package (e.g. `"Exclusive"`, `"Isolated"`), then creates a `ReportObject` for every process that fails to meet the expected configuration.

The function is part of the **scheduling** package, which provides helpers for verifying Kubernetes container runtimes against security best practices.  It is typically called by higher‑level test suites that iterate over all containers in a pod.

---

### Signature
```go
func ProcessPidsCPUScheduling(
    processes []*crclient.Process,
    container *provider.Container,
    name string,
    logger *log.Logger,
) []*testhelper.ReportObject
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `processes` | `[]*crclient.Process` | List of process objects extracted from the container’s namespace. Each element contains fields such as PID, policy, and priority. |
| `container` | `*provider.Container` | The container being evaluated; used for setting report metadata (`SetContainerProcessValues`). |
| `name` | `string` | Name of the rule / test that is being executed (e.g., `"CPU scheduling"`). |
| `logger` | `*log.Logger` | Logger instance used to emit debug, info, and error messages. |

**Return value**

An array of pointers to `testhelper.ReportObject`.  
Each object represents a failure for a specific process.  If all processes satisfy the rule, the slice is empty.

---

### How it Works

1. **Logging start** – Emits a debug message indicating that CPU scheduling checks are beginning.
2. **Iterate over each process**
   * Calls `GetProcessCPUSchedulingFn` to translate the numeric policy/priority values into their string equivalents (`policyString`, `priorityString`).  
     - If translation fails, an error is logged and the function skips that process.
3. **Comparison against expectations**
   * Expected values are derived from global constants:
     ```go
     CurrentSchedulingPolicy  // e.g., "RoundRobin"
     CurrentSchedulingPriority // e.g., "0"
     ```
   * The actual `policyString` / `priorityString` is compared with the expected ones.
4. **Report generation for mismatches**
   * If a mismatch occurs:
     - A new `ReportObject` is created via `NewContainerReportObject`.
     - The container and process metadata are attached using `SetContainerProcessValues`.
     - A descriptive message (e.g., `"PID 123: Expected policy RoundRobin but got Exclusive"`) is assembled with `fmt.Sprint`.
     - The report object is appended to the result slice.
5. **Return** – After all processes have been processed, the accumulated failures are returned.

---

### Dependencies

| Dependency | Purpose |
|------------|---------|
| `Debug`, `Info`, `Error` (logger) | Emit diagnostic messages. |
| `GetProcessCPUSchedulingFn` | Translates numeric policy/priority to strings. |
| `SetContainerProcessValues` | Attaches container/process context to a report object. |
| `NewContainerReportObject` | Creates a new reporting structure for a failure. |
| `fmt.Sprint` | Builds human‑readable error messages. |

---

### Side Effects & State

* **No mutation of input data** – The function only reads from the supplied slices and structs.
* **Creates report objects** – These are new allocations that capture failures; they are returned to the caller for aggregation or printing.
* **Logs information** – Uses the provided logger, so callers must supply a configured `log.Logger`.

---

### Placement in the Package

The `scheduling` package focuses on verifying runtime settings such as CPU scheduling, isolation levels, and priority.  
`ProcessPidsCPUScheduling` is the core routine that implements the *CPU scheduling* rule; it is invoked by test orchestration code (e.g., a pod‑level runner) after processes have been enumerated.

A simplified call flow:

```
ContainerRunner
   └─> gatherProcesses() → []*crclient.Process
          |
          v
    ProcessPidsCPUScheduling(processes, container, "CPU scheduling", logger)
          |
          v
   []ReportObject  (collected failures)
```

---

### Example Usage

```go
reports := scheduling.ProcessPidsCPUScheduling(
    procs,
    cont,
    "CPU scheduling",
    log.New(os.Stdout, "", log.LstdFlags),
)

for _, r := range reports {
    fmt.Println(r.Message) // e.g., “PID 42: Expected policy RoundRobin but got Exclusive”
}
```

---

### Summary

`ProcessPidsCPUScheduling` is a read‑only validator that:

1. Inspects each process in a container.
2. Checks its CPU scheduling policy and priority against the package defaults.
3. Reports any deviations as structured `ReportObject`s.

It ties together the low‑level process data (`crclient.Process`) with the high‑level reporting API (`testhelper.ReportObject`), enabling automated compliance checks for Kubernetes containers.
