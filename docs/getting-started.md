# Getting Started with APIRight

A step-by-step guide to building your first API with APIRight.

## Prerequisites

Before you begin, make sure you have:

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **SQLC** - `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **Protocol Buffers** - `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

Optional for gRPC:
- **protoc** - [Install protoc](https://grpc.io/docs/protoc-installation/)

## Step 1: Install APIRight

```bash
go install github.com/bata94/apiright/cmd/apiright@latest
```

Verify installation:
```bash
apiright --version
```

## Step 2: Create a New Project

```bash
apiright init myapi --database sqlite
cd myapi
```

This creates:
```
myapi/
├── main.go
├── apiright.yaml
├── sqlc.yaml
└── migrations/
    └── 001_create_items_table.sql
```

## Step 3: Define Your Schema

Edit `migrations/001_create_items_table.sql`:

```sql
CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

For a more complex schema with relationships, see the [Blog Example](../examples/blog/).

## Step 4: Generate Code

```bash
apiright gen
```

This generates:
- `gen/sql/*_ar_gen.sql` - CRUD queries
- `gen/go/` - sqlc-generated Go code
- `gen/proto/` - Protobuf definitions

## Step 5: Run the Server

```bash
go mod tidy
go run main.go --dev
```

Server starts at:
- **HTTP**: http://localhost:8080
- **gRPC**: localhost:9090

## Step 6: Test Your API

### Create an Item

```bash
curl -X POST http://localhost:8080/v1/items \
  -H "Content-Type: application/json" \
  -d '{"name":"My First Item","description":"Hello World"}'
```

### List Items

```bash
curl http://localhost:8080/v1/items
```

### Get Single Item

```bash
curl http://localhost:8080/v1/items/1
```

### Update Item

```bash
curl -X PUT http://localhost:8080/v1/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}'
```

### Delete Item

```bash
curl -X DELETE http://localhost:8080/v1/items/1
```

## Step 7: Try Different Formats

APIRight automatically handles content negotiation:

```bash
# JSON (default)
curl http://localhost:8080/v1/items

# XML
curl -H "Accept: application/xml" http://localhost:8080/v1/items

# YAML
curl -H "Accept: application/yaml" http://localhost:8080/v1/items

# Plain Text
curl -H "Accept: text/plain" http://localhost:8080/v1/items
```

## Adding Custom Queries

Sometimes you need more than basic CRUD. Add custom queries in `queries/`:

```sql
-- queries/items.sql
-- name: GetItemsByName :many
SELECT * FROM items WHERE name LIKE '%' || ? || '%';
```

Run `apiright gen` to regenerate with your custom queries.

## Adding Tables

1. Create a new migration: `migrations/002_create_categories_table.sql`

```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

ALTER TABLE items ADD COLUMN category_id INTEGER REFERENCES categories(id);
```

2. Run migrations: `apiright migrate up`
3. Regenerate code: `apiright gen`

## Using Plugins

Plugins add custom functionality. See [Plugin Examples](../examples/plugins/).

## Configuration

Edit `apiright.yaml`:

```yaml
server:
  http_port: 8080
  grpc_port: 9090

database:
  type: "sqlite"
  name: "myapi.db"

generation:
  content_types:
    - "application/json"
    - "application/xml"
```

## Next Steps

- Explore the [Todo Example](../examples/todo/) - Full-featured app
- Explore the [Blog Example](../examples/blog/) - Multiple tables
- Read the [Plugin Guide](../plugins/developer-guide.md)
- Check the [PLAN.md](../../PLAN.md) for upcoming features

## Common Issues

### "sqlc not found"
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### "protoc not found"
```bash
# macOS
brew install protobuf

# Linux
sudo apt install protobuf-compiler
```

### Migration fails
```bash
# Reset database
apiright db reset

# Or rollback and re-run
apiright migrate down
apiright migrate up
```

## Need Help?

- Open an issue on GitHub
- Check existing examples in `examples/`
- Review PLAN.md for feature roadmap
