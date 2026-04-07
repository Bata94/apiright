# APIRight

**A Go framework that auto-generates production-ready CRUD APIs from SQL schemas.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/bata94/apiright)](https://goreportcard.com/report/github.com/bata94/apiright)

## Why APIRight?

- **Fast**: Generate complete CRUD APIs in under 5 seconds
- **Simple**: Define your database schema, get an API
- **Flexible**: Customize with custom queries and plugins
- **Modern**: HTTP + gRPC with automatic content negotiation

## Library Usage

Import the root package for convenient access to all core types:

```go
import "github.com/bata94/apiright"

func main() {
    // Use types directly from root package
    cfg := apiright.DefaultConfig()
    logger, _ := apiright.NewLogger(false)
    ctx := apiright.NewGenerationContext(".")
    
    // Generate code
    gen, _ := apiright.NewGenerator(".", apiright.GenerateOptions{}, logger, nil)
    gen.Generate(ctx, apiright.GenerateOptions{})
}
```

Available exports include:
- **Types**: `Schema`, `Table`, `Column`, `Database`, `DualServer`, `Config`, `Generator`
- **Functions**: `NewLogger`, `NewGenerator`, `NewServer`, `LoadConfig`, `NewDatabase`
- **Constants**: Content types, log levels, status codes

## Features

### Server

| Feature | Description |
|---------|-------------|
| **Dual Protocol** | HTTP + gRPC serving from single source |
| **Protocol Toggle** | Enable/disable HTTP and gRPC independently |
| **CRUD Routes** | List, Get, Create, Update, Delete at `/{base_path}/{api_version}/{table}` |
| **Health Check** | `/health` endpoint with status and version info |
| **Service Registry** | Auto-loading generated services with mock support |
| **Middleware Pipeline** | HTTP and gRPC middleware chain with priority ordering |
| **Table Discovery** | Automatic discovery from SQL migration files |
| **Configurable Paths** | API version (`v0`, `v1`) and base path (`/api`, `/v1`) |

### Middleware

| Feature | Description |
|---------|-------------|
| **CORS** | Configurable origins, methods, headers, credentials |
| **Rate Limiting** | Per-client IP with sliding window algorithm |
| **Request Logging** | Structured logging with color support (dev mode) |
| **Validation** | Required, MinLen, MaxLen, Email, MinValue, MaxValue rules |
| **IP Extraction** | X-Forwarded-For, X-Real-IP header support |

### Content Negotiation

| Feature | Description |
|---------|-------------|
| **JSON** | Default format with proper Content-Type |
| **XML** | Full XML serialization with root element |
| **YAML** | YAML 1.1 serialization |
| **Protobuf** | Binary protobuf with proto descriptor support |
| **Plain Text** | Human-readable format |
| **q-value Parsing** | Accept header quality values (e.g., `Accept: application/xml;q=0.9`) |

### Code Generation

| Feature | Description |
|---------|-------------|
| **SQL CRUD** | Get, List, Create, Update, Delete with `_ar_gen` suffix |
| **Protobuf** | Message definitions and service stubs |
| **OpenAPI 3.0** | YAML and JSON specification files |
| **Go Services** | Service implementations using sqlc Querier |
| **Multi-dialect** | SQLite, PostgreSQL, MySQL support |
| **Generation Cache** | SHA-256 hash invalidation for incremental builds |

### Database

| Feature | Description |
|---------|-------------|
| **Multi-database** | SQLite, PostgreSQL, MySQL drivers |
| **Migrations** | Sequential migration execution |
| **Rollback** | Rollback last migration |
| **Transactions** | Atomic execution with rollback on failure |
| **Checksum Validation** | SHA-256 migration integrity tracking |

### Plugins

| Feature | Description |
|---------|-------------|
| **Hook System** | Before/after generation hooks |
| **Source Loader** | Load plugins from Go source files |
| **Security Validation** | Blocks exec, syscall, network, file deletion |
| **Middleware Provider** | Custom middleware via plugins |
| **Proto Extensions** | Custom protobuf extensions |

## Quick Start

### 1. Install

```bash
go install github.com/bata94/apiright/cmd/apiright@latest
```

### 2. Initialize Project

```bash
apiright init myapi --database sqlite
cd myapi
```

### 3. Define Schema

Edit `migrations/001_create_items_table.sql`:

```sql
CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);
```

### 4. Generate Code

```bash
apiright gen
```

### 5. Run Server

```bash
go run main.go
```

**Done!** API available at http://localhost:8080

## Generated API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v0/items` | List all items |
| POST | `/api/v0/items` | Create item |
| GET | `/api/v0/items/:id` | Get item by ID |
| PUT | `/api/v0/items/:id` | Update item |
| DELETE | `/api/v0/items/:id` | Delete item |

Plus gRPC at `localhost:9090`

## Content Negotiation

Request any format with the `Accept` header:

```bash
# JSON (default)
curl http://localhost:8080/api/v0/items

# XML
curl -H "Accept: application/xml" http://localhost:8080/api/v0/items

# YAML
curl -H "Accept: application/yaml" http://localhost:8080/api/v0/items

# Protobuf
curl -H "Accept: application/protobuf" http://localhost:8080/api/v0/items

# Plain Text
curl -H "Accept: text/plain" http://localhost:8080/api/v0/items
```

## Examples

| Example | Path | Description |
|---------|------|-------------|
| **Todo API** | `examples/todo/` | Full-featured app with all formats |
| **Blog Platform** | `examples/blog/` | Multiple tables with relationships |
| **Simple API** | `examples/simple-api/` | Minimal setup (< 100 lines) |

```bash
# Run the Todo example
cd examples/todo
go mod tidy
go run main.go --dev
```

## CLI Commands

```bash
apiright --help              # Show all commands
apiright init <name>         # Create new project
apiright gen                 # Generate CRUD code
apiright serve               # Start development server
apiright migrate up          # Run migrations
apiright migrate down        # Rollback last migration
apiright migrate status      # Check migration status
apiright db reset            # Reset database
```

## Configuration

Edit `apiright.yaml` to configure your server:

```yaml
server:
  enable_http: true           # Enable HTTP server (default: true)
  enable_grpc: true          # Enable gRPC server (default: true)
  api_version: v0            # API version prefix (default: v0)
                             # v0 = generated routes, v1 = your custom routes
  base_path: /api            # Base path prefix (default: /api)
                             # Routes: /api/v0/items
  http_port: 8080            # HTTP server port
  grpc_port: 9090            # gRPC server port
  host: localhost
  timeout: 30

database:
  type: sqlite               # sqlite, postgres, mysql
  name: app.db               # Database name or path
```

### Route Structure

| Config | Route Example |
|--------|---------------|
| Default | `/api/v0/items` |
| `base_path: /`, `api_version: v1` | `/v1/items` |
| `base_path: /myapp` | `/myapp/v0/items` |

### Protocol Toggle

```yaml
# HTTP only
server:
  enable_http: true
  enable_grpc: false

# gRPC only
server:
  enable_http: false
  enable_grpc: true
```

## Project Structure

```
myapi/
├── main.go                  # Your application
├── apiright.yaml           # Framework configuration
├── sqlc.yaml               # sqlc configuration
├── migrations/            # Database migrations
│   └── 001_create_items_table.sql
├── queries/               # Custom SQL queries (optional)
│   └── custom_queries.sql
└── gen/                   # Generated code (gitignored)
    ├── sql/              # Auto-generated CRUD queries
    ├── go/               # sqlc-generated Go code
    └── proto/            # Generated protobuf definitions
```

## Documentation

- [Getting Started Guide](docs/getting-started.md) - Complete tutorial
- [Plugin Development](docs/plugins/developer-guide.md) - Create custom plugins
- [Plugin Examples](examples/plugins/) - Working plugin code
- [PLAN.md](PLAN.md) - Development roadmap

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  SQL Schema │ ──► │ apiright gen│ ──► │   gen/       │
│             │     │             │     │   sql/       │
└─────────────┘     └─────────────┘     │   go/        │
                                         │   proto/     │
                                         └─────────────┘
                                                 │
                                                 ▼
                                         ┌─────────────┐
                                         │   Server    │
                                         │   HTTP:8080 │
                                         │   gRPC:9090 │
                                         └─────────────┘
```

## License

GNU General Public License v3.0 - See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! See [PLAN.md](PLAN.md) for the development roadmap.

## Status

✅ **Phase 1-3 Complete** - Framework core with all major features

🚀 **Phase 4 In Progress** - Polish, examples, documentation

📦 **v0.1.0 MVP** - Ready for testing

## Code Quality

```bash
just ci  # lint + vet + format + test
```

| Check | Status |
|-------|--------|
| go fmt | ✅ |
| go vet | ✅ |
| golangci-lint | ✅ (0 issues) |
| Tests | ✅ (32 passing) |
