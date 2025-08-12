Check` – A single testable requirement

| Field | Type | Purpose |
|-------|------|---------|
| `ID` | `string` | Human‑readable identifier used in logs and result tables. |
| `Labels` | `[]string` | Tags that can be matched against a label expression to filter checks. |
| `SkipMode` | `skipMode` | Strategy for evaluating the slice of `SkipCheckFns`.  `skipModeAll` means *all* must return false; `skipModeAny` (default) aborts on the first true. |
| `SkipCheckFns` | `[]func() (bool, string)` | Optional functions that decide whether to skip this check and optionally provide a reason. |
| `BeforeCheckFn` / `AfterCheckFn` | `func(*Check) error` | Hooks run immediately before/after the main test logic. |
| `CheckFn` | `func(*Check) error` | The actual requirement implementation. |
| `StartTime`, `EndTime` | `time.Time` | Wall‑clock timestamps marking when the check started and finished. |
| `Timeout` | `time.Duration` | Optional maximum run time; if exceeded a timeout error is recorded. |
| `Result` | `CheckResult` | Enumerated status (`Passed`, `Failed`, `Skipped`, `Aborted`, `Errored`). |
| `Error` | `error` | If non‑nil, the check panicked or returned an error. |
| `CapturedOutput` | `string` | Captured stdout/stderr from the test run (used in reporting). |
| `details` | `string` | Human‑readable description of what happened; usually set by `SetResult`. |
| `skipReason` | `string` | Reason returned by a skip function, shown in reports. |
| `abortChan` | `chan string` | Channel used to signal an external abort; closed or sent on by `Abort()`. |
| `logArchive` | `*strings.Builder` | Internal buffer that stores all log messages for later retrieval via `GetLogs()`. |
| `logger` | `*log.Logger` | Per‑check logger that writes into `logArchive`; created by `NewCheck`. |
| `mutex` | `sync.Mutex` | Protects concurrent access to mutable fields (`Result`, `Error`, logs, etc.). |

## Key Methods

| Method | Signature | What it does |
|--------|-----------|--------------|
| `Run()` | `func() error` | Executes the check lifecycle: <br>1. Records start time.<br>2. Runs `BeforeCheckFn`. <br>3. Evaluates skip functions (`ShouldSkip`). <br>4. If not skipped, runs `CheckFn` (with timeout support). <br>5. Runs `AfterCheckFn`. <br>6. Captures end time and logs result via `recordCheckResult()`. |
| `SetResult(ok []*testhelper.ReportObject, err []*testhelper.ReportObject)` | `func()` | Transforms internal report objects into a string summary stored in `details` and updates `Result`. Logs warning if counts differ. |
| `SetResultSkipped(reason string)`, `SetResultError(err string)`, `SetResultAborted(abortMsg string)` | `func(string)` | Convenience helpers that lock the struct, set `Result` appropriately, and optionally log a message. |
| `Abort(msg string)` | `func(string)` | Signals an abort by writing to `abortChan`. Also triggers panic with `AbortPanicMsg`, allowing callers to recover. |
| Logging helpers (`LogDebug`, `LogInfo`, `LogWarn`, `LogError`, `LogFatal`) | `func(string, ...any)` | Wrap the internal logger’s `Logf` with a level prefix; `LogFatal` also writes to stderr and exits. |
| `GetLogger()` / `GetLogs()` | `*log.Logger` / `string` | Accessors for external inspection or custom formatting. |
| Chainable configurators (`WithCheckFn`, `WithBeforeCheckFn`, etc.) | `func(func(*Check) error)` or similar | Return the same `Check` after setting a field, enabling fluent construction. |

## How it fits the package

* **Single responsibility** – Each `Check` represents one CNF‑cert requirement test.
* **Group orchestration** – The `ChecksGroup` type aggregates many `Check`s and orchestrates their execution order (before/after hooks, abort handling).  
  `RunChecks()` in `ChecksGroup` iterates over the slice of `Check`, calls each’s `Run()`, and records results with `recordCheckResult`.
* **Result aggregation** – The `Check.Result` value is used by higher‑level reporting logic (e.g., `PrintCheckPassed/Failed/...`).  
  The string summaries from `SetResult` are later included in CSV or JSON output.
* **Thread safety** – All mutable state changes go through the struct’s mutex, so a check can be run concurrently with others.

## Typical usage

```go
chk := NewCheck("cnf-001", []string{"critical"})
chk.WithCheckFn(func(c *check.Check) error {
    // perform test logic here
    return nil
}).WithTimeout(30*time.Second)

if err := chk.Run(); err != nil {
    log.Printf("check %s failed: %v", chk.ID, err)
}
```

The returned `error` from `Run()` is only non‑nil if the caller explicitly aborts or a panic occurs; normal test failures are recorded in `chk.Result`.

---

### Mermaid diagram (optional)

```mermaid
flowchart TD
    A[NewCheck] --> B{Configure}
    B -->|WithCheckFn| C[Set CheckFn]
    B -->|WithTimeout| D[Set Timeout]
    C --> E[Run()]
    E --> F{BeforeCheckFn}
    F --> G{Skip?}
    G -- yes --> H[Mark Skipped & Log]
    G -- no --> I[Execute CheckFn]
    I --> J[AfterCheckFn]
    J --> K[Record Result]
```

This diagram visualises the core execution path of a `Check`.
