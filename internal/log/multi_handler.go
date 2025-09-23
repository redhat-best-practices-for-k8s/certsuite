package log

import (
	"context"
	"log/slog"
)

// MultiHandler combines multiple logging handlers into one
//
// It holds a slice of slog.Handler values and forwards each logging call to
// every handler in the slice. For enabled checks, it returns true if any
// underlying handler is enabled for the given level. When handling a record, it
// clones the record before passing it to each handler, stopping early only if
// an error occurs. Attribute and group additions are propagated by creating new
// handlers with the specified attributes or groups.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler Creates a composite handler for multiple slog handlers
//
// This function takes any number of slog.Handler instances and returns a new
// MultiHandler that aggregates them. The returned object holds the provided
// handlers in order, enabling log records to be dispatched to each underlying
// handler when emitted. No additional processing or filtering is performed; it
// simply stores the handlers for later use.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// MultiHandler.Enabled True when any contained handler accepts the log level
//
// The method iterates over all handlers stored in the MultiHandler and queries
// each one to see if it would handle messages at the specified level. If any
// handler reports enabled, the function immediately returns true; otherwise it
// returns false after checking all handlers.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// MultiHandler.Handle distributes a log record to all registered handlers
//
// The method iterates over each handler stored in the MultiHandler, cloning the
// incoming record before passing it to ensure isolation between handlers. If
// any handler returns an error, that error is immediately returned and no
// further handlers are invoked. When all handlers succeed, the method completes
// without error.
//
//nolint:gocritic
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for i := range h.handlers {
		if err := h.handlers[i].Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}

	return nil
}

// MultiHandler.WithAttrs creates a new handler that adds attributes to all sub-handlers
//
// This method iterates over each contained handler, invoking its
// attribute-adding function with the supplied slice of attributes. It collects
// the resulting handlers into a new slice and constructs a fresh multi-handler
// from them. The returned handler behaves like the original but ensures every
// log record includes the provided attributes.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlersWithAttrs := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithAttrs[i] = h.handlers[i].WithAttrs(attrs)
	}
	return NewMultiHandler(handlersWithAttrs...)
}

// MultiHandler.WithGroup Adds a named group to all underlying handlers
//
// This method creates a new slice of slog.Handler by iterating over the
// existing handlers and invoking each one's WithGroup method with the provided
// name. The resulting handlers are then wrapped into a new MultiHandler
// instance, which is returned as a slog.Handler. This allows grouping log
// entries consistently across multiple output destinations.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlersWithGroup := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithGroup[i] = h.handlers[i].WithGroup(name)
	}
	return NewMultiHandler(handlersWithGroup...)
}
