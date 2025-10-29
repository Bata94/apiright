# Agent Guidelines for Apiright

## Project Overview
Apiright is a Go web framework in development, focusing on easy developer experience (DX) and fast prototyping. It wraps the stdlib `net/http` router to add functionality while maintaining compatibility with the standard library when needed. The goal is that no `net/http` functions or methods should be required to use this framework, though stdlib compatibility is preserved for advanced use cases.

## Build/Lint/Test Commands
- **Run all tests**: `just test` or `go test ./... -v`
- **Run single test**: `go test -v ./pkg/core -run TestNewApp`
- **Lint**: `just lint` or `golangci-lint run ./...`
- **Format**: `just fmt` or `gofmt -w -s ./...`
- **Vet**: `just check` or `go vet ./...`
- **Build CLI**: `just build-cli`
- **Build example**: `just build-example`
- **Dev mode**: `just dev`

## Project Structure
- **cmd/**: CLI commands (generate, root)
- **pkg/core/**: Main framework code (app, router, middleware, handlers)
- **pkg/auth/**: Authentication modules (JWT)
- **pkg/logger/**: Logging utilities
- **pkg/templ/**: Template rendering and code generation
- **example/**: Example application demonstrating framework usage
- **apiright.go**: CLI entry point

## Key Dependencies
- **github.com/a-h/templ**: HTML templating
- **github.com/golang-jwt/jwt/v5**: JWT authentication
- **github.com/spf13/cobra**: CLI framework
- **github.com/spf13/viper**: Configuration management

## Environment Setup
- **Nix Flakes**: Use `nix develop` or `direnv allow` (with .envrc) for development environment
- **Go Version**: 1.24.3 (managed by flake.nix)
- **Tools**: just, air, golangci-lint, templ, tailwindcss, vegeta, etc.

## Environment Variables
- **ENV=dev**: Enables debug logging and trace level logging
- **ENV=DEV**: Alternative for CLI debug logging

## CLI Usage
- **Generate Routes**: `apiright generate -i ./ui/pages -o ./uirouter/routes_gen.go`
- **Help**: `apiright --help` or `apiright generate --help`

## Development Workflow
- **Code Generation**: Use `apiright generate` to create routes from .templ files
- **Live Reload**: Use `just dev` for development with hot reloading
- **Testing**: Run `just test` for comprehensive testing
- **Linting**: Run `just lint` before commits
- **Checkup**: Run `just pre-release` before telling the user, that you are done with the task
- **Committing**: Commit with a descriptive message
- **Pull Requests**: Create a pull request with a descriptive title and reference the issue number (if applicable)
- **Pushing**: Never push to GitHub, if your are not directly asked to do so

## Testing Best Practices
- **Table-Driven Tests**: Use struct slices for test cases with httptest
- **HTTP Testing**: Use `httptest.NewRequest()` and `httptest.NewRecorder()`
- **Mock Handlers**: Create simple mock handlers for middleware testing
- **Test Coverage**: Aim for comprehensive coverage of core functionality

## File Conventions
- **Generated Files**: Use `_gen.go` suffix (e.g., `routes_gen.go`)
- **Test Files**: Standard Go `_test.go` naming
- **Template Files**: Use `.templ` extension for HTML templates
- **Package Names**: Use descriptive names (core, auth, logger, templ)

## Common Patterns
- **Functional Options**: Configuration using option functions (e.g., `AppTitle()`, `AppAddr()`)
- **Middleware Chain**: Request processing through middleware stack
- **Table-Driven Tests**: Test cases defined as struct slices
- **Context Usage**: Request context for session data and cancellation
- **Error Wrapping**: Use `fmt.Errorf` with `%w` for error chaining

## Code Style Guidelines
- **Go version**: 1.24.3
- **Imports**: Standard library first, third-party packages, then local packages
- **Naming**: PascalCase for exported, camelCase for unexported
- **Error handling**: Use `fmt.Errorf` with `%w` for wrapping
- **Types**: Use struct tags for JSON/XML/YAML (`json:"field" xml:"field"`)
- **Functions**: Use functional options pattern for configuration
- **Comments**: Use `//` for single-line comments
- **Formatting**: Run `go fmt` and `gofmt -s` before committing
- **Testing**: Use table-driven tests with struct slices and httptest
- **Logging**: Use structured logging with the logger package
