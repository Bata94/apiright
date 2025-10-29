package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry.
type LogLevel int

const (
	// PanicLevel level, highest level of severity.
	PanicLevel LogLevel = iota
	// FatalLevel level, logs and then calls os.Exit(1).
	FatalLevel
	// ErrorLevel level, used for errors that should definitely be noted.
	ErrorLevel
	// WarnLevel level, used for non-critical entries that deserve attention.
	WarnLevel
	// InfoLevel level, general operational entries about what's going on.
	InfoLevel
	// DebugLevel level, very verbose logging.
	DebugLevel
	// TraceLevel level, denotes finer-grained informational events than the Debug.
	TraceLevel
)

var levelNames = []string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"
)

// getColor returns the ANSI color code for a log level
func getColor(level LogLevel) string {
	switch level {
	case PanicLevel, FatalLevel:
		return colorRed
	case ErrorLevel:
		return colorRed
	case WarnLevel:
		return colorYellow
	case InfoLevel:
		return colorGreen
	case DebugLevel, TraceLevel:
		return colorCyan
	default:
		return colorWhite
	}
}

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	if int(l) < len(levelNames) {
		return levelNames[l]
	}
	return "UNKNOWN"
}

// Logger defines the interface for logging.
// This interface is compatible with logrus and other popular Go loggers
type Logger interface {
	// Standard logging methods
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)

	// Level management
	SetLevel(level LogLevel)
	GetLevel() LogLevel

	// Output management
	SetOutput(output io.Writer)

	// Color management
	SetColors(enabled bool)
}

// DefaultLogger is a simple implementation of the Logger interface
type DefaultLogger struct {
	level  LogLevel
	output io.Writer
	logger *log.Logger
	colors bool
}

// isTerminal checks if the writer is outputting to a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return f == os.Stdout || f == os.Stderr
	}
	return false
}

// NewDefaultLogger creates a new instance of DefaultLogger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  WarnLevel,
		output: os.Stdout,
		logger: log.New(os.Stdout, "", 0),
		colors: isTerminal(os.Stdout),
	}
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel returns the current logging level
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
}

// SetColors is not supported for StdLoggerWrapper
func (w *StdLoggerWrapper) SetColors(enabled bool) {
	// StdLoggerWrapper doesn't support colors
}

// SetOutput sets the output destination for the logger
func (l *DefaultLogger) SetOutput(output io.Writer) {
	l.output = output
	l.logger = log.New(output, "", 0)
	l.colors = isTerminal(output)
}

// SetColors enables or disables colored output
func (l *DefaultLogger) SetColors(enabled bool) {
	l.colors = enabled
}

// shouldLog checks if a message should be logged based on the current level
func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	return level <= l.level
}

// logf formats and logs a message with the given level
func (l *DefaultLogger) logf(level LogLevel, format string, args ...any) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	var logLine string
	if l.colors {
		color := getColor(level)
		logLine = fmt.Sprintf("%s[%s] %s: %s%s", color, timestamp, level.String(), message, colorReset)
	} else {
		logLine = fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)
	}

	l.logger.Println(logLine)
}

// log logs a message with the given level
func (l *DefaultLogger) log(level LogLevel, args ...any) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)

	var logLine string
	if l.colors {
		color := getColor(level)
		logLine = fmt.Sprintf("%s[%s] %s: %s%s", color, timestamp, level.String(), message, colorReset)
	} else {
		logLine = fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)
	}

	l.logger.Println(logLine)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(args ...any) {
	l.log(DebugLevel, args...)
}

// Debugf logs a formatted debug message
func (l *DefaultLogger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(args ...any) {
	l.log(InfoLevel, args...)
}

// Infof logs a formatted info message
func (l *DefaultLogger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(args ...any) {
	l.log(WarnLevel, args...)
}

// Warnf logs a formatted warning message
func (l *DefaultLogger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(args ...any) {
	l.log(ErrorLevel, args...)
}

// Errorf logs a formatted error message
func (l *DefaultLogger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}

// Fatal logs a fatal message and exits the program
func (l *DefaultLogger) Fatal(args ...any) {
	l.log(FatalLevel, args...)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *DefaultLogger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *DefaultLogger) Panic(args ...any) {
	message := fmt.Sprint(args...)
	l.log(PanicLevel, message)
	panic(message)
}

// Panicf logs a formatted panic message and panics
func (l *DefaultLogger) Panicf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	l.logf(PanicLevel, format, args...)
	panic(message)
}

// Global logger instance
var defaultLogger Logger = NewDefaultLogger()

// SetLogger allows users to replace the default logger with their own implementation
func SetLogger(logger Logger) {
	defaultLogger = logger
}

// GetLogger returns the current logger instance
func GetLogger() Logger {
	return defaultLogger
}

// Package-level convenience functions that use the default logger
func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

func Info(args ...any) {
	defaultLogger.Info(args...)
}

func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...any) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

func Fatal(args ...any) {
	defaultLogger.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	defaultLogger.Fatalf(format, args...)
}

func Panic(args ...any) {
	defaultLogger.Panic(args...)
}

func Panicf(format string, args ...any) {
	defaultLogger.Panicf(format, args...)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func GetLevel() LogLevel {
	return defaultLogger.GetLevel()
}

func SetOutput(output io.Writer) {
	defaultLogger.SetOutput(output)
}

// StdLoggerWrapper wraps the standard library log.Logger to implement our Logger interface
// This allows users to use stdlib log.Logger with APIRight
type StdLoggerWrapper struct {
	logger *log.Logger
	level  LogLevel
}

// NewStdLoggerWrapper creates a new wrapper around a standard library logger
func NewStdLoggerWrapper(stdLogger *log.Logger) *StdLoggerWrapper {
	return &StdLoggerWrapper{
		logger: stdLogger,
		level:  InfoLevel,
	}
}

// WrapStdLogger wraps the default standard library logger
func WrapStdLogger() *StdLoggerWrapper {
	return NewStdLoggerWrapper(log.Default())
}

func (w *StdLoggerWrapper) shouldLog(level LogLevel) bool {
	return level <= w.level
}

func (w *StdLoggerWrapper) Debug(args ...any) {
	if w.shouldLog(DebugLevel) {
		w.logger.Print(append([]any{"[DEBUG]"}, args...)...)
	}
}

func (w *StdLoggerWrapper) Debugf(format string, args ...any) {
	if w.shouldLog(DebugLevel) {
		w.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (w *StdLoggerWrapper) Info(args ...any) {
	if w.shouldLog(InfoLevel) {
		w.logger.Print(append([]any{"[INFO]"}, args...)...)
	}
}

func (w *StdLoggerWrapper) Infof(format string, args ...any) {
	if w.shouldLog(InfoLevel) {
		w.logger.Printf("[INFO] "+format, args...)
	}
}

func (w *StdLoggerWrapper) Warn(args ...any) {
	if w.shouldLog(WarnLevel) {
		w.logger.Print(append([]any{"[WARN]"}, args...)...)
	}
}

func (w *StdLoggerWrapper) Warnf(format string, args ...any) {
	if w.shouldLog(WarnLevel) {
		w.logger.Printf("[WARN] "+format, args...)
	}
}

func (w *StdLoggerWrapper) Error(args ...any) {
	if w.shouldLog(ErrorLevel) {
		w.logger.Print(append([]any{"[ERROR]"}, args...)...)
	}
}

func (w *StdLoggerWrapper) Errorf(format string, args ...any) {
	if w.shouldLog(ErrorLevel) {
		w.logger.Printf("[ERROR] "+format, args...)
	}
}

func (w *StdLoggerWrapper) Fatal(args ...any) {
	w.logger.Fatal(append([]any{"[FATAL]"}, args...)...)
}

func (w *StdLoggerWrapper) Fatalf(format string, args ...any) {
	w.logger.Fatalf("[FATAL] "+format, args...)
}

func (w *StdLoggerWrapper) Panic(args ...any) {
	w.logger.Panic(append([]any{"[PANIC]"}, args...)...)
}

func (w *StdLoggerWrapper) Panicf(format string, args ...any) {
	w.logger.Panicf("[PANIC] "+format, args...)
}

func (w *StdLoggerWrapper) SetLevel(level LogLevel) {
	w.level = level
}

func (w *StdLoggerWrapper) GetLevel() LogLevel {
	return w.level
}

func (w *StdLoggerWrapper) SetOutput(output io.Writer) {
	w.logger.SetOutput(output)
}

// SlogWrapper wraps the standard library slog.Logger to implement our Logger interface
// This allows users to use slog with APIRight for structured logging
type SlogWrapper struct {
	logger *slog.Logger
	level  LogLevel
}

// NewSlogWrapper creates a new wrapper around a slog logger
func NewSlogWrapper(slogLogger *slog.Logger) *SlogWrapper {
	return &SlogWrapper{
		logger: slogLogger,
		level:  InfoLevel,
	}
}

// WrapDefaultSlog wraps the default slog logger
func WrapDefaultSlog() *SlogWrapper {
	return NewSlogWrapper(slog.Default())
}

// NewSlogLogger creates a new slog logger with JSON handler
func NewSlogLogger(level LogLevel, output io.Writer) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case PanicLevel, FatalLevel:
		slogLevel = slog.LevelError
	case ErrorLevel:
		slogLevel = slog.LevelError
	case WarnLevel:
		slogLevel = slog.LevelWarn
	case InfoLevel:
		slogLevel = slog.LevelInfo
	case DebugLevel:
		slogLevel = slog.LevelDebug
	case TraceLevel:
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return slog.New(handler)
}

// NewColoredSlogLogger creates a new slog logger with colored text output
func NewColoredSlogLogger(level LogLevel, output io.Writer) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case PanicLevel, FatalLevel:
		slogLevel = slog.LevelError
	case ErrorLevel:
		slogLevel = slog.LevelError
	case WarnLevel:
		slogLevel = slog.LevelWarn
	case InfoLevel:
		slogLevel = slog.LevelInfo
	case DebugLevel:
		slogLevel = slog.LevelDebug
	case TraceLevel:
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add colors based on level
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				var color string
				switch {
				case level >= slog.LevelError:
					color = colorRed
				case level >= slog.LevelWarn:
					color = colorYellow
				case level >= slog.LevelInfo:
					color = colorGreen
				default:
					color = colorCyan
				}
				a.Value = slog.StringValue(color + a.Value.String() + colorReset)
			}
			return a
		},
	})

	return slog.New(handler)
}

// NewStructuredLogger creates a new SlogWrapper with JSON output
func NewStructuredLogger(level LogLevel, output io.Writer) *SlogWrapper {
	slogLogger := NewSlogLogger(level, output)
	return NewSlogWrapper(slogLogger)
}

// NewColoredStructuredLogger creates a new SlogWrapper with colored text output
func NewColoredStructuredLogger(level LogLevel, output io.Writer) *SlogWrapper {
	slogLogger := NewColoredSlogLogger(level, output)
	return NewSlogWrapper(slogLogger)
}

// parseArgs parses the arguments into message and attributes
func (w *SlogWrapper) parseArgs(args []any) (string, []slog.Attr) {
	if len(args) == 0 {
		return "", nil
	}

	// First argument is the message
	msg, ok := args[0].(string)
	if !ok {
		msg = fmt.Sprint(args[0])
	}

	// Remaining arguments are key-value pairs
	var attrs []slog.Attr
	for i := 1; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				attrs = append(attrs, slog.Any(key, args[i+1]))
			}
		}
	}

	return msg, attrs
}

func (w *SlogWrapper) shouldLog(level LogLevel) bool {
	return level <= w.level
}

func (w *SlogWrapper) Debug(args ...any) {
	msg, attrs := w.parseArgs(args)
	w.logger.LogAttrs(context.TODO(), slog.LevelDebug, msg, attrs...)
}

func (w *SlogWrapper) Debugf(format string, args ...any) {
	w.logger.Debug(fmt.Sprintf(format, args...))
}

func (w *SlogWrapper) Info(args ...any) {
	msg, attrs := w.parseArgs(args)
	w.logger.LogAttrs(context.TODO(), slog.LevelInfo, msg, attrs...)
}

func (w *SlogWrapper) Infof(format string, args ...any) {
	w.logger.Info(fmt.Sprintf(format, args...))
}

func (w *SlogWrapper) Warn(args ...any) {
	msg, attrs := w.parseArgs(args)
	w.logger.LogAttrs(context.TODO(), slog.LevelWarn, msg, attrs...)
}

func (w *SlogWrapper) Warnf(format string, args ...any) {
	w.logger.Warn(fmt.Sprintf(format, args...))
}

func (w *SlogWrapper) Error(args ...any) {
	msg, attrs := w.parseArgs(args)
	w.logger.LogAttrs(context.TODO(), slog.LevelError, msg, attrs...)
}

func (w *SlogWrapper) Errorf(format string, args ...any) {
	w.logger.Error(fmt.Sprintf(format, args...))
}

func (w *SlogWrapper) Fatal(args ...any) {
	w.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

func (w *SlogWrapper) Fatalf(format string, args ...any) {
	w.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (w *SlogWrapper) Panic(args ...any) {
	message := fmt.Sprint(args...)
	w.logger.Error(message)
	panic(message)
}

func (w *SlogWrapper) Panicf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	w.logger.Error(message)
	panic(message)
}

func (w *SlogWrapper) SetLevel(level LogLevel) {
	w.level = level
	// Note: slog level is handled by the handler, not the logger directly
	// Users should create a new logger with the desired level if they need to change it
}

func (w *SlogWrapper) GetLevel() LogLevel {
	return w.level
}

func (w *SlogWrapper) SetOutput(output io.Writer) {
	// Note: Changing output requires creating a new handler
	// This is a limitation of slog design
	// Users should create a new logger if they need different output
}

// SetColors is not supported for SlogWrapper (uses JSON format)
func (w *SlogWrapper) SetColors(enabled bool) {
	// SlogWrapper uses JSON format, colors don't apply
}
