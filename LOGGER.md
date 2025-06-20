# APIRight Logger

APIRight includes a simple but extensible logging package that is compatible with popular Go loggers like logrus.

## Features

- **Default Logger**: Built-in logger with timestamp formatting
- **Replaceable**: Easy to replace with your own logger implementation
- **Log Levels**: Support for Panic, Fatal, Error, Warn, Info, Debug, and Trace levels
- **Logrus Compatible**: Interface compatible with logrus and other popular loggers
- **Package-level Functions**: Convenient package-level logging functions

## Usage

### Using the Default Logger

```go
app := apiright.NewApp()
app.Logger.Info("Application started")
app.Logger.Debugf("Debug message with parameter: %s", "value")
```

### Using Package-level Functions

```go
import "github.com/bata94/apiright"

apiright.Info("This is an info message")
apiright.Errorf("Error occurred: %v", err)
apiright.SetLevel(apiright.DebugLevel)
```

### Replacing with Custom Logger

```go
// Implement the Logger interface
type MyCustomLogger struct {
    // your fields
}

func (l *MyCustomLogger) Info(args ...interface{}) {
    // your implementation
}
// ... implement all other Logger interface methods

// Use your custom logger
app := apiright.NewApp()
app.SetLogger(&MyCustomLogger{})
```

### Using with Logrus

```go
import (
    "github.com/sirupsen/logrus"
    "github.com/bata94/apiright"
)

// Logrus is already compatible with the Logger interface
logger := logrus.New()
app := apiright.NewApp()
app.SetLogger(logger)
```

### Using with Standard Library log.Logger

```go
import (
    "log"
    "os"
    "github.com/bata94/apiright"
)

// Wrap the standard library logger
stdLogger := log.New(os.Stdout, "APP: ", log.LstdFlags)
wrappedLogger := apiright.NewStdLoggerWrapper(stdLogger)

app := apiright.NewApp()
app.SetLogger(wrappedLogger)

// Or use the default standard logger
app.SetLogger(apiright.WrapStdLogger())
```

## Log Levels

- `PanicLevel`: Logs and then calls panic
- `FatalLevel`: Logs and then calls os.Exit(1)
- `ErrorLevel`: Error conditions
- `WarnLevel`: Warning conditions
- `InfoLevel`: General information (default level)
- `DebugLevel`: Debug information
- `TraceLevel`: Very detailed debug information

## Logger Interface

The Logger interface is designed to be compatible with logrus and other popular Go loggers:

```go
type Logger interface {
    Debug(args ...interface{})
    Debugf(format string, args ...interface{})
    Info(args ...interface{})
    Infof(format string, args ...interface{})
    Warn(args ...interface{})
    Warnf(format string, args ...interface{})
    Error(args ...interface{})
    Errorf(format string, args ...interface{})
    Fatal(args ...interface{})
    Fatalf(format string, args ...interface{})
    Panic(args ...interface{})
    Panicf(format string, args ...interface{})
    
    SetLevel(level LogLevel)
    GetLevel() LogLevel
    SetOutput(output io.Writer)
}
```

## Integration with APIRight App

The logger is automatically integrated into the APIRight App and used for:
- Route registration logging
- Error handling and recovery
- Request processing information

All internal logging uses the configured logger, so you can control APIRight's logging behavior by setting your preferred logger implementation.