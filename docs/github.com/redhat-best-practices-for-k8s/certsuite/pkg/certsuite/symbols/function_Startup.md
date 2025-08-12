Startup` – package initialisation routine

## Purpose
`Startup` performs all one‑time setup required by the **certsuite** command line tool before any tests are executed.  
It:

1. Parses and validates command‑line / environment configuration (`GetTestParameters`).  
2. Configures label filtering logic (`InitLabelsExprEvaluator`).  
3. Sets up logging (creates a global log file).  
4. Loads the test checks database.  
5. Prints diagnostic information – banner, version, parameters, etc.

It returns a `func()` that should be called at program exit to perform any necessary cleanup (currently only flushing logs).

---

## Signature
```go
func Startup() func()
```

* **Returns**: A closure for teardown; the caller typically defers this call.

---

## Key Steps & Dependencies

| Step | Action | Called Functions |
|------|--------|------------------|
| 1 | Read test parameters | `GetTestParameters()` |
| 2 | Initialise label expression evaluator | `InitLabelsExprEvaluator()` |
| 3 | Create global log file | `CreateGlobalLogFile(logFileName)` |
| 4 | Load checks database | `LoadChecksDB()` |
| 5 | Print diagnostic banner & version info | `PrintBanner()`, `GitVersion()` |
| 6 | Log environment details (OS, Go version, env vars) | `Info()`, `Printf()` |
| 7 | Return teardown closure | – |

### Global constants used
- `claimFileName` – name of the file where test claims are stored.  
- `collectorAppURL` – URL for metrics/collector integration.  
- `junitXMLOutputFileName` – output path for JUnit XML reports.  
- `noLabelsFilterExpr` – default label‑filter expression when none supplied.  
- `timeoutDefaultvalue` – default timeout used by tests.

---

## Side Effects

| Effect | Description |
|--------|-------------|
| **Logging** | A global log file is created (`CreateGlobalLogFile`). All subsequent logs are written to this file. |
| **Error handling** | If any initialization step fails, the function prints an error and calls `os.Exit(1)` – terminating the program immediately. |
| **Environment awareness** | The routine reads environment variables (e.g., `CERTSUITE_LOG_FILE`) for configuration; it also logs current OS/Go version. |
| **Resource cleanup** | The returned closure should be invoked on normal exit to flush log buffers or release resources. |

---

## How It Fits the Package

`certsuite` is a CLI tool that runs security checks against Kubernetes clusters.  
The `Startup` function must be called once at program start:

```go
func main() {
    defer certsuite.Startup()()
    // ... run tests ...
}
```

This guarantees that:
- Test parameters are available globally (`GetTestParameters()` is used by many other components).  
- Label filtering logic (used by test selection) is correctly configured.  
- All logs go to a predictable file and are flushed on exit.  
- The checks database is loaded into memory for quick access during tests.

Without calling `Startup`, the tool would lack configuration, logging, and data needed to execute any check.
