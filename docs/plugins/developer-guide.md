# Plugin Development Guide

Learn how to create custom plugins for APIRight.

## Overview

Plugins extend APIRight by hooking into the code generation pipeline or adding runtime functionality.

### Plugin Types

1. **Generation Plugins** - Modify the code generation process
2. **Middleware Plugins** - Add HTTP/gRPC middleware
3. **Validation Plugins** - Add custom validation logic

## Getting Started

### 1. Create Plugin File

```go
package main

import (
    "github.com/bata94/apiright/pkg/core"
)

type MyPlugin struct {
    *core.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{
        BasePlugin: core.NewBasePlugin("my-plugin", "1.0.0"),
    }
}
```

### 2. Implement Plugin Interface

```go
func (p *MyPlugin) Generate(ctx *core.GenerationContext) error {
    // Your generation logic here
    ctx.Log("info", "Running my plugin")
    return nil
}

func (p *MyPlugin) Validate(schema *core.Schema) error {
    // Validate plugin configuration
    return nil
}
```

### 3. Compile as Plugin

```bash
mkdir -p plugins
go build -buildmode=plugin -o plugins/my-plugin.so
```

### 4. Load in Application

```go
// main.go
import "github.com/bata94/apiright/pkg/plugins"

func main() {
    registry := plugins.NewPluginRegistry(nil)
    loader := plugins.NewPluginLoader(registry)
    
    if err := loader.LoadFromFile("./plugins/my-plugin.so"); err != nil {
        log.Fatal(err)
    }
}
```

## Hook Points

Plugins can hook into the generation pipeline:

| Hook | When | Common Use |
|------|------|------------|
| `HookBeforeGeneration` | Before any generation | Prepare directories |
| `HookBeforeSQLC` | Before sqlc runs | Modify schema |
| `HookAfterSQLC` | After sqlc completes | Post-process Go code |
| `HookBeforeProtobuf` | Before proto generation | Add custom messages |
| `HookAfterProtobuf` | After proto completes | Post-process proto |
| `HookAfterGeneration` | After everything | Final validation |

### Registering Hooks

```go
func (p *MyPlugin) Generate(ctx *core.GenerationContext) error {
    ctx.RegisterHook(plugins.HookAfterGeneration, plugins.PluginHook{
        Name:     "my-hook",
        Priority: 100,
        Func:     p.myHookFunction,
    })
    return nil
}

func (p *MyPlugin) myHookFunction(ctx *core.GenerationContext) error {
    ctx.Log("info", "Hook executed!")
    return nil
}
```

## Example: Custom Logging Plugin

```go
package main

import (
    "os"
    "time"
    
    "github.com/bata94/apiright/pkg/core"
    "github.com/bata94/apiright/pkg/plugins"
)

type LoggingPlugin struct {
    *core.BasePlugin
    logFile *os.File
}

func NewLoggingPlugin() (*LoggingPlugin, error) {
    f, err := os.OpenFile("plugin.log", os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return nil, err
    }
    
    return &LoggingPlugin{
        BasePlugin: core.NewBasePlugin("custom-logging", "1.0.0"),
        logFile:    f,
    }, nil
}

func (p *LoggingPlugin) Name() string    { return "custom-logging" }
func (p *LoggingPlugin) Version() string { return "1.0.0" }

func (p *LoggingPlugin) Generate(ctx *core.GenerationContext) error {
    ctx.Log("info", "Custom logging plugin started")
    
    ctx.RegisterHook(plugins.HookAfterGeneration, plugins.PluginHook{
        Name:     "log-generation",
        Priority: 0,
        Func:     p.logGeneration,
    })
    
    return nil
}

func (p *LoggingPlugin) logGeneration(ctx *core.GenerationContext) error {
    timestamp := time.Now().Format(time.RFC3339)
    p.logFile.WriteString(timestamp + " - Generation completed\n")
    return nil
}

func (p *LoggingPlugin) Validate(schema *core.Schema) error {
    return nil
}
```

## Example: Custom Validation Plugin

```go
package main

import (
    "fmt"
    "strings"
    
    "github.com/bata94/apiright/pkg/core"
)

type ValidationPlugin struct {
    *core.BasePlugin
}

func NewValidationPlugin() *ValidationPlugin {
    return &ValidationPlugin{
        BasePlugin: core.NewBasePlugin("custom-validation", "1.0.0"),
    }
}

func (p *ValidationPlugin) Validate(schema *core.Schema) error {
    for _, table := range schema.Tables {
        if err := p.validateTableNaming(table); err != nil {
            return err
        }
    }
    return nil
}

func (p *ValidationPlugin) validateTableNaming(table *core.Table) error {
    name := table.Name
    if strings.HasPrefix(name, "tbl_") {
        return fmt.Errorf("table name %q should not use 'tbl_' prefix", name)
    }
    if strings.ToLower(name) != name {
        return fmt.Errorf("table name %q should be lowercase", name)
    }
    return nil
}

func (p *ValidationPlugin) Generate(ctx *core.GenerationContext) error {
    return nil
}
```

## Example: Middleware Plugin

```go
package main

import (
    "net/http"
    
    "github.com/bata94/apiright/pkg/core"
    "github.com/bata94/apiright/pkg/middleware"
)

type MiddlewareProvider interface {
    Middleware() []middleware.MiddlewareFunc
}

type MyMiddlewarePlugin struct {
    *core.BasePlugin
}

func NewMyMiddlewarePlugin() *MyMiddlewarePlugin {
    return &MyMiddlewarePlugin{
        BasePlugin: core.NewBasePlugin("custom-middleware", "1.0.0"),
    }
}

func (p *MyMiddlewarePlugin) Name() string    { return "custom-middleware" }
func (p *MyMiddlewarePlugin) Version() string { return "1.0.0" }

func (p *MyMiddlewarePlugin) Middleware() []middleware.MiddlewareFunc {
    return []middleware.MiddlewareFunc{
        p.myMiddleware,
    }
}

func (p *MyMiddlewarePlugin) myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Custom logic here
        r.Header.Set("X-Custom-Header", "Hello from plugin")
        next.ServeHTTP(w, r)
    })
}

func (p *MyMiddlewarePlugin) Generate(ctx *core.GenerationContext) error {
    return nil
}

func (p *MyMiddlewarePlugin) Validate(schema *core.Schema) error {
    return nil
}
```

## Configuration

Configure plugins in `apiright.yaml`:

```yaml
plugins:
  - name: "custom-logging"
    path: "./plugins/custom-logging.so"
    enabled: true
    version: "1.0.0"
    config:
      log_level: "debug"
      log_file: "app.log"
  
  - name: "custom-validation"
    enabled: true
    config:
      strict_naming: true
```

## Debugging

### Enable Debug Logging

```go
logger, _ := core.NewLoggerWithLevel("debug", true)
```

### Check Plugin Loading

```bash
# Run with verbose output
go run main.go --v
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Plugin not found | Check path is correct (relative to working directory) |
| Symbol not found | Ensure exported functions and types |
| Version mismatch | Check plugin was compiled with compatible Go version |
| Hook not called | Verify hook name matches exactly |

## Best Practices

1. **Use BasePlugin** - Provides common functionality
2. **Log everything** - Use ctx.Log() for debugging
3. **Validate early** - Check configuration in Validate()
4. **Handle errors** - Return meaningful error messages
5. **Test independently** - Unit test plugin logic separately

## Working Examples

See these files for complete, working examples:

- `examples/plugins/logging.go` - Custom logging
- `examples/plugins/validation.go` - Custom validation
- `examples/plugins/middleware.go` - Custom middleware

## Next Steps

- Review the [Plugin README](README.md)
- Check example applications in `examples/`
- Contribute plugins to the community!
