skipAll`

| Item | Details |
|------|---------|
| **Signature** | `func([]*Check, string)()` |
| **Exported?** | No – internal helper used only inside the package |
| **Package** | `checksdb` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb`) |

### Purpose
`skipAll` creates a *deferred* action that marks every check in a given slice as skipped.  
It is typically used when a test‑group or a full run cannot be executed because of a higher‑level failure (e.g., missing prerequisites).  By returning a closure, the caller can defer the call until just before the function exits, ensuring all checks are marked consistently regardless of how the exit path occurs.

### Parameters
| Name | Type | Meaning |
|------|------|---------|
| `checks` | `[]*Check` | The list of check objects that should be skipped. Each element represents a single test case in the database. |
| `reason` | `string` | Human‑readable explanation for why the checks are being skipped (e.g., `"Missing required dependency"`). |

### Return Value
A zero‑argument function `func()` that, when invoked, will iterate over `checks` and call `skipCheck(c, reason)` on each.  
The returned closure has **no return value**; its sole side effect is to modify the state of each `Check`.

### Key Dependency
- `skipCheck(*Check, string)`: The helper function that actually sets a check’s result to `SKIPPED` and records the provided reason. It lives in the same package and does not have external side effects beyond mutating the passed `Check`.

### Side Effects
1. **State mutation** – each `*Check`’s internal fields (e.g., status, error message) are altered to reflect a skipped state.
2. No global variables are touched directly; all changes happen on the supplied slice.

### How It Fits the Package
The `checksdb` package maintains an in‑memory database of checks and groups (`dbByGroup`, `resultsDB`).  Test runners call `skipAll` when they need to abort a group or entire run early.  
Because test functions often use `defer` for cleanup, returning a closure allows callers to schedule the skip logic *after* all setup but *before* any results are committed:

```go
func runGroup(groupName string) {
    checks := dbByGroup[groupName].Checks

    // Abort early if prerequisites fail
    if !prereqsMet() {
        defer skipAll(checks, "Prerequisites not met")()
    }

    // ... normal execution ...
}
```

This pattern keeps the code tidy and guarantees that every check in the group ends up with a deterministic status.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[runGroup] --> B{prereqsMet?}
  B -- no --> C[defer skipAll(checks, reason)()]
  C --> D[exit]
  B -- yes --> E[execute checks]
```

The diagram shows how `skipAll` is invoked as a deferred cleanup when prerequisites fail.

---

**Summary**  
`skipAll` is an internal helper that produces a closure to mark all supplied checks as skipped with a given reason. It plays a key role in graceful error handling and consistent result reporting within the `checksdb` package.
