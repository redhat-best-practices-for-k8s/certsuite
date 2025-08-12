CustomHandler` – A Structured Log Handler

The **`CustomHandler`** type implements the standard library’s
[`slog.Handler`](https://pkg.go.dev/log/slog?tab=doc#Handler) interface.
It formats log records into a compact, human‑readable line that contains:

```
LOG_LEVEL [TIME] [SOURCE_FILE] [CUSTOM_ATTRS] MSG
```

Typical usage is to replace the default slog handler in an application
so that logs are written to a file or stdout with this custom format.

> **Why not use the built‑in handlers?**  
> The standard `TextHandler` prints JSON‑style key/value pairs; the
> `Formatter` interface (added in Go 1.22) still requires a separate
> package for custom formatting.  `CustomHandler` keeps everything in one
> place and is intentionally lightweight.

---

## Fields

| Field | Type | Purpose |
|-------|------|---------|
| `attrs []slog.Attr` | slice of attributes that are added to every log record (via `WithAttrs`) |
| `mu *sync.Mutex` | protects concurrent writes to `out` and the internal state |
| `opts slog.HandlerOptions` | configuration options such as level, time format, etc. |
| `out io.Writer` | destination for the formatted log line (`os.Stdout`, a file, …) |

---

## Constructor

```go
func NewCustomHandler(out io.Writer, opts *slog.HandlerOptions) *CustomHandler
```

* **Parameters**
  * `out`: where to write logs.
  * `opts`: optional handler options; if `nil` defaults are used.

* **Returns** a fully‑initialized `CustomHandler`.

---

## Methods

### `Enabled(ctx context.Context, level slog.Level) bool`

Checks whether the given log level is enabled according to
`h.opts.Level`.  
It simply delegates to `opts.Level`, respecting any custom level logic.

> **Side effect:** none.  
> **Return value:** `true` if logging should proceed.

---

### `Handle(ctx context.Context, r slog.Record) error`

The core formatting routine:

1. Builds a byte slice starting with the log level string.
2. Appends the timestamp (`r.Time`) formatted via `opts.TimeFormat`.
3. Resolves the call stack to find the source file of the caller
   and appends it (basename only).
4. Adds any custom attributes that were set through `WithAttrs`
   by calling `appendAttr` for each.
5. Appends the log message (`r.Message`) and any arguments.
6. Writes the final line to `out`, protected by `mu`.

> **Side effect:** writes a single line to `h.out`.  
> **Return value:** always `nil`; errors are ignored because the logger
> cannot do much with them.

---

### `WithAttrs(attrs []slog.Attr) slog.Handler`

Creates a new handler that will prepend `attrs` to every subsequent log
record.  The implementation copies the slice and returns a shallow copy
of the handler, preserving thread safety.

> **Side effect:** none.  
> **Return value:** a new `CustomHandler` with merged attributes.

---

### `WithGroup(name string) slog.Handler`

Not implemented – returns `nil`.  This satisfies the interface but
prevents grouping of attributes.  If grouping is needed in the future,
this method should be updated to return a handler that tracks groups.

> **Side effect:** none.  
> **Return value:** always `nil`.

---

### `appendAttr(b []byte, a slog.Attr) []byte` (unexported)

Helper that serializes an attribute into the log line format:

```
key=value
```

The formatting depends on the attribute’s kind:
* Numbers → `%v`
* Booleans → `%t`
* Times → formatted with `opts.TimeFormat`
* Others → string representation

It appends a space after each attribute.

> **Side effect:** none.  
> **Return value:** the expanded byte slice.

---

## How it fits into the package

`log` is an internal package providing a thin wrapper around Go’s
`slog`.  `CustomHandler` replaces the default handler when initializing
the logger:

```go
h := NewCustomHandler(os.Stdout, nil)
logger := slog.New(h)
```

The package also exposes convenience wrappers for creating loggers with
different output destinations or options.  By centralising formatting in
`CustomHandler`, all logs across the application share a consistent
layout without requiring external dependencies.

---

## Example

```go
// main.go
package main

import (
    "os"
    "log/slog"

    "github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

func main() {
    h := log.NewCustomHandler(os.Stdout, nil)
    logger := slog.New(h)

    // Add a global attribute
    logger = logger.WithAttrs([]slog.Attr{
        slog.String("app", "certsuite"),
        slog.Int("pid", os.Getpid()),
    })

    logger.Info("Starting up")
    logger.Warn("Configuration missing", slog.String("key", "timeout"))
}
```

Output (example):

```
INFO 2025-08-11T12:34:56Z main.go[42] app=certsuite pid=12345 Starting up
WARN 2025-08-11T12:34:56Z main.go[43] app=certsuite pid=12345 key=timeout Configuration missing
```

---
