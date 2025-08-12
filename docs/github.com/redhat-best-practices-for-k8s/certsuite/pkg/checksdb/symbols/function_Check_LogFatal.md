Check.LogFatal`

```go
func (c Check) LogFatal(msg string, args ...any) ()
```

### Purpose
`LogFatal` is a helper that writes an error message to the standard error stream and then terminates the process immediately with exit code 1.  
It is intended for fatal conditions that cannot be recovered from while executing a check (e.g., missing configuration or critical runtime failure).

> **Why not just `log.Fatal`?**  
> The method prefixes the log line with the check’s own identifier (`c.Name`) and uses the same formatting logic as the regular logging functions in this package.  It also ensures that any buffered logs are flushed before exiting.

### Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `msg`     | `string` | Format string, identical to `fmt.Printf`. |
| `args...` | `…any`    | Arguments to be substituted into the format string. |

The method does not return a value; its signature is `func(string, ...any)()`.

### Behaviour & Side‑Effects

1. **Log the message**  
   * Calls `c.Logf("%s: "+msg+"\n", args...)` – this writes to `stderr` with the check name as a prefix and respects any log‑level filtering configured elsewhere in the package.
2. **Print to stdout for visibility**  
   * Uses `fmt.Fprintf(os.Stderr, msg+"\n", args...)`.  This guarantees that the fatal message appears even if the standard logger is redirected or buffered.
3. **Terminate process**  
   * Calls `os.Exit(1)` immediately after printing.  No deferred functions run; no cleanup occurs.

### Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `c.Logf` | Formats and writes the prefixed log line. |
| `fmt.Fprintf` | Directly prints the raw fatal message to `stderr`. |
| `os.Exit` (imported as `Exit`) | Stops execution with exit code 1. |

No global variables are accessed or modified.

### Relationship within `checksdb`

* **Checks** – The method is a receiver on the `Check` struct, which represents an individual test in the checks database.  
* **Error handling strategy** – Other functions such as `Check.Run()` use `LogFatal` when encountering unrecoverable errors (e.g., missing label evaluator).  
* **Package design** – Keeps fatal error reporting consistent across all checks and centralizes exit logic, simplifying unit testing of non‑fatal paths.

### Example

```go
func (c Check) Run() {
    if c.Labels == nil {
        // Fatal: cannot evaluate labels without an evaluator.
        c.LogFatal("label evaluator not configured")
    }
}
```

In this example, the process will log the error and terminate immediately.
