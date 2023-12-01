package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"
)

type Logger struct {
	*slog.Logger
}

var logger *slog.Logger

func SetupLogger(logWriter io.Writer) {
	opts := Options{
		Level: slog.LevelDebug,
	}
	logger = slog.New(NewCustomHandler(logWriter, &opts))
}

func Debug(msg string, args ...any) {
	Logf(logger, slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	Logf(logger, slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	Logf(logger, slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	Logf(logger, slog.LevelError, msg, args...)
}

func GetMultiLogger(w io.Writer) *slog.Logger {
	opts := Options{
		Level: slog.LevelDebug,
	}
	return slog.New(NewMultiHandler(logger.Handler(), NewCustomHandler(w, &opts)))
}

// The Logf function should be called inside a log wrapper function.
// Otherwise the code source reference will be invalid.
func Logf(logger *slog.Logger, level slog.Level, format string, args ...any) {
	if !logger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	// skip [Callers, Log, LogWrapper]
	runtime.Callers(3, pcs[:]) //nolint:gomnd
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}
