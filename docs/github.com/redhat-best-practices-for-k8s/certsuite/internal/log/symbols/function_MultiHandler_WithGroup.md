MultiHandler.WithGroup`

| Item | Description |
|------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/internal/log` |
| **Receiver type** | `MultiHandler` (a composite slog.Handler that forwards records to several underlying handlers) |
| **Signature** | `func (mh MultiHandler) WithGroup(name string) slog.Handler` |
| **Exported?** | Yes – it implements the standard `slog.Handler` interface. |

### Purpose
`WithGroup` creates a new handler that prefixes every log record with a *group name*.  
For a `MultiHandler`, this means that each underlying handler receives its own grouped copy of the record.

The function is part of the `slog` API: any handler may implement `WithGroup` to support structured logging where fields are logically nested under named groups.

### Inputs / Outputs
| Parameter | Type | Notes |
|-----------|------|-------|
| `name` | `string` | The group name to prepend. |

| Return value | Type | Notes |
|--------------|------|-------|
| `slog.Handler` | A new handler that wraps the original `MultiHandler`. The returned handler forwards grouped records to all sub‑handlers of the receiver. |

### Implementation details

```go
func (mh MultiHandler) WithGroup(name string) slog.Handler {
    if len(mh.handlers) == 0 { // mh is a slice of underlying handlers
        return mh
    }
    // create a new slice with the same capacity as the original
    grouped := make([]slog.Handler, len(mh.handlers))
    for i, h := range mh.handlers {
        grouped[i] = h.WithGroup(name) // delegate to each handler
    }
    // build a new MultiHandler that holds the grouped handlers
    return NewMultiHandler(grouped...)
}
```

1. **Early exit** – If the receiver contains no sub‑handlers (`len(mh.handlers) == 0`), it simply returns itself.  
   This avoids allocating an empty slice and is safe because grouping a nil handler has no effect.

2. **Delegation** – For every underlying handler `h`, it calls `h.WithGroup(name)` to obtain a grouped variant of that handler.  
   The standard `slog.Handler` implementations (e.g., `TextHandler`, `JSONHandler`) return a new instance that prefixes the group name when formatting attributes.

3. **Reconstruction** – The grouped handlers are collected into a new slice and wrapped in a fresh `MultiHandler` via `NewMultiHandler`.  
   This preserves the original ordering of handlers while ensuring each receives records with the group applied.

### Dependencies
- **Standard library**
  - `slog.Handler.WithGroup`
- **Local package**
  - `NewMultiHandler` – constructor for creating a composite handler from a variadic list of handlers.
- No global variables are read or modified; the function is purely functional.

### Side effects & invariants
- **No mutation** – The original `MultiHandler` remains unchanged.  
- **Thread‑safety** – All operations are on local copies; the function can be called concurrently with other handler methods.

### How it fits the package

The `log` package provides a thin wrapper around Go’s `slog` infrastructure, adding conveniences such as global loggers and custom levels.  
`MultiHandler.WithGroup` enables structured logging across multiple outputs (e.g., console + file) while keeping the group hierarchy consistent in each destination.

**Typical usage**

```go
// Assume we have a global logger that uses a MultiHandler.
log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
groupedLog := log.WithGroup("request")
groupedLog.InfoContext(ctx, "Processing request", slog.String("id", reqID))
```

Each underlying handler in the `MultiHandler` receives the `"request"` group as part of the record’s attributes.

### Suggested Mermaid diagram

```mermaid
graph LR
  A[Original MultiHandler] --> B{Has Handlers?}
  B -- no --> C[Return self]
  B -- yes --> D[Create grouped slice]
  D --> E[Each handler.WithGroup(name)]
  E --> F[NewMultiHandler(grouped...)]
```

This diagram visualises the decision path and construction of the new grouped handler.
