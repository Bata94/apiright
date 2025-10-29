package core

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/bata94/apiright/pkg/logger"
	"github.com/bata94/gogen-openapi"
)

// log is the global logger instance for the core package
var log logger.Logger = logger.GetLogger()

// AppConfig holds the configuration for the application.
type AppConfig struct {
	title, serviceDescribtion, version, host, port string

	contact struct {
		Name, Email, URL string
	}
	license struct {
		Name, URL string
	}
	servers []struct {
		URL, Description string
	}
	logger  logger.Logger
	timeout time.Duration
}

// AppOption is a function that configures an AppConfig.
type AppOption func(*AppConfig)

// AppTitle sets the title of the application.
func AppTitle(title string) AppOption {
	return func(c *AppConfig) {
		c.title = title
	}
}

// AppDescription sets the description of the application.
func AppDescription(description string) AppOption {
	return func(c *AppConfig) {
		c.serviceDescribtion = description
	}
}

// AppVersion sets the version of the application.
func AppVersion(version string) AppOption {
	return func(c *AppConfig) {
		c.version = version
	}
}

// AppAddr sets the host and port the application will listen on.
func AppAddr(host, port string) AppOption {
	return func(c *AppConfig) {
		c.host = host
		c.port = port
	}
}

// AppLogger sets the logger for the application.
func AppLogger(logger logger.Logger) AppOption {
	return func(c *AppConfig) {
		c.logger = logger
	}
}

// AppContact sets the contact information for the application.
func AppContact(name, email, url string) AppOption {
	return func(c *AppConfig) {
		c.contact.Name = name
		c.contact.Email = email
		c.contact.URL = url
	}
}

// AppLicense sets the license information for the application.
func AppLicense(name, url string) AppOption {
	return func(c *AppConfig) {
		c.license.Name = name
		c.license.URL = url
	}
}

// AppAddServer adds a server to the list of servers for the application.
func AppAddServer(url, description string) AppOption {
	return func(c *AppConfig) {
		c.servers = append(c.servers, struct {
			URL, Description string
		}{url, description})
	}
}

// AppTimeout sets the timeout for the application.
func AppTimeout(timeout time.Duration) AppOption {
	return func(c *AppConfig) {
		c.timeout = timeout
	}
}

// GetListenAddress returns the address the application will listen on.
func (c AppConfig) GetListenAddress() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

// NewApp creates a new App instance.
func NewApp(opts ...AppOption) App {
	mainHandler := http.NewServeMux()

	defaultLogger := logger.NewDefaultLogger()

	if os.Getenv("ENV") == "dev" {
		defaultLogger.SetLevel(logger.TraceLevel)
		defaultLogger.Debug("development environment detected", "env", "dev", "log_level", "trace")
	}

	config := AppConfig{
		host:               "127.0.0.1",
		port:               "5500",
		logger:             defaultLogger,
		title:              "My App",
		serviceDescribtion: "My App Description",
		version:            "0.0.0",
	}

	for _, opt := range opts {
		opt(&config)
	}

	// Set the global logger for the core package and the logger package
	log = config.logger
	logger.SetLogger(config.logger)

	// Setup OpenApi Builder
	openapiGenerator := openapi.NewBasicGenerator(
		config.title,
		config.serviceDescribtion,
		config.version,
	)

	if config.contact.Email != "" || config.contact.Name != "" || config.contact.URL != "" {
		openapiGenerator.GetSpec().Info.Contact = &openapi.Contact{
			Name:  config.contact.Name,
			URL:   config.contact.URL,
			Email: config.contact.Email,
		}
	}

	if config.license.Name != "" || config.license.URL != "" {
		openapiGenerator.GetSpec().Info.License = &openapi.License{
			Name: config.license.Name,
			URL:  config.license.URL,
		}
	}

	if len(config.servers) == 0 {
		host := config.host
		port := config.port

		if host == "127.0.0.1" {
			host = "localhost"
		}

		openapiGenerator.AddServer(openapi.Server{
			URL:         fmt.Sprintf("http://%s:%s", host, port),
			Description: "Local DevServer",
		})
	} else {
		for _, server := range config.servers {
			openapiGenerator.AddServer(openapi.Server{
				URL:         server.URL,
				Description: server.Description,
			})
		}
	}

	openapiGenerator.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	})

	// Initialize timeout configuration
	timeoutConfig := DefaultTimeoutConfig()
	if config.timeout > 0 {
		timeoutConfig.Timeout = config.timeout
	}

	// Return App Object
	return App{
		config:           &config,
		handler:          mainHandler,
		Logger:           config.logger,
		router:           newRouter(""),
		defRouteHandler:  defCatchallHandler,
		openapiGenerator: openapiGenerator,
		timeoutConfig:    timeoutConfig,
		registeredRoutes: make(map[string][]string),
	}
}

// App is the main application struct.
// TODO: Implement a middleware to limit the maximum number of concurrent connections to prevent server overload.
// TODO: Implement a middleware for rate limiting to control the number of requests a client can make within a specific time frame.
// TODO: Implement a middleware to compress HTTP responses (e.g., using gzip or brotli) to reduce bandwidth usage.
// TODO: Implement a middleware for response caching to store and serve frequently requested responses, improving performance.
// TODO: Enhance the Middleware interface to include a mechanism (e.g., a 'Next' or 'Skip' function) allowing middlewares to conditionally bypass subsequent middlewares or handlers based on certain criteria (e.g., specific endpoints or request states).
type App struct {
	config  *AppConfig
	handler *http.ServeMux
	Logger  logger.Logger

	defRouteHandler Handler
	router          *Router

	openapiGenerator *openapi.Generator

	// Timeout configuration for request handling
	timeoutConfig TimeoutConfig

	registeredRoutes map[string][]string
}

// SetDefaultRoute sets the default route handler for the application.
func (a *App) SetDefaultRoute(handler Handler) {
	a.defRouteHandler = handler
}

// SetTimeout configures the request timeout for the application
func (a *App) SetTimeout(timeout time.Duration) {
	a.timeoutConfig.Timeout = timeout
}

// SetTimeoutConfig configures the complete timeout configuration for the application
func (a *App) SetTimeoutConfig(config TimeoutConfig) {
	a.timeoutConfig = config
}

// SetLogger sets the logger for the application.
func (a *App) SetLogger(logger logger.Logger) {
	a.Logger = logger
}

// TODO: Add RoutOptions
// NewRouter creates a new router with a given base path.
func (a *App) NewRouter(path string) *Router {
	// Creates and adds a new Router, with a BasePath
	newRouter := newRouter(path)
	newRouter.middlewares = a.router.middlewares

	a.router.groups = append(a.router.groups, newRouter)

	return newRouter
}

func (a *App) Use(m Middleware) {
	a.router.Use(m)
}

func (a *App) GET(path string, handler Handler, opt ...RouteOption) {
	a.router.GET(path, handler, opt...)
}

func (a *App) POST(path string, handler Handler, opt ...RouteOption) {
	a.router.POST(path, handler, opt...)
}

func (a *App) PUT(path string, handler Handler, opt ...RouteOption) {
	a.router.PUT(path, handler, opt...)
}

func (a *App) DELETE(path string, handler Handler, opt ...RouteOption) {
	a.router.DELETE(path, handler, opt...)
}

func (a *App) OPTIONS(path string, handler Handler, opt ...RouteOption) {
	a.router.OPTIONS(path, handler, opt...)
}

func (a *App) Redirect(path, url string, code int) {
	a.router.Redirect(path, url, code)
}

func (a *App) ServeStaticFile(urlPath, filePath string, opt ...StaticServFileOption) {
	a.router.ServeStaticFile(urlPath, filePath, opt...)
}

func (a *App) ServeStaticDir(urlPath, dirPath string, opt ...StaticServFileOption) {
	a.router.ServeStaticDir(urlPath, dirPath, a, opt...)
}

// TODO: Abstract ObjIn and Out Marshalling more, so that the marshalling interfaces can be set in App/Router config and exchanged by user if default isn't right
func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.ToPathString(), route.path)

	// Check if the route is already registered
	if methods, ok := a.registeredRoutes[route.path]; ok {
		if slices.Contains(methods, endPoint.method.ToPathString()) {
			log.Warn("route already registered", "path", route.path, "method", endPoint.method.ToPathString())
			return
		}
	}

	h := endPoint.handleFunc
	if len(router.middlewares) > 0 {
		for _, m := range router.middlewares {
			h = m(h)
		}
	}
	if len(endPoint.middlewares) > 0 {
		for _, m := range endPoint.middlewares {
			h = m(h)
		}
	}

	a.handler.HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		var err error

		currentHandler := h
		log.Debug("handling request", "method", r.Method, "url", r.URL.String(), "route_base", route.basePath, "router_base", router.GetBasePath())
		if r.URL.String() != "/" && (route.basePath == "/" && r.URL.Path != router.GetBasePath()) {
			currentHandler = a.defRouteHandler
		}

		c := NewCtx(w, r, route, endPoint)
		r = r.WithContext(c.Request.Context())
		// TODO: Integrate the ObjIn (object input) processing as a middleware to ensure it's part of the request handling chain.

		// TODO: Modify the input object validation to collect and return all type-related errors in a single response, instead of only the first encountered error.
		if endPoint.routeOptionConfig.ObjIn != nil {
			c.ObjIn = endPoint.routeOptionConfig.ObjIn
			c.ObjInType = reflect.TypeOf(c.ObjIn)

			contentTypeHeader := strings.TrimSpace(r.Header.Get("Content-Type"))
			var objInFunc func() error

			switch contentTypeHeader {
			case MIMETYPE_JSON.ToString():
				objInFunc = c.objInJson
			case MIMETYPE_XML.ToString():
				objInFunc = c.objInXml
			case MIMETYPE_YAML.ToString():
				objInFunc = c.objInYaml
			default:
				// TODO: Implement a MIME sniffer (e.g., using http.DetectContentType) to automatically determine the content type of the request body when the 'Content-Type' header is missing or generic.
				// http.DetectContentType(c.Request.B)
				c.Response.SetStatus(http.StatusUnsupportedMediaType)
				c.Response.SetMessage("This Content-Type isn't supported... yet... If u really need it, reach out.")
				goto ClosingFunc
			}

			if err = objInFunc(); err != nil {
				goto ClosingFunc
			}
			if !c.validateObjInType() {
				err = errors.New("input: parsed Object != wanted Object")
				goto ClosingFunc
			}
		}

		if err = currentHandler(c); err != nil {
			goto ClosingFunc
		}

		if endPoint.routeOptionConfig.ObjOut != nil {
			c.ObjOutType = reflect.TypeOf(endPoint.routeOptionConfig.ObjOut)
			if !c.validateObjOutType() {
				err = errors.New("output: parsed Object != wanted Object")
				goto ClosingFunc
			}

			// TODO: Add default MIMEType, when no Accept header is provided.
			var objOutFunc func() error
			acceptHeader := c.Request.Header.Get("accept")
			log.Debug("processing accept header", "accept", acceptHeader)
			acceptHeaders := strings.Split(acceptHeader, ",")

			for _, acceptHeader := range acceptHeaders {
				acceptHeader = strings.TrimSpace(acceptHeader)
				switch acceptHeader {
				case MIMETYPE_JSON.ToString():
					log.Debug("encoding response as JSON")
					objOutFunc = c.objOutJson
				case MIMETYPE_XML.ToString():
					log.Debug("encoding response as XML")
					objOutFunc = c.objOutXML
				case MIMETYPE_YAML.ToString():
					log.Debug("encoding response as YAML")
					objOutFunc = c.objOutYaml
				}
				if objOutFunc != nil {
					break
				}
			}

			if objOutFunc == nil {
				log.Error("unsupported MIME types requested", "accept", acceptHeaders)
				c.Response.SetStatus(http.StatusNotAcceptable)
				c.Response.SetMessage("The requested MIME Type isn't supported... yet... If u really need it, reach out.")
				goto ClosingFunc
			}

			if err = objOutFunc(); err != nil {
				goto ClosingFunc
			}
		}

	ClosingFunc:
		c.SendingReturn(w, err)
	})

	a.registeredRoutes[route.path] = append(a.registeredRoutes[route.path], endPoint.method.ToPathString())
}

func (a *App) addFuncToOpenApiGen(gen *openapi.Generator, route Route, endPoint Endpoint, _ Router) {
	if endPoint.method == METHOD_OPTIONS || !endPoint.routeOptionConfig.openApiEnabled {
		return
	}

	handlerPath := fmt.Sprintf("%s %s", endPoint.method.ToPathString(), route.path)
	a.Logger.Debugf("OpenApiGen adding function for path: %s", handlerPath)

	objInType := reflect.TypeOf(endPoint.routeOptionConfig.ObjIn)
	objOutType := reflect.TypeOf(endPoint.routeOptionConfig.ObjOut)
	errResType := reflect.TypeOf(ErrorResponse{})

	// TODO: Implement logic to allow choosing between a default summary/description for OpenAPI documentation and a custom one provided by the user.
	summary := "Default Summary"
	desc := "Default Description"
	if endPoint.routeOptionConfig.openApiConfig.summary != "" {
		summary = endPoint.routeOptionConfig.openApiConfig.summary
	}
	if endPoint.routeOptionConfig.openApiConfig.description != "" {
		desc = endPoint.routeOptionConfig.openApiConfig.description
	}

	allAvailableMIME := []string{MIMETYPE_JSON.ToString(), MIMETYPE_YAML.ToString(), MIMETYPE_XML.ToString()}

	newEndpointBuilder := openapi.NewEndpointBuilder().
		Summary(summary).
		Description(desc).
		// TODO: Implement security handling for the OpenAPI endpoint, allowing definition of security schemes (e.g., API keys, OAuth2) and their application to specific endpoints.
		// Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(400, "Invalid request data", allAvailableMIME, errResType).
		Response(422, "Validation errors in request data", allAvailableMIME, errResType).
		Response(500, "Internal server error", allAvailableMIME, errResType)

	if strings.Contains(route.path, "{") {
		pNames := strings.SplitSeq(route.path, "{")
		for pName := range pNames {
			if strings.Contains(pName, "}") {
				pName = strings.TrimSuffix(pName, "}")
				newEndpointBuilder.PathParam(pName, "", reflect.TypeOf(""))
			}
		}
	}

	if len(endPoint.routeOptionConfig.queryParams) != 0 {
		for _, queryParam := range endPoint.routeOptionConfig.queryParams {
			newEndpointBuilder.QueryParam(queryParam.Name, queryParam.Description, queryParam.Required, queryParam.Type)
		}
	}

	if len(endPoint.routeOptionConfig.openApiConfig.tags) != 0 {
		newEndpointBuilder.Tags(endPoint.routeOptionConfig.openApiConfig.tags...)
	}
	if endPoint.routeOptionConfig.openApiConfig.deprecated {
		newEndpointBuilder.Deprecated()
	}

	if endPoint.routeOptionConfig.openApiConfig.jwtAuth {
		newEndpointBuilder.Security(openapi.SecurityRequirement{"BearerAuth": []string{}})
	}

	if endPoint.routeOptionConfig.ObjIn != nil {
		// TODO: Dynamically add the specific MIME types that this endpoint expects for the request body based on configuration or reflection.
		newEndpointBuilder.RequestType(objInType)
		newEndpointBuilder.RequestBody("Request body", true, []string{MIMETYPE_JSON.ToString(), MIMETYPE_YAML.ToString(), MIMETYPE_XML.ToString()}, objInType)
		// TODO: Implement dynamic generation or inclusion of request body examples for OpenAPI documentation based on the ObjIn type.
		// RequestExample([]UpdateUserRequest{
		// 	{
		// 		FirstName: openapi.StringPtr("John"),
		// 		LastName:  openapi.StringPtr("Doe"),
		// 	},
		// 	{
		// 		Avatar: openapi.StringPtr("https://example.com/new-avatar.jpg"),
		// 	},
		// }).
	}
	if endPoint.routeOptionConfig.ObjOut != nil {
		newEndpointBuilder.Response(200, "Success", allAvailableMIME, objOutType)
	} else {
		// TODO: Refactor this reflection statement to provide a more straightforward and idiomatic way to define the success response type (e.g., for empty or simple text responses) in OpenAPI documentation when ObjOut is not specified.
		newEndpointBuilder.Response(200, "Success", []string{MIMETYPE_TEXT.ToString()}, reflect.TypeOf("i"))
	}

	if err := gen.AddEndpointWithBuilder(endPoint.method.ToPathString(), route.path, newEndpointBuilder); err != nil {
		log.Fatal("Failed to add endpoint: ", err)
	}
	a.Logger.Debug("✅ Endpoint added - ", handlerPath)
}

func (a *App) addRoutesToHandler() {
	a.Logger.Info("registering available routes", "global_routes", len(a.router.routes), "router_groups", len(a.router.groups))

	for _, r := range a.router.routes {
		for _, e := range r.endpoints {
			a.handleFunc(*r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Debug("registering router group routes", "group_routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.handleFunc(*r, e, *group)
			}
		}
	}
}

func (a *App) addRoutesToOpenApi() {
	a.Logger.Info("documenting available routes", "global_routes", len(a.router.routes), "router_groups", len(a.router.groups))

	for _, r := range a.router.routes {
		for _, e := range r.endpoints {
			a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Debug("documenting router group routes", "group_routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
			}
		}
	}
}

func (a *App) genOpenApiFiles() []string {
	// Generate the documentation
	a.Logger.Info("generating OpenAPI documentation")

	writer := openapi.NewWriter(a.openapiGenerator)
	if err := writer.WriteFiles(); err != nil {
		log.Error("failed to generate documentation", "error", err)
		os.Exit(1)
	}
	a.Logger.Info("documentation files generated")

	// Generate markdown documentation
	if err := writer.GenerateMarkdownDocs(); err != nil {
		a.Logger.Error("failed to generate markdown documentation", "error", err)
		os.Exit(1)
	}
	a.Logger.Info("markdown documentation generated")

	// Get and display statistics
	stats := a.openapiGenerator.GetStatistics()
	a.Logger.Info("documentation statistics", "total_endpoints", stats.TotalEndpoints, "total_schemas", stats.TotalSchemas)

	// List all generated files
	files := writer.GetGeneratedFiles()
	a.Logger.Info("generated files", "count", len(files), "files", files)

	// List all endpoints
	endpoints := a.openapiGenerator.ListEndpoints()
	a.Logger.Debug("generated endpoints", "endpoints", endpoints)

	return files
}

func (a *App) Run() error {
	a.addRoutesToOpenApi()
	a.genOpenApiFiles()

	a.ServeStaticDir("/docs", "docs/")
	a.addRoutesToHandler()

	log.Info("server starting", "address", a.config.GetListenAddress())
	return http.ListenAndServe(a.config.GetListenAddress(), a.handler)
}
