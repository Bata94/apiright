# ===== Configuration =====
BINARY := "apiright"
MAIN := "cmd/main.go"
PKGS := "./..."
TEST_PKGS := "./tests/..."

# ===== Default =====
# List all available recipes
default: list

# ===== Build & Run =====
# Compile the apiright binary
build:
    go build -o {{BINARY}} {{MAIN}}

# Build and run apiright with arguments
run +args:
    go run {{MAIN}} {{args}}

# ===== Code Quality =====
# Format all Go source files
fmt:
    go fmt {{PKGS}}

# Run Go's static analysis
vet:
    go vet {{PKGS}}

# Run golangci-lint for comprehensive linting
lint:
    golangci-lint run ./...

# Run fmt, vet, and tests
check: fmt vet test

# ===== Testing =====
# Run all tests
test:
    go test {{PKGS}}

# Run tests with verbose output
test-verbose:
    go test -v {{PKGS}}

# Run performance benchmarks
bench:
    go test -bench=. -benchmem {{TEST_PKGS}}

# ===== Dependencies =====
# Tidy go.mod and go.sum
tidy:
    go mod tidy

# Download all dependencies
deps:
    go mod download

# ===== APIRight CLI Commands =====
# Generate CRUD operations from SQL schema
gen:
    go run {{MAIN}} gen

# Generate SQL queries only
gen-sql:
    go run {{MAIN}} gen --sql-only

# Generate Go code only
gen-go:
    go run {{MAIN}} gen --go-only

# Generate protobuf definitions only
gen-proto:
    go run {{MAIN}} gen --proto-only

# Force regeneration bypassing cache
gen-force:
    go run {{MAIN}} gen --force

# Start the development server
serve:
    go run {{MAIN}} serve

# Start server with development mode (hot reload)
serve-dev:
    go run {{MAIN}} serve --dev

# Run pending database migrations
migrate:
    go run {{MAIN}} migrate up

# Rollback the last migration
migrate-down:
    go run {{MAIN}} migrate down

# Show migration status
migrate-status:
    go run {{MAIN}} migrate status

# Create a new migration file (provide name: just migrate-create add_users_table)
migrate-create name:
    go run {{MAIN}} migrate create {{name}}

# Create the database
db-create:
    go run {{MAIN}} db create

# Drop the database
db-drop:
    go run {{MAIN}} db drop

# Drop and recreate the database
db-reset:
    go run {{MAIN}} db reset

# Test database connectivity
db-ping:
    go run {{MAIN}} db ping

# Run seed data
db-seed:
    go run {{MAIN}} db seed

# Clear the generation cache
cache-clean:
    go run {{MAIN}} cache clean

# Show cache status
cache-status:
    go run {{MAIN}} cache status

# Initialize a new project (provide name: just init myproject)
init name:
    go run {{MAIN}} init {{name}}

# Initialize with specific database (provide name and db: just init-db myproject postgres)
init-db name db:
    go run {{MAIN}} init {{name}} -d {{db}}

# Show version information
version:
    go run {{MAIN}} version

# ===== Examples =====
# Run the simple-api example
example-simple:
    go run examples/simple-api/main.go

# Run the todo example
example-todo:
    go run examples/todo/main.go

# Run the blog example
example-blog:
    go run examples/blog/main.go

# ===== Maintenance =====
# Remove build artifacts and cache
clean:
    rm -f {{BINARY}}
    rm -rf .apiright_cache
    go clean

# Install git hooks for pre-commit checks
setup-hooks:
    git config core.hooksPath hooks

# ===== CI =====
# Full CI pipeline: tidy, fmt, vet, lint, test, bench
ci: tidy fmt vet lint test bench

# ===== Help =====
# List all recipes
list:
    @just --list

# Interactive recipe selection
help:
    @just --choose
