package log

import (
	"context"
	"log/slog"
)

// MultiHandler aggregates multiple slog.Handler instances.
//
// It implements the slog.Handler interface by forwarding log records to each
// contained handler and combining their enabled checks. The struct holds a
// slice of handlers that are invoked in order when Enabled, Handle, WithAttrs,
// or WithGroup is called. NewMultiHandler returns a pointer to this struct
// initialized with the provided handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a handler that forwards log records to multiple underlying handlers.
//
// It accepts any number of slog.Handler arguments and returns a MultiHandler
// that dispatches each record to all provided handlers in order. The returned
// handler can be used with slog.New to combine logging outputs, such as console,
// file, or custom destinations. If no handlers are supplied it returns nil.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Enabled reports whether the MultiHandler will log at the given level.
//
// It accepts a context and a slog.Level and returns true if any of the
// underlying handlers are enabled for that level, otherwise false.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle processes a slog record by dispatching it to all underlying handlers in the MultiHandler.
//
// It receives a context and a Record, clones the record for each handler,
// and calls each embedded handler's Handle method. Any error returned
// from an inner handler is propagated back to the caller.
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for i := range h.handlers {
		if err := h.handlers[i].Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}

	return nil
}

// WithAttrs returns a new slog.Handler that adds the provided attributes to all log records produced by the MultiHandler.
//
// It accepts a slice of slog.Attr which are appended to each record before it is dispatched
// to the underlying handlers. The function creates a new handler for each child handler,
// applies WithAttrs on them, and combines them into a new MultiHandler using NewMultiHandler.
// If no attributes are supplied or the receiver has no children, it returns nil.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlersWithAttrs := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithAttrs[i] = h.handlers[i].WithAttrs(attrs)
	}
	return NewMultiHandler(handlersWithAttrs...)
}

// WithGroup returns a handler that writes log records with the specified group prefix.
//
// It creates a new MultiHandler containing the existing handlers of the receiver,
// each wrapped to prepend the given group name to their log records.
// If the receiver has no handlers, it returns nil.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlersWithGroup := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithGroup[i] = h.handlers[i].WithGroup(name)
	}
	return NewMultiHandler(handlersWithGroup...)
}
