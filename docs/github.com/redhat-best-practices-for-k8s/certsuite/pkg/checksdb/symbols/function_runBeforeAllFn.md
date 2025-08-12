runBeforeAllFn`

```go
func runBeforeAllFn(g *ChecksGroup, checks []*Check) error
```

### Purpose  
`runBeforeAllFn` executes the **before‑all** hook for a group of checks (`*ChecksGroup`).  
A before‑all function is meant to perform one‑time setup that all checks in the group share (e.g. creating temporary resources, loading configuration).  
The helper handles panic recovery and propagates any error returned by the hook.

### Parameters  

| Name   | Type            | Description |
|--------|-----------------|-------------|
| `g`    | `*ChecksGroup`  | The group whose before‑all function is to be run. |
| `checks` | `[]*Check` | The list of checks belonging to the group (used only for logging). |

### Return value  

- `error`:  
  - `nil` if the hook ran successfully.  
  - An error describing a panic or an explicit failure returned by the hook.

### Key Steps & Dependencies

| Step | Code | Explanation |
|------|------|-------------|
| 1 | `defer func() { ... }()` | A deferred function recovers from any panic that occurs inside the before‑all hook. It logs the stack trace (`Stack`) and returns a wrapped error using `onFailure`. |
| 2 | `g.beforeAllFn(checks)` | Calls the actual hook stored in the group. The hook receives the slice of checks for context. |
| 3 | Handle returned error | If the hook returns an error, it is logged (`Debug`) and re‑wrapped with `onFailure` before being propagated. |

The function uses several helpers from the package:

- **`Debug`, `Error`** – logging utilities.
- **`Stack`** – captures stack trace for panic diagnostics.
- **`onFailure`** – converts an error into a standard failure representation (used elsewhere in the package).
- **`beforeAllFn`** – the field of `ChecksGroup` that holds the hook function.

### Side‑effects

- Logs debug information about the execution and any errors.  
- Does not modify global state (`dbByGroup`, `resultsDB`, etc.) – it only interacts with the passed group and checks.  
- The panic recovery ensures the caller never crashes; instead an error is returned.

### How it fits in the package

`runBeforeAllFn` is a private helper used by the checks execution engine when preparing to run all checks in a `ChecksGroup`. It guarantees that:

1. All checks share a common setup routine.
2. Errors or panics in this routine are surfaced as a single failure for the group, allowing the rest of the suite to continue safely.

The function is invoked from higher‑level orchestration code (e.g., when executing a test plan) and forms part of the lifecycle management of checks within `checksdb`.
