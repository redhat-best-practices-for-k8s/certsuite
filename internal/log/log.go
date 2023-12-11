package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

const (
	LogFileName        = "cnf-certsuite.log"
	LogFilePermissions = 0o644
)

var logger *slog.Logger

func SetupLogger(logWriter io.Writer, level slog.Level) {
	opts := Options{
		Level: level,
	}
	logger = slog.New(NewCustomHandler(logWriter, &opts))
}

func SetLogger(l *slog.Logger) {
	logger = l
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

func ParseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}

	return 0, fmt.Errorf("not a valid slog Level: %q", level)
}

// The Logf function should be called inside a log wrapper function.
// Otherwise the code source reference will be invalid.
func Logf(logger *slog.Logger, level slog.Level, format string, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}
	if !logger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	// skip [Callers, Log, LogWrapper]
	runtime.Callers(3, pcs[:]) //nolint:gomnd
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.Handler().Handle(context.Background(), r)
}
