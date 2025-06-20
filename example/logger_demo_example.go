package main

import (
	"fmt"
	"io"
	"os"

	ar "github.com/bata94/apiright"
)

// CustomLogger demonstrates how to implement a custom logger
type CustomLogger struct {
	prefix string
}

func NewCustomLogger(prefix string) *CustomLogger {
	return &CustomLogger{prefix: prefix}
}

func (l *CustomLogger) log(level string, args ...any) {
	fmt.Printf("[%s] %s: %s\n", l.prefix, level, fmt.Sprint(args...))
}

func (l *CustomLogger) logf(level string, format string, args ...any) {
	fmt.Printf("[%s] %s: %s\n", l.prefix, level, fmt.Sprintf(format, args...))
}

func (l *CustomLogger) Debug(args ...any)                 { l.log("DEBUG", args...) }
func (l *CustomLogger) Debugf(format string, args ...any) { l.logf("DEBUG", format, args...) }
func (l *CustomLogger) Info(args ...any)                  { l.log("INFO", args...) }
func (l *CustomLogger) Infof(format string, args ...any)  { l.logf("INFO", format, args...) }
func (l *CustomLogger) Warn(args ...any)                  { l.log("WARN", args...) }
func (l *CustomLogger) Warnf(format string, args ...any)  { l.logf("WARN", format, args...) }
func (l *CustomLogger) Error(args ...any)                 { l.log("ERROR", args...) }
func (l *CustomLogger) Errorf(format string, args ...any) { l.logf("ERROR", format, args...) }
func (l *CustomLogger) Fatal(args ...any)                 { l.log("FATAL", args...); os.Exit(1) }
func (l *CustomLogger) Fatalf(format string, args ...any) { l.logf("FATAL", format, args...); os.Exit(1) }
func (l *CustomLogger) Panic(args ...any)                 { msg := fmt.Sprint(args...); l.log("PANIC", msg); panic(msg) }
func (l *CustomLogger) Panicf(format string, args ...any) { msg := fmt.Sprintf(format, args...); l.logf("PANIC", format, args...); panic(msg) }
func (l *CustomLogger) SetLevel(level ar.LogLevel)                {}
func (l *CustomLogger) GetLevel() ar.LogLevel                     { return ar.InfoLevel }
func (l *CustomLogger) SetOutput(output io.Writer)                {}

func demoLogger() {
	fmt.Println("=== Logger Demo ===")
	
	// Demo 1: Using default logger
	fmt.Println("\n1. Using default logger:")
	app1 := ar.NewApp()
	app1.Logger.Info("This is using the default logger")
	app1.Logger.Debugf("Debug message with parameter: %s", "test")
	
	// Demo 2: Using custom logger
	fmt.Println("\n2. Using custom logger:")
	app2 := ar.NewApp()
	customLogger := NewCustomLogger("CUSTOM")
	app2.SetLogger(customLogger)
	app2.Logger.Info("This is using a custom logger")
	app2.Logger.Debugf("Debug message with parameter: %s", "test")
	
	// Demo 3: Using package-level logger functions
	fmt.Println("\n3. Using package-level logger functions:")
	ar.Info("Package-level info message")
	ar.Debugf("Package-level debug message with parameter: %s", "test")
	
	// Demo 4: Setting log level
	fmt.Println("\n4. Setting log level to ERROR (debug messages won't show):")
	ar.SetLevel(ar.ErrorLevel)
	ar.Debug("This debug message won't show")
	ar.Error("This error message will show")
	
	// Reset level for normal operation
	ar.SetLevel(ar.InfoLevel)
	
	fmt.Println("\n=== End Logger Demo ===")
}

func main() {
	demoLogger()
}