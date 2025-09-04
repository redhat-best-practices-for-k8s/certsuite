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

// Logger Encapsulates a structured logger with convenience methods
//
// This type wraps an slog.Logger to provide simple debug, info, warn, error,
// fatal, and context‑aware logging functions. It also offers a With method
// that attaches key/value pairs to the underlying logger, returning a new
// Logger instance for fluent chaining.
type Logger struct {
	l *slog.Logger
}

var (
	globalLogger   *Logger
	globalLogLevel slog.Level
	globalLogFile  *os.File
)

// CreateGlobalLogFile Creates or replaces the global log file for test output
//
// The function removes any existing log file in the specified directory, then
// opens a new one with read/write permissions. It configures the logger to
// write to this file at the requested level and stores the file handle
// globally. Errors during removal or opening are returned as formatted
// messages.
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

// CloseGlobalLogFile Closes the globally opened log file
//
// The function invokes the Close method on the global log file handle and
// returns any error that occurs during closure. It does not take any arguments
// and only provides an error result indicating success or failure of the
// operation.
func CloseGlobalLogFile() error {
	return globalLogFile.Close()
}

// SetupLogger configures global logging with a custom level and writer
//
// This function parses the supplied log level string, falling back to INFO if
// parsing fails, and sets the global logger to write formatted slog entries to
// the provided io.Writer. It uses a custom handler that replaces standard level
// strings with user‑defined names when necessary. The resulting Logger
// instance is stored globally for use throughout the application.
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

// SetLogger Sets the package-wide logger instance
//
// This function assigns the provided Logger to a global variable used
// throughout the logging package, making it available for all subsequent log
// operations. It performs no validation or side effects beyond the assignment
// and does not return any value.
func SetLogger(l *Logger) {
	globalLogger = l
}

// GetLogger Retrieves the package-wide logger instance
//
// This function provides access to a globally shared Logger object that is used
// throughout the application for consistent logging behavior. It simply returns
// the reference stored in the internal variable, allowing other packages to
// obtain and use the same logger without creating new instances.
func GetLogger() *Logger {
	return globalLogger
}

// GetMultiLogger Creates a logger that writes to multiple destinations
//
// The function builds a set of slog handlers, including an optional global
// handler if one is configured, and wraps each supplied writer in a custom
// handler with the current log level settings. It then combines these handlers
// into a multi-handler so that every log record is emitted to all specified
// writers simultaneously. The resulting Logger instance is returned for use
// throughout the application.
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

// Debug Logs a message at the debug level
//
// This function forwards its arguments to the internal logging system, tagging
// them with a debug severity. It accepts a format string followed by any number
// of values, which are passed to the underlying logger for formatting and
// output. The global logger instance is used, ensuring consistent log
// configuration across the application.
func Debug(msg string, args ...any) {
	Logf(globalLogger, LevelDebug, msg, args...)
}

// Info Logs a message at the informational level
//
// This function sends a formatted log entry to the package's global logger with
// an informational severity. It accepts a message string and optional
// arguments, which are passed through to formatting before dispatching to the
// underlying logging system. The call is non‑blocking and does not return any
// value.
func Info(msg string, args ...any) {
	Logf(globalLogger, LevelInfo, msg, args...)
}

// Warn Logs a message at warning level
//
// The function forwards its arguments to Logf, supplying the global logger and
// a warning severity indicator. It accepts a format string followed by optional
// values, which are interpolated into the log entry. The resulting record is
// written using slog's handling mechanisms.
func Warn(msg string, args ...any) {
	Logf(globalLogger, LevelWarn, msg, args...)
}

// Error Logs an error message with optional formatting
//
// This function accepts a format string and optional arguments, then forwards
// the call to a lower-level logging routine that writes the message at the
// error severity level. It uses the global logger instance, ensuring
// consistency across the application. The formatted output is sent to the
// configured log handler.
func Error(msg string, args ...any) {
	Logf(globalLogger, LevelError, msg, args...)
}

// Fatal Logs a fatal error message and terminates the program
//
// This function writes the supplied formatted message to both the configured
// logger at the fatal level and directly to standard error. After logging, it
// exits the process with status code one, ensuring that the application stops
// immediately.
func Fatal(msg string, args ...any) {
	Logf(globalLogger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// Logger.Debug Logs a formatted message at debug level
//
// The method calls the generic logging helper Logf, passing the logger instance
// and the debug log level together with the supplied format string and
// arguments. It formats the message using fmt.Sprintf before emitting it
// through the underlying slog handler, only if the current logger is enabled
// for debug logs.
func (logger *Logger) Debug(msg string, args ...any) {
	Logf(logger, LevelDebug, msg, args...)
}

// Logger.Info Logs an informational message
//
// This method forwards the supplied format string and arguments to the internal
// logging routine at the info level. It relies on Logf to create a log record
// with the appropriate severity, ensuring the message is emitted only if the
// logger’s configuration allows that level. No value is returned.
func (logger *Logger) Info(msg string, args ...any) {
	Logf(logger, LevelInfo, msg, args...)
}

// Logger.Warn Logs a warning message with optional formatting
//
// This method takes a format string and an arbitrary number of arguments,
// passes them to the underlying Logf function along with the warning level
// constant. It records the warning using the logger's handler if the warning
// level is enabled for the current context. The call does not return any value.
func (logger *Logger) Warn(msg string, args ...any) {
	Logf(logger, LevelWarn, msg, args...)
}

// Logger.Error Logs a formatted message at the error level
//
// This method receives a format string followed by optional arguments, then
// delegates to the generic logging helper passing the error severity. It uses
// the Logger instance if provided; otherwise it falls back to the default
// logger. The resulting entry is emitted immediately without returning any
// value.
func (logger *Logger) Error(msg string, args ...any) {
	Logf(logger, LevelError, msg, args...)
}

// Logger.Fatal Outputs a fatal error message, writes to stderr and exits the program
//
// The method logs a formatted fatal message using the Logger’s Logf helper,
// then prints the same message prefixed with "FATAL:" to standard error for
// visibility. After displaying the message it terminates the process by calling
// os.. No return value is produced because execution stops immediately.
func (logger *Logger) Fatal(msg string, args ...any) {
	Logf(logger, LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// Logger.With Creates a child logger with added contextual fields
//
// The method accepts any number of key-value pairs or structured arguments and
// forwards them to the underlying logger’s With function. It constructs a new
// Logger instance that preserves the original logger while extending its
// context, allowing subsequent log entries to include these additional fields.
// The returned logger can be used independently for further logging calls.
func (logger *Logger) With(args ...any) *Logger {
	return &Logger{
		l: logger.l.With(args...),
	}
}

// parseLevel Converts a string into a slog logging level
//
// The function takes a textual log level, normalizes it to lowercase, and
// matches it against known levels such as debug, info, warn, error, and fatal.
// If the input corresponds to one of these names, the matching slog.Level
// constant is returned; otherwise an error is produced indicating the value is
// invalid.
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

// Logf Logs a formatted message at the specified level
//
// The function accepts a logger, a string representing the log level, a format
// string, and optional arguments. It parses the level, checks if logging is
// enabled for that level, retrieves the caller information, creates a slog
// record with a timestamp and formatted message, and passes it to the
// logger’s handler.
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
