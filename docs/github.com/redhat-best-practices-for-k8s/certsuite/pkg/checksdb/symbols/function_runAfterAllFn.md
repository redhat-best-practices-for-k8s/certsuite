runAfterAllFn`

| Feature | Details |
|---------|---------|
| **Location** | `pkg/checksdb/checksgroup.go:130` |
| **Visibility** | Unexported (`runAfterAllFn`) – used only inside the package. |
| **Signature** | `func(*ChecksGroup, []*Check) error` |

### Purpose
`runAfterAllFn` is a helper that executes an *after‑all* hook defined on a `ChecksGroup`.  
An after‑all function runs once all checks in the group have finished (regardless of success or failure). It allows the group to perform cleanup, aggregate results, or raise a final error.

### Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `group`   | `*ChecksGroup` | The group whose after‑all function should be invoked. |
| `checks`  | `[]*Check`      | Slice of all checks that belong to the group, in the order they were run. |

### Return value
- `error` – any error produced by the after‑all function or by the internal error handling logic.

### Execution Flow

```mermaid
flowchart TD
    A[Start] --> B{Check if group.afterAllFn exists}
    B -- No --> C[Return nil]
    B -- Yes --> D[Invoke group.afterAllFn(checks)]
    D --> E{panic?}
    E -- Yes --> F[Recover, log debug, call onFailure(err), return err]
    E -- No --> G{err returned by fn}
    G -- Non‑nil --> H[Call onFailure(err), return err]
    G -- nil --> I[Return nil]
```

1. **Skip early** – If the group has no `afterAllFn`, nothing is done and `nil` is returned.
2. **Invocation** – The hook is called with the slice of checks.
3. **Panic recovery** – Any panic inside the hook is caught, logged via `Debug`, wrapped in an error with stack trace, passed to `onFailure`, and returned.
4. **Error handling** – If the hook returns a non‑nil error, it is forwarded to `onFailure` and then returned.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `group.afterAllFn` | The user‑supplied function executed after all checks. |
| `Debug`, `Error` | Logging utilities used for debug output and error wrapping. |
| `onFailure` | Helper that records the failure in the group's state (not shown in snippet). |
| `recover` | Safeguards against panics within the hook. |

### Side Effects
- May modify internal state of the `ChecksGroup` via `onFailure`.
- Logs debug information if a panic occurs.
- No external side effects beyond the group’s own data.

### Package Context
Within **checksdb**, each test suite is organized into *check groups*.  
These groups can declare:
- **BeforeAll** – run once before any check,
- **AfterAll** – run once after all checks,
- **BeforeEach / AfterEach** – per‑check hooks.

`runAfterAllFn` is the implementation that ties the `afterAllFn` to the group’s lifecycle. It is called by higher‑level orchestration code (e.g., when a suite finishes) to ensure proper cleanup and error reporting for each group.

---
