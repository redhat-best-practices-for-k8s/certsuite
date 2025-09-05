package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Custom log levels
const CustomLevelFatal = slog.Level(12)

var CustomLevelNames = map[slog.Leveler]string{
	CustomLevelFatal: "FATAL",
}

// CustomHandler Formats and writes structured log entries to an output stream
//
// The handler collects attributes and optional context information such as
// level, time, source file, and message. It serializes these into a single line
// using a custom attribute formatting routine before writing them atomically to
// the configured writer. The handler supports adding default attributes via
// WithAttrs while preserving thread safety with a mutex.
type CustomHandler struct {
	opts  slog.HandlerOptions
	attrs []slog.Attr
	mu    *sync.Mutex
	out   io.Writer
}

// NewCustomHandler Creates a thread‑safe log handler that writes to an io.Writer
//
// This function constructs a CustomHandler with the supplied writer and
// optional slog.HandlerOptions. If options are nil or lack a level, it defaults
// to slog.LevelInfo. The resulting handler can be used by other components to
// emit structured logs in a concurrency‑safe manner.
func NewCustomHandler(out io.Writer, opts *slog.HandlerOptions) *CustomHandler {
	h := &CustomHandler{out: out, mu: &sync.Mutex{}}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}

	return h
}

// CustomHandler.Enabled Determines if a log level is enabled based on configuration
//
// The method compares the supplied logging level against the handler's
// configured threshold, returning true when the level meets or exceeds that
// threshold. It ignores the context parameter because the decision relies
// solely on static settings.
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// CustomHandler.Handle writes a formatted log line to the output
//
// This method receives a context and a slog.Record, builds a byte buffer
// containing level, time, source file, custom attributes, and message in a
// specific format, then writes it to an underlying writer. It locks a mutex
// during the write to ensure thread safety and returns any error from the write
// operation.
//
//nolint:gocritic // r param is heavy but defined in the slog.Handler interface
func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	var buf []byte
	// Level
	var levelAttr slog.Attr
	if h.opts.ReplaceAttr != nil {
		levelAttr = h.opts.ReplaceAttr(nil, slog.Any(slog.LevelKey, r.Level))
	} else {
		levelAttr = slog.Any(slog.LevelKey, r.Level)
	}
	buf = h.appendAttr(buf, levelAttr)
	// Time
	if !r.Time.IsZero() {
		buf = h.appendAttr(buf, slog.Time(slog.TimeKey, r.Time))
	}
	// Source
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf = h.appendAttr(buf, slog.String(slog.SourceKey, fmt.Sprintf("%s: %d", filepath.Base(f.File), f.Line)))
	}
	// Attributes
	for _, attr := range h.attrs {
		buf = h.appendAttr(buf, attr)
	}
	// Message
	buf = h.appendAttr(buf, slog.String(slog.MessageKey, r.Message))
	buf = append(buf, "\n"...)
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

// CustomHandler.WithGroup Returns a new handler scoped to a named group
//
// When called, this method ignores the provided group name and simply returns a
// nil handler, indicating that grouping functionality is not implemented for
// CustomHandler. It satisfies the slog.Handler interface but does not create
// any new handler instance or modify state.
func (h *CustomHandler) WithGroup(_ string) slog.Handler {
	return nil
}

// CustomHandler.WithAttrs Creates a handler that includes additional attributes
//
// The method takes a slice of attributes, merges them with the handler’s
// existing ones, and returns a new handler instance containing the combined
// set. If no attributes are supplied it simply returns the original handler to
// avoid unnecessary copying. The returned handler is a copy of the receiver so
// that modifications do not affect the original.
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	// Create a new handler with default attributes
	h2 := *h
	// A deep copy of the attributes is required
	h2.attrs = make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(h2.attrs, h.attrs)
	h2.attrs = append(h2.attrs, attrs...)

	return &h2
}

// CustomHandler.appendAttr Formats a logging attribute for output
//
// The function resolves the attribute’s value, skips empty attributes, then
// formats the output based on the kind of value. String values are printed
// plainly or in brackets; time values use a millisecond timestamp; other kinds
// include level or key/value pairs with appropriate spacing. The resulting
// bytes are appended to the buffer and returned.
func (h *CustomHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	// Resolve the Attr's value before doing anything else.
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return buf
	}
	switch a.Value.Kind() {
	case slog.KindString:
		if a.Key == slog.MessageKey {
			buf = fmt.Appendf(buf, "%s", a.Value.String())
		} else {
			buf = fmt.Appendf(buf, "[%s] ", a.Value.String())
		}
	case slog.KindTime:
		buf = fmt.Appendf(buf, "[%s] ", a.Value.Time().Format(time.StampMilli))
	default:
		if a.Key == slog.LevelKey {
			buf = fmt.Appendf(buf, "%-5s ", a.Value.String())
		} else {
			buf = fmt.Appendf(buf, "[%s: %s] ", a.Key, a.Value)
		}
	}
	return buf
}
