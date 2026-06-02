package log

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

func TestNewCustomHandlerDefaultLevel(t *testing.T) {
	t.Parallel()
	h := NewCustomHandler(&bytes.Buffer{}, nil)

	require.NotNil(t, h)
	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestNewCustomHandlerWithOpts(t *testing.T) {
	t.Parallel()
	h := NewCustomHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	require.NotNil(t, h)
	assert.True(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestEnabled(t *testing.T) {
	t.Parallel()
	h := NewCustomHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})

	tests := []struct {
		name    string
		level   slog.Level
		enabled bool
	}{
		{"debug below threshold", slog.LevelDebug, false},
		{"info below threshold", slog.LevelInfo, false},
		{"warn at threshold", slog.LevelWarn, true},
		{"error above threshold", slog.LevelError, true},
		{"fatal above threshold", CustomLevelFatal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.enabled, h.Enabled(context.Background(), tt.level))
		})
	}
}

func TestHandleWritesFormattedOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := NewCustomHandler(&buf, nil)

	r := slog.NewRecord(testTime, slog.LevelInfo, "hello world", 0)
	err := h.Handle(context.Background(), r)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello world")
	assert.Contains(t, output, "\n")
}

func TestHandleWritesErrorLevel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := NewCustomHandler(&buf, nil)

	r := slog.NewRecord(testTime, slog.LevelError, "something failed", 0)
	err := h.Handle(context.Background(), r)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "something failed")
}

func TestWithAttrsReturnsNewHandler(t *testing.T) {
	t.Parallel()
	original := NewCustomHandler(&bytes.Buffer{}, nil)
	modified := original.WithAttrs([]slog.Attr{
		slog.String("key", "value"),
	})

	require.NotNil(t, modified)
	assert.NotEqual(t, original, modified)

	modifiedHandler, ok := modified.(*CustomHandler)
	require.True(t, ok)
	assert.NotEmpty(t, modifiedHandler.attrs)
	assert.Empty(t, original.attrs)
}

func TestWithAttrsEmptyReturnsOriginal(t *testing.T) {
	t.Parallel()
	original := NewCustomHandler(&bytes.Buffer{}, nil)
	same := original.WithAttrs(nil)

	assert.Equal(t, original, same)
}

func TestWithGroupReturnsNil(t *testing.T) {
	t.Parallel()
	h := NewCustomHandler(&bytes.Buffer{}, nil)

	assert.Nil(t, h.WithGroup("test"))
}

func TestHandleWithAttrs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := NewCustomHandler(&buf, nil)
	h2 := h.WithAttrs([]slog.Attr{
		slog.String("request_id", "abc-123"),
	})

	handler, ok := h2.(*CustomHandler)
	require.True(t, ok)

	r := slog.NewRecord(testTime, slog.LevelInfo, "with attrs", 0)
	err := handler.Handle(context.Background(), r)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "abc-123")
	assert.Contains(t, buf.String(), "with attrs")
}
