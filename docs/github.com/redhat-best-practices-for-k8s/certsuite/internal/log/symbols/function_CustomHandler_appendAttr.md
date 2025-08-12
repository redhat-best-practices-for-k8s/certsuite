CustomHandler.appendAttr`

### Purpose
`appendAttr` is an internal helper of the `CustomHandler` type that serialises a single log attribute into a buffer.  
The handler builds log lines by repeatedly calling this function for each key/value pair in a record.

### Signature
```go
func (h CustomHandler) appendAttr(buf []byte, attr slog.Attr) []byte
```

| Parameter | Type          | Description                                 |
|-----------|---------------|---------------------------------------------|
| `buf`     | `[]byte`      | The output buffer that will contain the log line. The function appends to it and returns the updated slice. |
| `attr`    | `slog.Attr`   | The attribute (key/value pair) to serialise. |

The method is **unexported**; callers must use the public `CustomHandler.Handle` implementation which internally uses `appendAttr`.

### How it works
1. **Resolve the value**  
   `attr.Value.Resolve()` normalises the value (e.g., evaluates deferred values).  
2. **Key handling**  
   The attribute key is written first, followed by a space.  
3. **Value formatting** – based on the resolved kind:  
   - **String / Any** – use `fmt.Sprintf("%q", v.String())` to quote strings or other printable values.  
   - **Time** – format with RFC3339 via `t.Format(time.RFC3339)`.  
   - **Duration** – format as string (`v.Duration().String()`).  
   - **Int, Float, Bool, etc.** – use the value’s `String()` representation.  
4. Each key/value pair is separated by a space; the function returns the extended buffer ready for the next attribute.

### Dependencies
| Dependency | Role |
|------------|------|
| `slog.Attr` | Input structure representing a log key/value. |
| `Resolve`, `Kind`, `String` (methods on `slog.Value`) | Resolve value and determine its type. |
| `Appendf` from the `bytes` package | Efficient string concatenation to avoid repeated allocations. |
| Standard formatting functions (`fmt.Sprintf`, `time.RFC3339`) | Produce human‑readable representations for special kinds. |

### Side effects
* The function only mutates and returns the supplied buffer; it does **not** modify any global state or file handles.
* It relies on the attribute’s value being valid; panics may occur if `Resolve()` fails (unlikely in normal use).

### Package context
`CustomHandler` is part of the internal logging package.  
It implements a custom `slog.Handler` that writes log records to a pre‑opened file (`globalLogFile`).  
During handling, the handler iterates over all attributes of a record and calls `appendAttr` for each, building a single line before writing it to disk.

```mermaid
flowchart LR
    Record -->|for each attr| appendAttr
    appendAttr -->|returns buf| Handler.writeLine
```

### Summary
`CustomHandler.appendAttr` is the low‑level routine that turns an `slog.Attr` into a textual representation appended to a byte buffer.  
It is central to how the custom handler formats log messages before persisting them.
