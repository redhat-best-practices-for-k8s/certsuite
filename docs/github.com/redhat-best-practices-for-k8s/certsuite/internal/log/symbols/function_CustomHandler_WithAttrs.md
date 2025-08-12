## `WithAttrs` – Augment a `CustomHandler` with extra attributes

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/internal/log` |
| **Signature** | `func (h CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler` |
| **Exported?** | Yes – public API of the custom logging handler. |

---

### Purpose

`WithAttrs` implements the `slog.Handler` interface’s `WithAttrs` method, allowing callers to create a *derived* handler that will prepend a fixed set of attributes (`attrs`) to every log record it processes.  
This is used when you want to attach contextual information (e.g., request ID, component name) to all subsequent log entries without passing the attributes explicitly each time.

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `attrs` | `[]slog.Attr` | A slice of slog attributes that should be added in front of every log record handled by the returned handler. |

### Return value

| Value | Type | Meaning |
|-------|------|---------|
| `slog.Handler` | `CustomHandler` (wrapped) | A new `CustomHandler` instance containing a copy of the original handler’s state plus the supplied attributes. |

> **Note:** The function returns the *same* concrete type (`CustomHandler`) that implements `slog.Handler`. The returned value satisfies the interface, so it can be used wherever an `slog.Handler` is expected.

### Implementation details

1. **Copy existing attrs**  
   ```go
   // Start with a new slice that has room for both old and new attributes.
   newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
   ```
2. **Append existing handler attributes**  
   Existing attributes (`h.attrs`) are copied to `newAttrs` using `copy`. This preserves any context previously attached by earlier calls to `WithAttrs`.
3. **Add new attributes**  
   The supplied `attrs` slice is appended, ensuring that the caller‑supplied attributes appear *after* any already present ones.
4. **Return a new handler**  
   ```go
   return CustomHandler{attrs: newAttrs}
   ```
   No other fields of `h` are used; `CustomHandler` contains only the `attrs` slice, so creating a copy is inexpensive.

### Side effects

* None beyond allocating a new slice.  
* The original handler (`h`) remains unchanged because a fresh slice is created for the returned instance.

### Interaction with the rest of the package

- **Global logger** – The custom handler is typically wrapped by `globalLogger` (see `log.go`). When that global logger calls `WithAttrs`, it obtains a new handler that carries extra attributes, which are then forwarded to the underlying log file or output stream.
- **Logging flow** – During a log call (`Info`, `Error`, etc.), the handler’s `Handle` method will receive the combined attribute list (original + new) and format them into the final log entry.

---

### Usage example

```go
// Attach request ID to all logs for this request context.
h := globalLogger.Handler()
ctxHandler := h.WithAttrs([]slog.Attr{
    slog.String("request_id", "abc123"),
})

// Use ctxHandler in a new logger:
ctxLogger := slog.New(ctxHandler)
ctxLogger.Info("Processing request")
```

The resulting log line will include the `"request_id":"abc123"` attribute alongside any other fields emitted by `Info`.
