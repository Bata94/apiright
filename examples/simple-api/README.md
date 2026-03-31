# Simple API - Minimal Example

The smallest possible APIRight project. 4 files, under 100 lines total.

## Files

```
simple-api/
├── main.go                    # 55 lines - Server setup
├── apiright.yaml              # 12 lines - Configuration
├── sqlc.yaml                  # 11 lines - sqlc setup
└── migrations/
    └── 001_create_items_table.sql  # 6 lines - Schema
```

**Total: 4 files, ~85 lines**

## Quick Start (2 minutes)

```bash
# 1. Create new project
go run github.com/bata94/apiright/cmd/apiright init my-api --database sqlite
cd my-api

# 2. Define schema
# Edit migrations/001_create_items_table.sql

# 3. Generate code
go run github.com/bata94/apiright/cmd/apiright gen

# 4. Start server
go run main.go
```

## Minimal Schema

```sql
CREATE TABLE items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);
```

## Generated API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/items | List all items |
| POST | /v1/items | Create item |
| GET | /v1/items/1 | Get item #1 |
| PUT | /v1/items/1 | Update item #1 |
| DELETE | /v1/items/1 | Delete item #1 |

## Minimal main.go

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/middleware"
	"github.com/bata94/apiright/pkg/server"
)

func main() {
	logger, _ := core.NewLogger(true)
	defer logger.Sync()

	cfg, _ := config.LoadConfig(".")
	db, _ := database.NewDatabase(&cfg.Database, logger)
	db.Connect()
	db.Migrate()

	srv := server.NewServer(&cfg.Server, db, logger)
	srv.RegisterGeneratedServices(".")

	go srv.Start(context.Background())
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
```

## Learning Path

1. Run the server, test the API
2. Add a column to the schema
3. Re-run `apiright gen`
4. Add custom queries in `queries/`
5. Enable more content types

## This is the fastest way to:
- Create a new API in 2 minutes
- Prototype a data model
- Test APIRight features
- Get started with the framework
