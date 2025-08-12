Check.SetResultSkipped`

### Purpose
`SetResultSkipped` marks a check as **skipped** in the test results database.  
It records a user‑supplied message explaining why the check was skipped and updates the
check’s internal state accordingly.

### Receiver
```go
func (c *Check) SetResultSkipped(msg string)
```
The method operates on a pointer to `Check`, which represents an individual test
within the checks database. The `Check` struct holds fields such as:

| Field | Description |
|-------|-------------|
| `result` | Current status (`CheckResultPassed`, `Failed`, etc.) |
| `message` | Optional message attached to the result |
| `lastResultTime` | Timestamp of the last state change |

(Other fields exist but are not relevant for this method.)

### Parameters
* **msg** – A string that will be stored in the check’s `message` field.  
  It can contain a short reason, a reference to a bug tracker, or any context
  useful when reviewing test results.

### Return value
The function returns *nothing*.  
Its signature is `func(string)()`, but the empty return type indicates that the
caller does not receive a value; the method simply mutates the check’s state.

### Side‑effects & synchronization

1. **Mutex protection** – The package owns a global `sync.Mutex` named `dbLock`
   (see `globals.dbLock`).  
   `SetResultSkipped` locks this mutex at entry and unlocks it before returning,
   ensuring that concurrent updates to the checks database are serialized.

2. **State mutation** – Within the critical section:
   * `c.result` is set to `CheckResultSkipped`.
   * `c.message` receives the supplied `msg`.
   * `c.lastResultTime` is updated to the current time (`time.Now()`).

3. **No external calls** – The method does not call other package functions or
   rely on global state beyond the mutex lock.

### Interaction with the rest of the package

* The set of possible result constants (`CheckResultPassed`, `Failed`, etc.) is
  defined in `check.go`.  
  `SetResultSkipped` uses the `CheckResultSkipped` constant to mark the status.
* Results are later aggregated by functions such as `PrintResults()` or
  exported via HTTP endpoints. Skipped checks appear with a distinct marker (e.g.,
  “SKIPPED”) and their message is displayed alongside the check name.
* The global maps (`dbByGroup`, `resultsDB`) store pointers to `Check` objects;
  mutating a `Check` instance automatically updates the stored data because
  they are referenced by pointer.

### Usage pattern

```go
// Suppose we have a Check instance c
if someCondition {
    // Skip this check with an explanatory message
    c.SetResultSkipped("Skipping due to missing dependency")
}
```

Afterwards, any reporting function will show that `c` was skipped and display the
provided message.

---

**Mermaid diagram (suggestion)**

```mermaid
flowchart TD
    A[Caller] -->|calls SetResultSkipped(msg)| B(Check.SetResultSkipped)
    B --> C{Lock dbLock}
    C --> D[c.result = CheckResultSkipped]
    D --> E[c.message = msg]
    E --> F[c.lastResultTime = now()]
    F --> G[Unlock dbLock]
```

This diagram illustrates the critical section guarded by `dbLock` and the
state changes performed on the `Check` instance.
