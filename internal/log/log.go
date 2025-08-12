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
	LogFileName        = "certsuite.log"
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

// Logger provides a convenient wrapper around slog.Logger to emit structured logs at various levels.
//
// It holds an underlying *slog.Logger instance and exposes methods such as Debug, Info, Warn, Error, Fatal, and With.
// Each method forwards formatted messages to the underlying logger with the appropriate severity level,
// automatically including contextual information like timestamps and caller details. The With method
// allows adding key/value pairs that are attached to all subsequent log entries produced by the returned Logger.
type Logger struct {
	l *slog.Logger
}

var (
	globalLogger   *Logger
	globalLogLevel slog.Level
	globalLogFile  *os.File
)

// CreateGlobalLogFile creates and configures a global log file.
//
// It removes any existing file at the specified path, then opens a new
// file with the given permissions. If opening the file fails it returns an error.
// On success it sets up the global logger to write to this file and updates
// the package level globals. The function returns an error if any step of
// the process fails.
func CreateGlobalLogFile(outputDir, logLevel string) error {
	logFilePath := outputDir + "/" + LogFileName
	err := os.Remove(logFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not delete old log file, err: %v", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, LogFilePermissions)
	if err != nil {
		return fmt.Errorf("could not open a new log file, err: %v", err)
	}

	SetupLogger(logFile, logLevel)
	globalLogFile = logFile

	return nil
}

// CloseGlobalLogFile closes the global log file used by the package.
//
// It calls os.File.Close on the internally stored file handle and returns any
// error that occurs during the close operation. If the file is already closed or
// nil, it simply returns nil. The function does not affect the global logger
// state; it only ensures that the underlying file resource is released.
func CloseGlobalLogFile() error {
	return globalLogFile.Close()
}

// SetupLogger configures a global logger with the specified writer and log level string, returning a cleanup function that closes any opened log file.
//
// It parses the supplied level string into an slog.Level, creates a new Logger instance using a custom handler,
// assigns it to the package‑wide globalLogger variable, and stores the parsed level in globalLogLevel.
// If the logger writes to a file, it opens or reopens the file with LogFileName and LogFilePermissions.
// The returned function should be called when the application exits to close the log file if one was opened.
func SetupLogger(logWriter io.Writer, level string) {
	logLevel, err := parseLevel(level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse log level, err: %v. Defaulting to DEBUG.", err)
		globalLogLevel = slog.LevelInfo
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

// SetLogger configures the global logger instance.
//
// It accepts a pointer to a Logger and returns a function that restores
// the previous logger when called. The returned function should be deferred
// by callers to ensure the original logger is reinstated after temporary
// changes. This allows tests or components to temporarily redirect logging
// output without affecting other parts of the application.
func SetLogger(l *Logger) {
	globalLogger = l
}

// GetLogger returns the singleton logger instance.
//
// It lazily creates a Logger if one does not already exist and
// ensures that only a single global logger is used throughout the package.
// The returned *Logger can be used to log messages at various levels.
// This function has no parameters and returns a pointer to the shared Logger.
func GetLogger() *Logger {
	return globalLogger
}

// GetMultiLogger creates a logger that writes to multiple writers.
//
// It accepts any number of io.Writer arguments and constructs a multi-handler
// that forwards log records to each writer. The returned *Logger is configured
// with the default level and formatter, allowing callers to use it as a drop‑in
// replacement for the global logger while directing output to custom destinations.
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

	var handlers []slog.Handler
	if globalLogger != nil {
		handlers = []slog.Handler{globalLogger.l.Handler()}
	}

	for _, writer := range writers {
		handlers = append(handlers, NewCustomHandler(writer, &opts))
	}

	return &Logger{l: slog.New(NewMultiHandler(handlers...))}
}

// Debug logs a message at the debug level.
//
// It accepts a format string and optional arguments, which are passed to Logf
// along with the debug log level. The function returns immediately after
// invoking Logf.
func Debug(msg string, args ...any) {
	Logf(globalLogger, LevelDebug, msg, args...)
}

// Info logs a message at the info level.
//
// It accepts a format string and optional arguments, which are passed to Logf
// for formatting. The resulting log entry is written using the current global
// logger configuration. No value is returned.
func Info(msg string, args ...any) {
	Logf(globalLogger, LevelInfo, msg, args...)
}

// Warn logs a warning message.
//
// It formats the provided string with optional arguments and writes the
// result to the global logger at the warning level. The function does not
// return any value; it simply records the event.
func Warn(msg string, args ...any) {
	Logf(globalLogger, LevelWarn, msg, args...)
}

// Error logs a message at the error level and returns a no-op function.
//
// It accepts a format string and optional arguments, formats them,
// and writes the result to the configured logger using Logf.
// The returned function is intended to be deferred but performs no action.
func Error(msg string, args ...any) {
	Logf(globalLogger, LevelError, msg, args...)
}

// Fatal logs a message at the fatal level and terminates the program.
//
// It accepts a format string followed by optional arguments, writes the formatted
// output to the configured logger, then calls os.Exit(1) to exit the process.
func Fatal(msg string, args ...any) {
	Logf(globalLogger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// Debug logs a message at the debug level.
//
// It accepts a format string followed by optional arguments, formats them
// using Logf, and writes the result to the logger's output if the current
// log level permits debug messages. The function returns immediately,
// allowing callers to ignore the returned value.
func (logger *Logger) Debug(msg string, args ...any) {
	Logf(logger, LevelDebug, msg, args...)
}

// Info logs a message at the info level.
//
// It accepts a format string and optional arguments, formats the message
// using fmt.Sprintf semantics, and writes it to the logger's output
// with the log level set to LevelInfo. The function does not return any
// value; any errors from the underlying Logf call are ignored.
func (logger *Logger) Info(msg string, args ...any) {
	Logf(logger, LevelInfo, msg, args...)
}

// Warn logs a message at the warning level.
//
// It accepts a format string and optional arguments, formats them,
// and writes the result using the logger's underlying Logf method
// with the warning log level. The function returns no value.
func (logger *Logger) Warn(msg string, args ...any) {
	Logf(logger, LevelWarn, msg, args...)
}

// Error logs a message at the error level.
//
// It accepts a format string and optional arguments, formats them,
// and writes the result to the logger using the underlying Logf method.
func (logger *Logger) Error(msg string, args ...any) {
	Logf(logger, LevelError, msg, args...)
}

// ```go
// Fatal logs a fatal message and exits the application.
//
// It formats the message using fmt.Sprintf style arguments, writes it to the log,
// then terminates the program with os.Exit(1). The method accepts a format string
// followed by optional arguments that are interpolated into the format.
// After logging, no further code in the current goroutine will execute.```
func (logger *Logger) Fatal(msg string, args ...any) {
	Logf(logger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// With creates a new logger with additional attributes.
//
// It accepts any number of key/value pairs and returns a new *Logger
// that contains the supplied attributes in its context. The returned
// logger can be used to log messages that automatically include these
// attributes. The original logger remains unchanged.
func (logger *Logger) With(args ...any) *Logger {
	return &Logger{
		l: logger.l.With(args...),
	}
}

// parseLevel converts a string representation of a log level into an slog.Level value and returns an error if the input is invalid.
//
// It accepts a single string argument, normalizes it to lower case,
// looks up the corresponding slog.Level in the predefined level maps,
// and returns that level along with nil error. If no matching level is found,
// it returns an error describing the unknown level. The function is used internally
// by the logging package to parse configuration values into log levels.
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

// Logf logs a formatted message at the specified level using the provided Logger.
//
// It should be called from within a wrapper function; otherwise, the source
// reference will point to this helper instead of the caller.
// The function accepts a Logger pointer, a log level string, a format string,
// and optional arguments. It determines if logging is enabled for that level,
// creates a record with the current timestamp, formats the message,
// and passes the record to the logger's handler. If the level corresponds
// to Fatal, it will terminate the program after logging. The function
// returns an empty tuple.
func Logf(logger *Logger, level, format string, args ...any) {
	if logger == nil {
		logger = &Logger{
			l: slog.Default(),
		}
	}

	logLevel, err := parseLevel(level)
	if err != nil {
		logger.Fatal("Error when parsing log level, err: %v", err)
	}

	if !logger.l.Enabled(context.TODO(), logLevel) {
		return
	}
	var pcs [1]uintptr
	// skip [Callers, Log, LogWrapper]
	runtime.Callers(3, pcs[:]) //nolint:mnd
	r := slog.NewRecord(time.Now(), logLevel, fmt.Sprintf(format, args...), pcs[0])
	_ = logger.l.Handler().Handle(context.TODO(), r)
}
