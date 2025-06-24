package apiright

import (
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
	"github.com/bata94/apiright/pkg/openapi"
)

type App core.App
type Ctx core.Ctx
type ApiResponse core.ApiResponse
type Middleware core.Middleware
type Handler core.Handler
type RouteOption core.RouteOption
type Route core.Route
type Router core.Router
type Logger logger.Logger
type Endpoint core.Endpoint
type RequestMethod core.RequestMethod
type Response core.Response
type RouteOptionConfig core.RouteOptionConfig

// OpenAPI types
type OpenAPIGenerator openapi.Generator
type OpenAPIConfig openapi.Config
type OpenAPISpec openapi.OpenAPISpec
type EndpointBuilder openapi.EndpointBuilder

var (
	NewApp = core.NewApp
	
	// OpenAPI functions
	NewOpenAPIGenerator = openapi.NewGenerator
	OpenAPIQuickStart   = openapi.QuickStart
	NewEndpointBuilder  = openapi.NewEndpointBuilder
)
