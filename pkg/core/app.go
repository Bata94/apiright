package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/bata94/apiright/pkg/logger"
	"github.com/bata94/apiright/pkg/openapi"
)

var (
	log logger.Logger
)

type AppConfig struct {
	title, serviceDescribtion, version, host, port string
	contact struct{
		Name, Email, URL string
	}
	license struct{
		Name, URL string
	}
	servers []struct{
		URL, Description string
	}
	logger  logger.Logger
}

type AppOption func(*AppConfig)

func AppTitle(title string) AppOption {
  return func(c *AppConfig) {
    c.title = title
  }
}

func AppDescription(description string) AppOption {
  return func(c *AppConfig) {
    c.serviceDescribtion = description
  }
}

func AppVersion(version string) AppOption {
  return func(c *AppConfig) {
    c.version = version
  }
}

func AppAddr(host, port string) AppOption {
  return func(c *AppConfig) {
    c.host = host
    c.port = port
  }
}

func AppLogger(logger logger.Logger) AppOption {
  return func(c *AppConfig) {
    c.logger = logger
  }
}

func AppContact(name, email, url string) AppOption {
  return func(c *AppConfig) {
    c.contact.Name = name
    c.contact.Email = email
    c.contact.URL = url
  }
}

func AppLicense(name, url string) AppOption {
  return func(c *AppConfig) {
    c.license.Name = name
    c.license.URL = url
  }
}

func AppAddServer(url, description string) AppOption {
  return func(c *AppConfig) {
    c.servers = append(c.servers, struct{
      URL, Description string
    }{url, description})
  }
}

func (c AppConfig) GetListenAddress() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

func NewApp(opts ...AppOption) App {
	handler := http.NewServeMux()

	defaultLogger := logger.NewDefaultLogger()
	defaultLogger.SetLevel(logger.TraceLevel)

	config := AppConfig{
		host: "127.0.0.1",
		port: "5500",
		logger: defaultLogger,
		title: "My App",
		serviceDescribtion: "My App Description",
		version: "0.0.0",
	}

	for _, opt := range opts {
    opt(&config)
  }

	log = config.logger

	// Setup OpenApi Builder
	openapiGenerator := openapi.QuickStart(
		config.title,
		config.serviceDescribtion,
		config.version,
	)

	if config.contact.Email != "" || config.contact.Name != "" || config.contact.URL != "" {
    openapiGenerator.GetSpec().Info.Contact = &openapi.Contact{
    	Name:  config.contact.Name,
    	URL:  config.contact.URL,
    	Email:config.contact.Email,
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

	// Return App Object
	return App{
		config:          &config,
		handler:         handler,
		Logger:          config.logger,
		router:          newRouter(""),
		defRouteHandler: defCatchallHandler,
		openapiGenerator: openapiGenerator,
	}
}

type App struct {
	config  *AppConfig
	handler *http.ServeMux
	Logger  logger.Logger

	defRouteHandler Handler
	router          *Router

	openapiGenerator *openapi.Generator
}

func (a *App) SetDefaultRoute(handler Handler) {
	a.defRouteHandler = handler
}

func (a *App) SetLogger(logger logger.Logger) {
	a.Logger = logger
}

func (a *App) NewRouter(path string) *Router {
	// Creates and adds a new Router, with a BasePath
	newRouter := newRouter(path)

	a.router.groups = append(a.router.groups, newRouter)

	return newRouter
}

func (a App) getHttpHandler() *http.ServeMux {
	return a.handler
}

func (a *App) GET(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_GET, path, handler, opt...)
}

func (a App) POST(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_POST, path, handler, opt...)
}

func (a App) PUT(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_PUT, path, handler, opt...)
}

func (a App) DELETE(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_DELETE, path, handler, opt...)
}

func (a App) OPTIONS(path string, handler Handler, opt ...RouteOption) {
	a.router.addEndpoint(METHOD_OPTIONS, path, handler, opt...)
}

// TODO: Refactor
// TODO: Add XML and YAML support, based on Request Header
func (a *App) handleFunc(route Route, endPoint Endpoint, router Router) {
	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("Registering route: %s", handlerPath)

	a.getHttpHandler().HandleFunc(handlerPath, func(w http.ResponseWriter, r *http.Request) {
		h := endPoint.handleFunc
		var err error

		if route.basePath == "/" && r.URL.Path != router.GetBasePath() {
			a.Logger.Debugf("Using default route handler for path: %s", r.URL.Path)
			h = a.defRouteHandler
		}

				panicMiddleware := PanicMiddleware()
				logMiddleware := LogMiddleware(a.Logger)

				// Create CORS config with permissive settings for quick integration
				corsConfig := CORSConfig{
					AllowOrigins:     []string{"*"},
					AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
					AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
					ExposeHeaders:    []string{"Content-Length", "Content-Type"},
					AllowCredentials: true,
					MaxAge:           86400,
				}
				corsMiddleware := CORSMiddleware(corsConfig)

				// Apply middlewares in the correct order
				// CORS should be first to handle preflight requests properly
				h = logMiddleware(h)
				h = panicMiddleware(h)
				h = corsMiddleware(h)

		c := NewCtx(w, r)

		if endPoint.routeOptionConfig.ObjIn != nil {
			c.ObjIn = endPoint.routeOptionConfig.ObjIn
			c.ObjInType = reflect.TypeOf(c.ObjIn)

			objInByte, err := io.ReadAll(r.Body)
			defer r.Body.Close()
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

		err = h(c)
		if err != nil { goto ClosingFunc }

		if endPoint.routeOptionConfig.ObjOut != nil {
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
		c.Response.SendingReturn(w, c, err)
	})
}

func (a App) addFuncToOpenApiGen(gen *openapi.Generator, route Route, endPoint Endpoint, router Router) {
	if endPoint.method == METHOD_OPTIONS { return }

	handlerPath := fmt.Sprintf("%s %s", endPoint.method.toPathString(), route.path)
	a.Logger.Debugf("OpenApiGen adding function for path: %s", handlerPath)

	objInType := reflect.TypeOf(endPoint.routeOptionConfig.ObjIn)
	objOutType := reflect.TypeOf(endPoint.routeOptionConfig.ObjOut)

	newEndpointBuilder := openapi.NewEndpointBuilder().
		Summary("Default Summary").
		Description("Default Description").
		// Tags("users", "bulk").
		RequestType(objInType).
		// RequestExample([]UpdateUserRequest{
		// 	{
		// 		FirstName: openapi.StringPtr("John"),
		// 		LastName:  openapi.StringPtr("Doe"),
		// 	},
		// 	{
		// 		Avatar: openapi.StringPtr("https://example.com/new-avatar.jpg"),
		// 	},
		// }).
		// Security(openapi.SecurityRequirement{"BearerAuth": []string{}}).
		Response(200, "Success", "application/json", objOutType).
		Response(400, "Invalid request data", "application/json", reflect.TypeOf(ErrorResponse{})).
		Response(422, "Validation errors in request data", "application/json", reflect.TypeOf(ErrorResponse{})).
		Response(500, "Internal server error", "application/json", reflect.TypeOf(ErrorResponse{}))

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
			a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
			a.handleFunc(*r, e, *a.router)
		}
	}

	for _, group := range a.router.groups {
		a.Logger.Infof("Router group with %d routes", len(group.routes))
		for _, r := range group.routes {
			for _, e := range r.endpoints {
				a.addFuncToOpenApiGen(a.openapiGenerator, *r, e, *a.router)
				a.handleFunc(*r, e, *group)
			}
		}
	}
}

func (a *App) Run() error {
	a.addRoutesToHandler()

	// Generate the documentation
	log.Info("\n📝 Generating documentation...")

	writer := openapi.NewWriter(a.openapiGenerator)
	if err := writer.WriteFiles(); err != nil {
		log.Fatal("Failed to generate documentation:", err)
	}
	log.Info("✅ Documentation files generated")

	// Generate markdown documentation
	if err := writer.GenerateMarkdownDocs(); err != nil {
		log.Fatal("Failed to generate markdown documentation:", err)
	}
	log.Info("✅ Markdown documentation generated")

	// Get and display statistics
	stats := a.openapiGenerator.GetStatistics()
	log.Info("📊 Documentation Statistics:")
	log.Infof("   Total endpoints: %d\n", stats.TotalEndpoints)
	log.Infof("   Total schemas: %d\n", stats.TotalSchemas)
	log.Infof("   Endpoints by method:\n")
	for method, count := range stats.EndpointsByMethod {
		log.Infof("     %s: %d\n", method, count)
	}
	log.Infof("   Endpoints by tag:\n")
	for tag, count := range stats.EndpointsByTag {
		log.Infof("     %s: %d\n", tag, count)
	}

	// List all generated files
	files := writer.GetGeneratedFiles()
	log.Info("📁 Generated files:")
	for _, file := range files {
		log.Infof("   %s\n", file)
	}

	// List all endpoints
	endpoints := a.openapiGenerator.ListEndpoints()
	log.Info("🔗 Generated endpoints:")
	for path, methods := range endpoints {
		log.Infof("   %s: %v\n", path, methods)
	}

	for _, f := range files {
		p := f

		if p == "docs/index.html" {
			p = "docs/"
		}
		a.getHttpHandler().HandleFunc(fmt.Sprintf("GET /%s", p), func(w http.ResponseWriter, r *http.Request) {

			content, err := os.ReadFile(f)
			if err != nil {
				if os.IsNotExist(err) {
					w.WriteHeader(404)
					w.Write([]byte("File not found"))
					return
				} else {
					w.WriteHeader(500)
					w.Write([]byte("File not readable"))
					return
				}
			}

			switch strings.Split(f, ".")[1] {
				case "html":
					w.Header().Add("Content-Type", "text/html; charset=utf-8")
				case "json":
					w.Header().Add("Content-Type", "application/json")
				case "yml":
					fallthrough
				case "yaml":
					w.Header().Add("Content-Type", "application/yaml")
			}
			w.WriteHeader(200)
			w.Write(content)
		})
	}

	return http.ListenAndServe(a.config.GetListenAddress(), a.getHttpHandler())
}
