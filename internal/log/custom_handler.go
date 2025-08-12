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

// CustomHandler implements slog.Handler with custom formatting.
//
// It writes log records to an io.Writer in the format:
// LOG_LEVEL [TIME] [SOURCE_FILE] [CUSTOM_ATTRS] MSG
//
// The handler supports adding attributes via WithAttrs, but group handling is not implemented.
// A mutex protects concurrent writes. Options are used to control time and level formatting.
type CustomHandler struct {
	opts  slog.HandlerOptions
	attrs []slog.Attr
	mu    *sync.Mutex
	out   io.Writer
}

// NewCustomHandler creates a custom slog handler that writes to the provided writer.
//
// It accepts an io.Writer where log output will be sent and optional HandlerOptions
// that control formatting, level filtering, and other behavior. The function returns
// a pointer to a CustomHandler instance configured with the supplied writer and options.
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

// Enabled determines if a log message at the given level should be emitted by this handler.
//
// Enabled reports whether logging is enabled for the supplied level.
//
// It accepts a context and a slog.Level, retrieves the current global
// log level via Level(), and returns true if the provided level is
// greater than or equal to that global threshold. This method allows
// callers to avoid expensive message construction when the message
// would be filtered out by the logger configuration.
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle writes a formatted log line to the configured output.
//
// It formats the record as:
// LOG_LEVEL [TIME] [SOURCE_FILE] [CUSTOM_ATTRS] MSG
// and writes it to the global log file.
// The function returns any error that occurs during writing.
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

// WithGroup returns a slog.Handler for the specified group.
//
// It takes a single string argument representing the name of the group and
// returns an slog.Handler that would handle log records belonging to that
// group. Currently this method is not implemented and always returns nil,
// indicating no handler is available for the requested group.
func (h *CustomHandler) WithGroup(_ string) slog.Handler {
	return nil
}

// WithAttrs returns a new handler that includes the provided attributes in each log record.
//
// It accepts a slice of slog.Attr and creates a copy of those attributes,
// then returns a slog.Handler configured to attach them to all subsequent logs.
// The returned handler preserves the original handler’s behavior while adding
// the supplied attributes.
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

// appendAttr appends a single slog.Attr to an existing byte slice, formatting the attribute according to its kind and returning the resulting slice.
//
// It resolves the attribute value, then writes the key followed by the formatted value.
// The format depends on the value type: strings are quoted, times are printed in RFC3339,
// numbers and booleans are written directly. If the attribute has a custom string
// representation via Format or Time methods, those are used. The function returns
// the byte slice with the new content appended.
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
