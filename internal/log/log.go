package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LogFileName        = "cnf-certsuite.log"
	LogFilePermissions = 0o644
)

// Log levels
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
)

type Logger struct {
	l *slog.Logger
}

var (
	globalLogger   *Logger
	globalLogLevel slog.Level
)

func SetupLogger(logWriter io.Writer, level string) {
	logLevel, err := parseLevel(level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse log level, err: %v. Defaulting to DEBUG.", err)
		globalLogLevel = slog.LevelDebug
	} else {
		globalLogLevel = logLevel
	}

	opts := slog.HandlerOptions{
		Level: globalLogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := CustomLevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}

	globalLogger = &Logger{
		l: slog.New(NewCustomHandler(logWriter, &opts)),
	}
}

func SetLogger(l *Logger) {
	globalLogger = l
}

func GetLogger() *Logger {
	return globalLogger
}

func GetMultiLogger(writers ...io.Writer) *Logger {
	opts := slog.HandlerOptions{
		Level: globalLogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := CustomLevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}

	handlers := []slog.Handler{globalLogger.l.Handler()}
	for _, writer := range writers {
		handlers = append(handlers, NewCustomHandler(writer, &opts))
	}

	return &Logger{l: slog.New(NewMultiHandler(handlers...))}
}

// Top-level log functions
func Debug(msg string, args ...any) {
	Logf(globalLogger, LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	Logf(globalLogger, LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	Logf(globalLogger, LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	Logf(globalLogger, LevelError, msg, args...)
}

func Fatal(msg string, args ...any) {
	Logf(globalLogger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...) //nolint: goconst
	os.Exit(1)
}

// Log methods for a logger instance
func (logger *Logger) Debug(msg string, args ...any) {
	Logf(logger, LevelDebug, msg, args...)
}

func (logger *Logger) Info(msg string, args ...any) {
	Logf(logger, LevelInfo, msg, args...)
}

func (logger *Logger) Warn(msg string, args ...any) {
	Logf(logger, LevelWarn, msg, args...)
}

func (logger *Logger) Error(msg string, args ...any) {
	Logf(logger, LevelError, msg, args...)
}

func (logger *Logger) Fatal(msg string, args ...any) {
	Logf(logger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

func (logger *Logger) With(args ...any) *Logger {
	return &Logger{
		l: logger.l.With(args...),
	}
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "fatal":
		return CustomLevelFatal, nil
	}

	return 0, fmt.Errorf("not a valid slog Level: %q", level)
}

// The Logf function should be called inside a log wrapper function.
// Otherwise the code source reference will be invalid.
func Logf(logger *Logger, level, format string, args ...any) {
	if logger == nil {
		logger = &Logger{
			l: slog.Default(),
		}
	}

	logLevel, err := parseLevel(level)
	if err != nil {
		logger.Error("Error when parsing log level, err: %v", err)
		os.Exit(1)
	}

	if !logger.l.Enabled(context.Background(), logLevel) {
		return
	}
	var pcs [1]uintptr
	// skip [Callers, Log, LogWrapper]
	runtime.Callers(3, pcs[:]) //nolint:gomnd
	r := slog.NewRecord(time.Now(), logLevel, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.l.Handler().Handle(context.Background(), r)
}
