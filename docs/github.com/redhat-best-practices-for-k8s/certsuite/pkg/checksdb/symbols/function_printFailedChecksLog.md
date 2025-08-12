printFailedChecksLog`

### Purpose
`printFailedChecksLog` is a *private* helper that returns a zero‑argument function used to output the accumulated logs of all checks that failed during a test run.  
The returned function is typically invoked via `defer` so that it executes after the current request has finished, guaranteeing that the failure log is printed even if the caller panics or returns early.

### Signature
```go
func printFailedChecksLog() func()
```

- **Input:** none.
- **Output:** a closure of type `func()` which, when called, prints detailed information about each failed check to standard output.

### Key Behavior

1. **Header & Separator**  
   - Prints the title `"Failed checks log:"` followed by a line of dashes whose length matches the header width (`len(header)`).
2. **Iterate Over Groups**  
   - Iterates over all `ChecksGroup`s stored in the global map `dbByGroup`.  
   - For each group, it calls the group's `GetLogs()` method (a method defined on `*ChecksGroup` that aggregates logs from its constituent checks). The call is made *without* locking; callers are responsible for ensuring thread safety.
3. **Print Group Logs**  
   - Prints the group name and a line of dashes underneath it, then prints each log entry returned by `GetLogs()`.  
4. **Footer**  
   - Prints a final newline to separate this section from any following output.

### Dependencies

| Dependency | Role |
|------------|------|
| `fmt.Sprintf`, `fmt.Println`, `strings.Repeat` | String formatting and console output |
| `unicode/utf8.RuneCountInString` | Determines the width of the header for proper dashes |
| `ChecksGroup.GetLogs()` | Retrieves per‑group log strings |

### Side Effects

- **Console Output** – The function writes to `os.Stdout`.  
- **No mutation** – It does not modify any global state; it only reads from `dbByGroup`.

### How It Fits the Package

The `checksdb` package manages a database of checks and their execution results.  
During test runs, each check may produce logs that are stored in its owning group.  
When a test completes, callers typically defer a call to this function:

```go
defer printFailedChecksLog()()
```

This guarantees that all accumulated failure logs are displayed once the request handling is finished, aiding debugging and auditability.

### Suggested Mermaid Diagram

```mermaid
graph TD;
    A[Caller] -->|defer| B(printFailedChecksLog);
    B --> C{Iterate dbByGroup};
    C --> D[GetLogs() on ChecksGroup];
    D --> E[Print header & logs];
```

This diagram shows the flow from the caller to the deferred printing of failure logs.
