MultiHandler.Handle`

| | |
|---|---|
| **Receiver** | `*MultiHandler` – a composite slog handler that forwards records to several underlying handlers. |
| **Signature** | `func (mh *MultiHandler) Handle(ctx context.Context, r slog.Record) error` |
| **Exported?** | Yes |

### Purpose
`Handle` is the core method that satisfies the `slog.Handler` interface for a `MultiHandler`.  
When a log record arrives, it must be sent to every child handler registered in the composite. This method orchestrates that forwarding while preserving context and record integrity.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Execution context passed through the logging pipeline; may contain cancellation or deadline signals, but is otherwise unused by this implementation. |
| `r` | `slog.Record` | The log record to be processed – contains level, time, source, message, and any attributes. |

### Output
| Return | Type | Meaning |
|--------|------|---------|
| `error` | `nil` or an error value | If any child handler fails while handling the record, that error is returned immediately; otherwise `nil`. The first error encountered stops further propagation. |

### Key Dependencies & Calls
- **`mh.handlers`** – a slice of `slog.Handler`s stored in the receiver.  
  Each element is invoked in order.
- For each handler:
  1. `Clone()` – creates an independent copy of the record so that subsequent handlers can modify or enrich it without affecting others.
  2. The cloned record’s `Handler` field is set to the current child handler (`h`).
  3. `h.Handle(ctx, clone)` – forwards the record to the underlying handler.

The method itself does not use any global variables; all state comes from the receiver and the incoming parameters.

### Side Effects
- **No mutation of the original record**: cloning guarantees that each child sees a pristine copy.
- **Short‑circuit on error**: as soon as one handler returns an error, propagation stops. This is intentional to surface failures promptly but means later handlers will not see the record if an earlier one fails.

### Relationship to Package
`MultiHandler` lives in `internal/log`. It allows callers to compose several logging backends (e.g., console, file, external service) into a single logical handler that can be passed to a top‑level logger. The `Handle` method is the glue that makes this composition work: it ensures all constituent handlers receive every log record while respecting Go’s `slog.Handler` contract.

> **Note**: The function body itself is intentionally simple; all heavy lifting (e.g., formatting, file I/O) happens inside each child handler. This design keeps the composite lightweight and easy to reason about.
