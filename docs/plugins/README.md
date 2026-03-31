# Plugin System

APIRight provides a plugin system that allows you to extend the code generation pipeline and add custom functionality.

## Plugin Interface

All plugins must implement the `core.Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Generate(ctx *core.GenerationContext) error
    Validate(schema *core.Schema) error
}
```

### Methods

- `Name()`: Returns the plugin name
- `Version()`: Returns the plugin version
- `Generate()`: Executes the plugin's generation logic
- `Validate()`: Validates the plugin configuration

## Plugin Hooks

Plugins can hook into the generation pipeline at the following points:

| Hook Name | Description |
|-----------|-------------|
| `before_generation` | Called before any generation starts |
| `before_sqlc` | Called before sqlc execution |
| `after_sqlc` | Called after sqlc execution |
| `before_protobuf` | Called before protobuf generation |
| `after_protobuf` | Called after protobuf generation |
| `after_generation` | Called after all generation completes |

## Plugin Configuration

Plugins are configured in `apiright.yaml`:

```yaml
plugins:
  - name: "custom-logger"
    enabled: true
    version: "1.0.0"
    config:
      level: "debug"
      format: "json"
```

## Example Plugins

See the `examples/plugins/` directory for working plugin examples:

- `logging.go` - Custom logging plugin
- `validation.go` - Custom validation plugin
- `middleware.go` - Custom middleware plugin
