package log

import (
	"context"
	"log/slog"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

//nolint:gocritic
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for i := range h.handlers {
		if err := h.handlers[i].Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}

	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlersWithAttrs := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithAttrs[i] = h.handlers[i].WithAttrs(attrs)
	}
	return NewMultiHandler(handlersWithAttrs...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlersWithGroup := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlersWithGroup[i] = h.handlers[i].WithGroup(name)
	}
	return NewMultiHandler(handlersWithGroup...)
}
