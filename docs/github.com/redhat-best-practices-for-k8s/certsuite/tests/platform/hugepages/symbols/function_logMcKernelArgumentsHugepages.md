logMcKernelArgumentsHugepages`

### Purpose
`logMcKernelArgumentsHugepages` is an internal helper used by the **hugepages** test package to emit diagnostic information about the current kernel configuration related to huge pages.  
It logs, via the test framework’s logger (`Info`), the mapping of *kernel parameter → value* that was read from `/proc/cmdline` or a similar source. The function is called during test setup so that failures can be traced back to incorrect kernel arguments.

### Signature
```go
func logMcKernelArgumentsHugepages(mc map[int]int, total int) func()
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `mc`      | `map[int]int` | A map where the key is an integer constant representing a kernel parameter (e.g., `HugepagesParam`, `HugepageszParam`) and the value is the parsed numeric value for that parameter. |
| `total`   | `int`  | The total number of huge pages requested or observed, typically derived from the map or system state. |

**Return value:**  
A closure (`func()`) that performs the actual logging when invoked. This design allows callers to defer the logging until after the test has fully initialized.

### Implementation details

1. **String building** – Two separate `bytes.Buffer` objects are created (via `WriteString` and `Sprintf`).  
   * The first buffer collects a human‑readable list of the kernel arguments that were processed (`HugepagesParam`, `HugepageszParam`, …).  
   * The second buffer compiles a summary line that shows the total number of huge pages.

2. **Logging** – The function calls `Info` (from the test framework’s logger) to write both buffers:
   ```go
   log.Info().Msg(buf.String())
   ```
   This results in two log entries: one for the individual parameters, another for the aggregate value.

3. **No side effects** – Apart from writing to the logger, the function does not mutate `mc`, `total`, or any package globals. It is purely observational.

### Dependencies

| Dependency | Role |
|------------|------|
| `bytes.Buffer` (via `WriteString`) | Builds log strings efficiently. |
| `fmt.Sprintf` | Formats integer values into string form. |
| `log.Info()` (from the test framework) | Emits structured logs. |
| `log.String()` | Helper to convert a buffer’s content into a log field. |

### How it fits the package

The **hugepages** package contains tests that validate kernel huge‑page configuration on various distributions.  
During setup, kernel arguments are parsed and stored in a map.  
`logMcKernelArgumentsHugepages` is invoked after this parsing to provide clear visibility into what was actually read from the kernel, aiding debugging when a test fails.

```mermaid
flowchart TD
    A[Parse /proc/cmdline] --> B[Build map[int]int (mc)]
    B --> C{Is logging enabled?}
    C -- Yes --> D[Call logMcKernelArgumentsHugepages(mc,total)]
    D --> E[Log individual params]
    D --> F[Log total huge pages]
```

### Summary

- **What it does** – Logs the parsed kernel arguments and their total huge‑page count.  
- **Inputs** – Map of parameter IDs to values, plus a total count.  
- **Output** – A closure that writes two log entries when called.  
- **Side effects** – None other than logging.  
- **Place in codebase** – Utility for test diagnostics within the hugepages package.
