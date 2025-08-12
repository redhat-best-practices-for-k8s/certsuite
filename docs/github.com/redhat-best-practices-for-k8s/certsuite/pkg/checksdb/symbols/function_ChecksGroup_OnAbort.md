ChecksGroup.OnAbort`

> **Signature**  
> ```go
> func (c *ChecksGroup) OnAbort(msg string) error
> ```

`OnAbort` is a public method on the `ChecksGroup` type that implements the abort‑handling logic for all checks belonging to that group.  
When any check in a group fails with an **abort** result, this callback is invoked automatically by the framework.

### Purpose

* **Propagate the abort state** – mark every remaining (unexecuted) check in the group as *skipped* or *aborted* according to its `SkipMode`.
* **Record the reason** – store the supplied abort message (`msg`) on each affected check.
* **Emit a summary line** – print an informative log line that indicates the group has been aborted.

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `msg` | `string` | Human‑readable description of why the group was aborted (typically derived from the failing check’s error). |

### Return Value

* Returns an `error`.  
  The current implementation never returns a non‑nil value; it always returns `nil`.  
  The signature exists for compatibility with potential future extensions.

### Key Operations & Dependencies

| Step | Called Function | Effect |
|------|-----------------|--------|
| 1 | `c.printCheckResult` | Prints the abort message for each affected check (uses group‑level formatting). |
| 2 | `c.SetResultSkipped()` / `SetResultAborted()` | Updates the internal result status of individual checks. The choice depends on the check’s `SkipMode`:  
  * **SkipModeAll** – every remaining check is marked as *skipped*.  
  * **SkipModeAny** – only checks that are not yet executed get *aborted* while already‑executed ones remain unchanged. |
| 3 | `fmt.Printf` / `strings.ToUpper` | Formats the group‑wide abort notice shown in logs. |
| 4 | `Eval(labelsExprEvaluator)` | Evaluates any label expressions attached to checks that may affect skip logic (not directly visible in this snippet). |

### Side Effects

* **State mutation** – All pending checks inside the group have their result status changed, and they are recorded as *aborted* or *skipped*.  
  This prevents them from running again and influences final test results.
* **Logging** – Emits a single line to standard output indicating that the group was aborted with the provided message.  
  The log format includes the group name in uppercase for visibility.

### Integration into the Package

`OnAbort` is part of the *abort handling* mechanism used by `checksdb`.  
During normal execution, each check calls `c.OnCheckResult` to report its outcome; if a check returns `CheckResultAborted`, the framework internally triggers `ChecksGroup.OnAbort`.  

The method relies on shared globals:

| Global | Role |
|--------|------|
| `labelsExprEvaluator` | Evaluates label expressions that may influence skip logic. |
| `dbLock` / `dbByGroup` | Not directly used here, but part of the larger database context in which groups live. |

In summary, `OnAbort` is a thin wrapper that ensures a coherent abort state for all checks in a group and provides a clear log entry. It is a critical piece of the overall check execution flow, guaranteeing that once an abort condition occurs, no further checks in the same group are executed.
