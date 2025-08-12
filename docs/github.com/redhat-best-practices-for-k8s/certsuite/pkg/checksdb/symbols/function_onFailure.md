onFailure` – internal helper for handling check failures

| Item | Details |
|------|---------|
| **Signature** | `func (msg string, errMsg string, group *ChecksGroup, chk *Check, other []*Check) error` |
| **Visibility** | unexported (`onFailure`) – used only inside the `checksdb` package. |
| **Purpose** | Centralises the logic that is executed when a check fails, errors or is aborted. It logs the failure, updates the result state of the affected checks and, depending on the group’s skip‑mode, may automatically skip the rest of the checks in the same group. |

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `msg` | `string` | Human‑readable message that will be printed to the console (via `Printf`). |
| `errMsg` | `string` | Error string passed to `SetResultError` – typically the underlying error from the failing check. |
| `group` | `*ChecksGroup` | The group that owns the check that just failed. |
| `chk` | `*Check` | The specific check that has failed (or aborted). |
| `other` | `[]*Check` | Optional slice of *additional* checks that should also be marked as skipped if the group’s skip‑mode is “any”. This can contain checks that were not yet executed when the failure occurred. |

### Return value

- `error`: returns the error created by `New(errMsg)` (the same string passed to `SetResultError`).  
  The function never returns `nil`; it always propagates the original failure.

### Key dependencies & calls

| Called | Purpose |
|--------|---------|
| `Printf` | Prints the failure message (`msg`) to stdout. |
| `chk.SetResultError(errMsg)` | Records the error state on the failing check. |
| `skipAll(group, other)` | If the group’s `SkipMode` is `SkipModeAny`, this marks **all** remaining checks in the group (and any passed `other` checks) as skipped. |
| `New(errMsg)` | Wraps the failure message into an `error`. |

### Side‑effects

1. **Logging** – prints a formatted message describing the failure.
2. **State mutation** – updates the failing check’s result to *Error* via `SetResultError`.
3. **Optional skip cascade** – may mark other checks as skipped if the group configuration requires it.
4. **Returns an error** that can be bubbled up by the caller.

### How it fits into the package

The `checksdb` package manages a collection of *check groups* (`ChecksGroup`). Each group contains multiple checks (`Check`). During execution, if any check fails (or is aborted), the surrounding logic calls `onFailure`. This helper:

- Keeps failure handling DRY by centralising logging and state updates.
- Enforces the configured *skip mode* semantics for a group.
- Provides an error that can be used to stop further processing or propagate the failure up the call stack.

```mermaid
flowchart TD
    A[Check execution] -->|Failure| B(onFailure)
    B --> C{SkipModeAny}
    C -- yes --> D(skipAll(...))
    B --> E(SetResultError)
    B --> F(Printf)
    E --> G(Return error)
```

**Note:**  
The function is deliberately kept private because its behaviour is tightly coupled to the internal data structures (`ChecksGroup`, `Check`) and the group‑level skip logic. It should only be invoked from within this package where those invariants are guaranteed.
