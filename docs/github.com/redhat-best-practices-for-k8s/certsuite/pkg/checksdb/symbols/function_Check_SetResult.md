Check.SetResult`

| Item | Details |
|------|---------|
| **Package** | `checksdb` – the central registry of all certificate‑suite checks |
| **Receiver** | `c Check` – a single check definition (identified by its ID, group, etc.) |
| **Signature** | `func(c *Check) SetResult(tests []*testhelper.ReportObject, errors []*testhelper.ReportObject) func()` |

### Purpose
`SetResult` records the outcome of running a particular check.  
It writes the provided test and error objects into the check’s internal slices and updates the overall status flags (`Passed`, `Failed`, etc.). The method returns a closure that should be executed **after** all dependent logic (e.g., logging) has finished; this pattern allows callers to defer side‑effects until they are certain no further mutation will occur.

### Parameters
| Param | Type | Meaning |
|-------|------|---------|
| `tests` | `[]*testhelper.ReportObject` | Objects representing individual sub‑tests that passed. |
| `errors` | `[]*testhelper.ReportObject` | Objects representing failures or errors encountered during the check. |

### Return value
A function of type `func()` which, when invoked, performs:
1. **Thread‑safe update** – locks `dbLock`, writes to the check’s slices (`c.tests`, `c.errors`) and updates status counters.
2. **Result stringification** – converts the test objects into a single string via `ResultObjectsToString`.
3. **Logging** – emits either an error log (if any errors were recorded) or a warning if the number of tests does not match expectations.

### Key dependencies
- **Synchronization** – uses the package‑wide `dbLock` mutex (`Lock()` / `Unlock()`) to serialize updates across goroutines.
- **Helpers**  
  - `ResultObjectsToString` formats test objects for logging.  
  - `LogError`, `LogWarn` write messages to the shared logger.  
- **Standard library** – built‑in `len` to check slice lengths.

### Side effects
* Mutates the receiver’s `tests` and `errors` slices.  
* Updates status booleans (`Passed`, `Failed`, `Skipped`) on the `Check`.  
* Emits log entries if inconsistencies are detected or errors occurred.

### How it fits in the package
The `checksdb` package maintains a global database of checks. Each check may be executed concurrently by various test runners. `SetResult` is the canonical entry point for reporting results back to that database, ensuring consistency and thread safety. After all tests finish, callers invoke the returned closure (often via `defer`) so that logging happens after any remaining concurrent activity.

### Example flow
```go
c := dbByGroup["groupA"].Checks[0]      // obtain a Check
defer c.SetResult(tests, errors)()     // schedule result handling
// ... run sub‑tests, fill `tests` and `errors`
```

This pattern guarantees that the check’s status is finalized only after all test logic has completed.
