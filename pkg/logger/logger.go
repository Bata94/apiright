package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	PanicLevel LogLevel = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
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

func (l LogLevel) String() string {
	if int(l) < len(levelNames) {
		return levelNames[l]
	}
	return "UNKNOWN"
}

// Logger interface defines the logging contract
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
}

// DefaultLogger is a simple implementation of the Logger interface
type DefaultLogger struct {
	level  LogLevel
	output io.Writer
	logger *log.Logger
}

// NewDefaultLogger creates a new instance of DefaultLogger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  InfoLevel,
		output: os.Stdout,
		logger: log.New(os.Stdout, "", 0),
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

// SetOutput sets the output destination for the logger
func (l *DefaultLogger) SetOutput(output io.Writer) {
	l.output = output
	l.logger = log.New(output, "", 0)
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
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)

	l.logger.Println(logLine)
}

// log logs a message with the given level
func (l *DefaultLogger) log(level LogLevel, args ...any) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)

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
