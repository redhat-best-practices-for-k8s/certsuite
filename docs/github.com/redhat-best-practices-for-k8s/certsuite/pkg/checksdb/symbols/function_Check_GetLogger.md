Check.GetLogger`

```go
func (c Check) GetLogger() *log.Logger
```

### Purpose
`GetLogger` returns the logger instance that is bound to a particular check.
The logger is used throughout the execution of a check (e.g., in the
`Run`, `Validate`, or any helper functions) to emit structured log
messages.  It also allows external callers (tests, reporters, or the
framework itself) to retrieve and inspect the same logger.

### Receiver
- **`c Check`** – A value copy of a check.  
  The method does not modify the receiver; it simply accesses the embedded
  `*log.Logger`.

### Return Value
- **`*log.Logger`** – The logger that was assigned to this check when it was
  created (via `NewCheck`, or by the checks database during registration).

### Dependencies & Side‑Effects
| Dependency | Reason |
|------------|--------|
| `log` package | Provides the `Logger` type. |
| None other | The method only reads a field; no global state is modified. |

No side effects occur: it simply returns a pointer to an existing logger.

### How It Fits Into the Package

The `checksdb` package maintains a registry of checks (`Check`) that are
executed against Kubernetes objects or cluster state.  Each check holds a
logger so that:

1. **Consistent Logging** – All log messages from a single check share the
   same output destination and formatting.
2. **Testability** – Tests can replace the logger with a custom implementation
   (e.g., an in‑memory buffer) by assigning to `Check.logger` before calling
   `GetLogger`.
3. **Extensibility** – Reporters or higher‑level orchestrators can capture
   logs per check without having to know how the logger was created.

The function is typically called internally when a check needs to log something:

```go
func (c Check) Run(ctx context.Context, data interface{}) error {
    c.GetLogger().Info("Running check")
    ...
}
```

Because `GetLogger` simply returns the pointer stored in the struct,
it can be used safely from concurrent goroutines that execute a check.

### Usage Example

```go
// Create a new check with its own logger.
chk := checksdb.NewCheck(
    "mycheck",
    ... // other params
)

// Retrieve the logger to add custom handlers or inspect logs.
logger := chk.GetLogger()
logger.SetPrefix("[MyCheck] ")
```

In this example, `GetLogger` gives direct access to the underlying logger,
allowing callers to customize logging behaviour for that particular check.

---

**Summary**

- **What it does:** returns the logger attached to a `Check`.  
- **Inputs/Outputs:** no parameters; outputs `*log.Logger`.  
- **Side effects:** none.  
- **Role in package:** provides a stable, per‑check logging handle used by
  check logic and external consumers.
