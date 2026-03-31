# Todo API - Complete Example

This example demonstrates all APIRight framework features in a single, runnable Todo application.

## Features Demonstrated

| Feature | Status | Description |
|---------|--------|-------------|
| SQL Generation | ✅ | CRUD queries with `_ar_gen` suffix |
| Content Negotiation | ✅ | All 5 formats: JSON, XML, YAML, Protobuf, Text |
| HTTP Server | ✅ | REST endpoints with `/v1/todos` |
| gRPC Server | ✅ | gRPC methods for all CRUD operations |
| Middleware | ✅ | Logging, Request ID, Recovery, Custom plugins |
| Plugin System | ✅ | Dynamic plugin loading |
| Database Migrations | ✅ | Automatic migration on startup |
| Validation | ✅ | Request/response validation |
| OpenAPI Docs | ✅ | Auto-generated API documentation |

## Quick Start

```bash
# Navigate to the example
cd examples/todo

# Install dependencies
go mod tidy

# Run migrations and start server
go run main.go --dev
```

Server starts at:
- **HTTP**: http://localhost:8080
- **gRPC**: localhost:9090

## API Examples

### Create Todo (JSON)
```bash
curl -X POST http://localhost:8080/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn APIRight","priority":1,"due_date":"2026-04-01T00:00:00Z"}'
```

### Get Todos (XML)
```bash
curl http://localhost:8080/v1/todos \
  -H "Accept: application/xml"
```

### Get Todos (YAML)
```bash
curl http://localhost:8080/v1/todos \
  -H "Accept: application/yaml"
```

### Get Todos (Protobuf)
```bash
curl http://localhost:8080/v1/todos \
  -H "Accept: application/protobuf"
```

### Get Todos (Plain Text)
```bash
curl http://localhost:8080/v1/todos \
  -H "Accept: text/plain"
```

### Get Single Todo
```bash
curl http://localhost:8080/v1/todos/1
```

### Update Todo
```bash
curl -X PUT http://localhost:8080/v1/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","completed":true}'
```

### Delete Todo
```bash
curl -X DELETE http://localhost:8080/v1/todos/1
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Content Negotiation

The API automatically detects the best format based on the `Accept` header:

| Accept Header | Format | Priority |
|---------------|--------|----------|
| `application/json` | JSON | Highest |
| `application/xml` | XML | High |
| `application/yaml` | YAML | Medium |
| `application/protobuf` | Protobuf | Medium |
| `text/plain` | Plain Text | Lowest |
| (none) | JSON | Default |

### Quality Values (q-values)

The server respects quality values in Accept headers:
```bash
# Prefer JSON but accept YAML
curl http://localhost:8080/v1/todos \
  -H "Accept: application/json;q=0.9,application/yaml;q=1.0"
```

## Middleware Stack

Requests pass through this middleware chain:

1. **Recovery** - Panic recovery, returns 500 on crash
2. **Request ID** - Adds unique ID to each request
3. **Logging** - Structured request/response logging
4. **Content Negotiation** - Format detection and serialization

## Custom Queries

The example includes custom queries in `queries/todos.sql`:

```sql
-- Get todos by completion status
SELECT * FROM todos WHERE completed = ?

-- Get high priority todos
SELECT * FROM todos WHERE priority >= ?

-- Get overdue todos
SELECT * FROM todos WHERE due_date < datetime('now') AND completed = FALSE

-- Get statistics
SELECT COUNT(*) as total, SUM(CASE WHEN completed THEN 1 ELSE 0 END) as completed FROM todos
```

## Plugin System

Plugins are loaded from `./plugins/` directory:

- `logging.so` - Custom logging plugin
- `validation.so` - Custom validation rules
- `middleware.so` - Custom middleware

### Loading Plugins

Plugins are loaded dynamically at startup:

```go
if err := loadPlugins(logger, mwRegistry); err != nil {
    logger.Warn("Some plugins failed to load", core.Error(err))
}
```

### Creating a Plugin

1. Implement the `Plugin` interface
2. Compile with: `go build -buildmode=plugin -o plugins/myplugin.so`
3. Place in `./plugins/` directory
4. Restart the server

See `examples/plugins/` for examples.

## Project Structure

```
examples/todo/
├── main.go                  # Application entry point
├── apiright.yaml            # Framework configuration
├── sqlc.yaml               # sqlc configuration
├── todo.db                 # SQLite database (auto-created)
├── migrations/             # Database migrations
│   └── 001_create_todos_table.sql
├── queries/                # Custom SQL queries
│   └── todos.sql
├── gen/                    # Generated code (gitignored)
│   ├── sql/               # Auto-generated CRUD queries
│   ├── go/                # sqlc-generated Go code
│   └── proto/             # Generated protobuf definitions
└── plugins/               # Plugin directory
    ├── logging.so
    ├── validation.so
    └── middleware.so
```

## Configuration

Edit `apiright.yaml` to customize:

```yaml
server:
  http_port: 8080
  grpc_port: 9090
  host: "localhost"

database:
  type: "sqlite"
  name: "todo.db"

generation:
  output_dir: "gen"
  gen_suffix: "_ar_gen"
  content_types:
    - "application/json"
    - "application/xml"
    - "application/yaml"
    - "application/protobuf"
    - "text/plain"
```

## CLI Commands

```bash
# Generate code
go run ../cmd/apiright gen

# Run migrations
go run ../cmd/apiright migrate up

# Check migration status
go run ../cmd/apiright migrate status

# Reset database
go run ../cmd/apiright db reset
```

## Development Mode

Use `--dev` flag for colored, human-readable logs:

```bash
go run main.go --dev
```

Use `--v` for verbose debug output:

```bash
go run main.go --dev --v
```

Production mode outputs JSON logs for log aggregation.

## Learning Path

1. **Start here**: Run the server, test the endpoints
2. **Read migrations**: See how schema changes are applied
3. **Check queries**: Understand custom vs generated queries
4. **Review generated code**: `gen/` directory contains all generated files
5. **Try plugins**: Build and load a custom plugin
6. **Modify schema**: Add columns, regenerate, see changes

## Next Steps

- Add more tables to `migrations/`
- Write custom queries in `queries/`
- Create custom protobuf extensions in `proto/`
- Build and load plugins from `examples/plugins/`
