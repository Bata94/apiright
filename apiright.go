// Package apiright provides a framework for converting SQLC structs to ready-to-use CRUD APIs
package apiright

import (
	"github.com/bata94/apiright/pkg/apiright"
	"github.com/bata94/apiright/pkg/crud"
	"github.com/bata94/apiright/pkg/transform"
)

// Re-export main types and functions for easier usage
type (
	App         = apiright.App
	Config      = apiright.Config
	Middleware  = apiright.Middleware
	Option      = apiright.Option
	
	CRUDConfig     = crud.Config
	CRUDHandler[T any] = crud.CRUDHandler[T]
	Transformer    = crud.Transformer
	
	ModelTransformer = transform.Transformer
	JSONTransformer  = transform.JSONTransformer
)

// App creation functions
var (
	New           = apiright.New
	DefaultConfig = apiright.DefaultConfig
	WithDatabase  = apiright.WithDatabase
	WithConfig    = apiright.WithConfig
)

// Response functions
var (
	JSONResponse    = apiright.JSONResponse
	ErrorResponse   = apiright.ErrorResponse
	SuccessResponse = apiright.SuccessResponse
)

// Middleware functions
var (
	CORSMiddleware = apiright.CORSMiddleware
	LoggingMiddleware = apiright.LoggingMiddleware
	JSONMiddleware = apiright.JSONMiddleware
	AuthMiddleware = apiright.AuthMiddleware
)

// CRUD functions - Note: Generic functions cannot be assigned to variables in Go

// Transform functions
var (
	NewTransformer = transform.NewTransformer
	TimeToString   = transform.TimeToString
	StringToTime   = transform.StringToTime
)

// Register is a generic function to register CRUD endpoints for any type
func Register[T any](app *App, config CRUDConfig) *CRUDHandler[T] {
	return crud.Register[T](app, config)
}