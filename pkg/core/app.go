package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/bata94/apiright/pkg/logger"
	"github.com/bata94/gogen-openapi"
)

var (
	log logger.Logger
)

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
	handler := http.NewServeMux()

	defaultLogger := logger.NewDefaultLogger()

	if os.Getenv("ENV") == "dev" {
		defaultLogger.SetLevel(logger.TraceLevel)
		defaultLogger.Debug("ENV = dev, set log level to trace")
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

	log = config.logger

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

	// Initialize timeout configuration
	timeoutConfig := DefaultTimeoutConfig()
	if config.timeout > 0 {
		timeoutConfig.Timeout = config.timeout
	}

	// Return App Object
	return App{
		config:           &config,
		handler:          handler,
		Logger:           config.logger,
		router:           newRouter(""),
		defRouteHandler:  defCatchallHandler,
		openapiGenerator: openapiGenerator,
		timeoutConfig:    timeoutConfig,
	}
}

// App is the main application struct.
// TODO: Add MaxConnection handling (Middleware??)
// TODO: Add RateLimit handling (Middleware??)
type App struct {
	config  *AppConfig
	handler *http.ServeMux
	Logger  logger.Logger

	defRouteHandler Handler
	router          *Router

	openapiGenerator *openapi.Generator

	// Timeout configuration for request handling
	timeoutConfig TimeoutConfig
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

func (a App) getHttpHandler() *http.ServeMux {
	return a.handler
}

func (a *App) GET(path string, handler Handler, opt ...RouteOption) {
	a.router.GET(path, handler, opt...)
}

func (a App) POST(path string, handler Handler, opt ...RouteOption) {
	a.router.POST(path, handler, opt...)
}

func (a App) PUT(path string, handler Handler, opt ...RouteOption) {
	a.router.PUT(path, handler, opt...)
}

func (a App) DELETE(path string, handler Handler, opt ...RouteOption) {
	a.router.DELETE(path, handler, opt...)
}

func (a App) OPTIONS(path string, handler Handler, opt ...RouteOption) {
	a.router.OPTIONS(path, handler, opt...)
}

func (a App) ServeStaticFile(urlPath, filePath string, opt ...StaticServFileOption) {
	err := a.router.ServeStaticFile(urlPath, filePath, opt...)
	if err != nil {
		panic(fmt.Errorf("error serving static file: %w", err))
	}
}

func (a App) ServeStaticDir(urlPath, dirPath string) {
	a.router.ServeStaticDir(urlPath, dirPath, a)
}

// TODO: Refactor
// TODO: Add XML and YAML support, based on Request Header
func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("Registering route: %s", handlerPath)

	h := endPoint.handleFunc
	// Add middlewares
	if len(router.middlewares) > 0 {
		log.Debug("Adding middlewares, from Router...")
		for _, m := range router.middlewares {
			h = m(h)
		}
	}

	if len(endPoint.middlewares) > 0 {
		log.Debug("Adding middlewares, from Route...")
		for _, m := range endPoint.middlewares {
			h = m(h)
		}
	}

	a.getHttpHandler().HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Handling request: ", r.URL.Path)
		var err error

		if route.basePath == "/" && r.URL.Path != router.GetBasePath() {
			a.Logger.Debugf("Using default route handler for path: %s", r.URL.Path)
			h = a.defRouteHandler
		}

		log.Debug("Setting CTX")
		c := NewCtx(w, r)
		r = r.WithContext(c.Request.Context())

		// TODO: Return all wrong types in respose, not only the first one
		if endPoint.routeOptionConfig.ObjIn != nil {
			log.Debug("Setting up ObjIn")
			c.ObjIn = endPoint.routeOptionConfig.ObjIn
			c.ObjInType = reflect.TypeOf(c.ObjIn)

			objInByte, err := io.ReadAll(r.Body)
			defer func() {
				_ = r.Body.Close()
			}()

			if err != nil {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Body could not be read, err: " + err.Error())
				goto ClosingFunc
			}

			err = json.Unmarshal(objInByte, &c.ObjIn)
			if err != nil {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Body could not be parsed to Object, err: " + err.Error())
				goto ClosingFunc
			}

			if reflect.TypeOf(c.ObjIn) != c.ObjInType {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Parsed Object != wanted Object")
				goto ClosingFunc
			}
		}

		log.Debug("Sending to handleFunc")
		err = h(c)
		if err != nil {
			goto ClosingFunc
		}

		if endPoint.routeOptionConfig.ObjOut != nil {
			log.Debug("Receiving ObjOut")
			c.ObjOutType = reflect.TypeOf(endPoint.routeOptionConfig.ObjOut)
			if reflect.TypeOf(c.ObjOut) != c.ObjOutType {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Error marshaling JSON, ObjOut != wanted ObjOut Type")
				goto ClosingFunc
			}

			c.Response.Data, err = json.Marshal(c.ObjOut)
			if err != nil {
				c.Response.SetStatus(http.StatusInternalServerError)
				c.Response.SetMessage("Error marshaling JSON: " + err.Error())
				goto ClosingFunc
			}
		}

	ClosingFunc:
		log.Debug("To Sending Return")
		c.SendingReturn(w, err)
	})
}

func (a App) addFuncToOpenApiGen(gen *openapi.Generator, route Route, endPoint Endpoint, router Router) {
	if endPoint.method == METHOD_OPTIONS || !endPoint.routeOptionConfig.openApiEnabled {
		return
	}

	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("OpenApiGen adding function for path: %s", handlerPath)

	objInType := reflect.TypeOf(endPoint.routeOptionConfig.ObjIn)
	objOutType := reflect.TypeOf(endPoint.routeOptionConfig.ObjOut)
	errResType := reflect.TypeOf(ErrorResponse{})

	// TODO: Choose between default summary and description
	summary := "Default Summary"
	desc := "Default Description"
	if endPoint.routeOptionConfig.openApiConfig.summary != "" {
		summary = endPoint.routeOptionConfig.openApiConfig.summary
	}
	if endPoint.routeOptionConfig.openApiConfig.description != "" {
		desc = endPoint.routeOptionConfig.openApiConfig.description
	}

	newEndpointBuilder := openapi.NewEndpointBuilder().
		Summary(summary).
		Description(desc).
		// TODO: Implement
		// Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(400, "Invalid request data", "application/json", errResType).
		Response(422, "Validation errors in request data", "application/json", errResType).
		Response(500, "Internal server error", "application/json", errResType)

	if len(endPoint.routeOptionConfig.openApiConfig.tags) != 0 {
		newEndpointBuilder.Tags(endPoint.routeOptionConfig.openApiConfig.tags...)
	}
	if endPoint.routeOptionConfig.openApiConfig.deprecated {
		newEndpointBuilder.Deprecated()
	}
	if endPoint.routeOptionConfig.ObjIn != nil {
		newEndpointBuilder.RequestType(objInType)
		// TODO: Implement
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
		newEndpointBuilder.Response(200, "Success", "application/json", objOutType)
	} else {
		// TODO: This reflect statement must be easier
		newEndpointBuilder.Response(200, "Success", "text/plain", reflect.TypeOf("i"))
	}

	if err := gen.AddEndpointWithBuilder(endPoint.method.toPathString(), route.path, newEndpointBuilder); err != nil {
		log.Fatal("Failed to add endpoint: ", err)
	}
	a.Logger.Debug("✅ Endpoint added - ", handlerPath)
}

func (a App) addRoutesToHandler() {
	a.Logger.Info("Registering available routes")

	a.Logger.Infof("Global Router with %d routes", len(a.router.routes))
	for _, r := range a.router.routes {
		for _, e := range r.endpoints {
			a.handleFunc(*r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Infof("Router group with %d routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.handleFunc(*r, e, *group)
			}
		}
	}
}

func (a App) addRoutesToOpenApi() {
	a.Logger.Info("Documenting available routes")

	a.Logger.Infof("Global Router with %d routes", len(a.router.routes))
	for _, r := range a.router.routes {
		for _, e := range r.endpoints {
			a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Infof("Router group with %d routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
			}
		}
	}
}

func (a *App) genOpenApiFiles() []string {
	// Generate the documentation
	a.Logger.Info("")
	a.Logger.Info("📝 Generating documentation...")

	writer := openapi.NewWriter(a.openapiGenerator)
	if err := writer.WriteFiles(); err != nil {
		log.Fatal("Failed to generate documentation:", err)
	}
	a.Logger.Info("✅ Documentation files generated")

	// Generate markdown documentation
	if err := writer.GenerateMarkdownDocs(); err != nil {
		a.Logger.Fatal("Failed to generate markdown documentation:", err)
	}
	a.Logger.Info("✅ Markdown documentation generated")

	// Get and display statistics
	stats := a.openapiGenerator.GetStatistics()
	a.Logger.Info("")
	a.Logger.Info("📊 Documentation Statistics:")
	a.Logger.Infof("   Total endpoints: %d", stats.TotalEndpoints)
	a.Logger.Infof("   Total schemas: %d", stats.TotalSchemas)
	a.Logger.Infof("   Endpoints by method:")
	for method, count := range stats.EndpointsByMethod {
		a.Logger.Infof("     %s: %d", method, count)
	}
	a.Logger.Infof("   Endpoints by tag:")
	for tag, count := range stats.EndpointsByTag {
		a.Logger.Infof("     %s: %d", tag, count)
	}

	// List all generated files
	files := writer.GetGeneratedFiles()
	a.Logger.Info("")
	a.Logger.Info("📁 Generated files:")
	for _, file := range files {
		a.Logger.Infof("   %s", file)
	}

	// List all endpoints
	endpoints := a.openapiGenerator.ListEndpoints()
	a.Logger.Info("")
	a.Logger.Info("🔗 Generated endpoints:")
	for path, methods := range endpoints {
		a.Logger.Infof("   %s: %v", path, methods)
	}
	return files
}

func (a *App) Run() error {
	a.addRoutesToOpenApi()
	a.genOpenApiFiles()

	a.addRoutesToHandler()
	a.ServeStaticDir("/docs", "docs")

	return http.ListenAndServe(a.config.GetListenAddress(), a.getHttpHandler())
}
