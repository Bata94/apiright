package core

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// Route represents a route in the application.
type Route struct {
	basePath, path string
	endpoints      []Endpoint
}

// RequestMethod is the HTTP request method.
type RequestMethod int

const (
	// METHOD_GET is the GET HTTP method.
	METHOD_GET RequestMethod = iota
	// METHOD_POST is the POST HTTP method.
	METHOD_POST
	// METHOD_PUT is the PUT HTTP method.
	METHOD_PUT
	// METHOD_PATCH is the PATCH HTTP method.
	METHOD_PATCH
	// METHOD_DELETE is the DELETE HTTP method.
	METHOD_DELETE
	// METHOD_HEAD is the HEAD HTTP method.
	METHOD_HEAD
	// METHOD_OPTIONS is the OPTIONS HTTP method.
	METHOD_OPTIONS
	// METHOD_TRACE is the TRACE HTTP method.
	METHOD_TRACE
	// METHOD_CONNECT is the CONNECT HTTP method.
	METHOD_CONNECT
)

var (
	requestMethodPathStrings = []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"HEAD",
		"OPTIONS",
		"TRACE",
		"CONNECT",
	}
)

func (m RequestMethod) toPathString() string {
	return requestMethodPathStrings[m]
}

// Endpoint represents an endpoint in the application.
type Endpoint struct {
	method            RequestMethod
	handleFunc        Handler
	middlewares       []Middleware
	routeOptionConfig RouteOptionConfig
}

// Handler is a function that handles a request.
type Handler func(*Ctx) error

type MIMEType int

const (
	MIMETYPE_JSON MIMEType = iota
	MIMETYPE_XML
	MIMETYPE_YAML
	MIMETYPE_FORM_URL
	MIMETYPE_MULTIPART_FORM_DATA
	MIMETYPE_OCTET_STREAM
	MIMETYPE_TEXT
)

var (
	mimeTypeStrings = []string{
		"application/json",
		"application/xml",
		"application/yaml",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"application/octet-stream",
		"text/plain",
	}
)

func (m MIMEType) toString() string {
	return mimeTypeStrings[m]
}

// NewCtx creates a new Ctx instance.
func NewCtx(w http.ResponseWriter, r *http.Request, route Route, ep Endpoint) *Ctx {
	c := &Ctx{
		Request:  r,
		Response: NewApiResponse(),

		PathParams:  make(map[string]string),
		QueryParams: make(map[string]string),
		Session:     make(map[string]any),

		conClosed:  make(chan bool, 1),
		conStarted: time.Now(),
	}

	rePathParams := regexp.MustCompile(`{([^}]+)}`)
	matchesPathParams := rePathParams.FindAllStringSubmatch(route.path, -1)
	for _, m := range matchesPathParams {
		if len(m) > 1 {
			c.PathParams[m[1]] = r.PathValue(m[1])
		}
	}

	queryValues := r.URL.Query()
	for k, v := range queryValues {
		c.QueryParams[k] = v[0]
	}

	return c
}

// Ctx is the context for a request.
type Ctx struct {
	// TODO: Refactor ApiResponse to an interface (e.g., ResponseWriter) to allow for different response types, such as HTML responses, beyond just API responses.
	Response *ApiResponse
	Request  *http.Request

	PathParams  map[string]string
	QueryParams map[string]string
	Session     map[string]any

	conClosed  chan (bool)
	conStarted time.Time
	conEnded   time.Time

	ObjIn     any
	ObjInType reflect.Type
	objInByte []byte

	ObjOut     any
	ObjOutType reflect.Type
}

func (c *Ctx) getObjInByte() []byte {
	if len(c.objInByte) != 0 {
		return c.objInByte
	}

	log.Debug("ObjInByte not set, reading from request body")
	b, err := io.ReadAll(c.Request.Body)
	defer func() { _ = c.Request.Body.Close() }()

	if err != nil {
		log.Fatal(err)
	}

	return b
}

func (c *Ctx) validateObjInType() bool {
	return reflect.TypeOf(c.ObjIn) == c.ObjInType
}

func (c *Ctx) objInJson() error {
	return json.Unmarshal(c.getObjInByte(), &c.ObjIn)
}

func (c *Ctx) objInXml() error {
	return xml.Unmarshal(c.getObjInByte(), &c.ObjIn)
}

func (c *Ctx) objInYaml() error {
	return yaml.Unmarshal(c.getObjInByte(), c.ObjIn)
}

func (c *Ctx) setObjOutData(b []byte, err error) error {
	if err != nil {
		return err
	}
	c.Response.Data = b
	return nil
}

func (c *Ctx) validateObjOutType() bool {
	log.Debugf("ObjOutType: %v", c.ObjOutType)
	log.Debugf("ObjOut: %v", c.ObjOut)
	return reflect.TypeOf(c.ObjOut) == c.ObjOutType
}

func (c *Ctx) objOutJson() error {
	return c.setObjOutData(json.Marshal(c.ObjOut))
}

func (c *Ctx) objOutXML() error {
	return c.setObjOutData(xml.Marshal(c.ObjOut))
}

func (c *Ctx) objOutYaml() error {
	return c.setObjOutData(yaml.Marshal(c.ObjOut))
}

type FileSaveOption func(*FileSaveConfig)

type FileSaveConfig struct {
	overwrite bool
	// TODO: Add more OPTIONS
	// autoTimestamp bool
	// autoFileSuffix bool
	// filename string
}

func NewFileSaveConfig(opts ...FileSaveOption) *FileSaveConfig {
	config := &FileSaveConfig{}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// WithOverwrite overwrites the file if it already exists.
func WithOverwrite() FileSaveOption {
	return func(c *FileSaveConfig) {
		c.overwrite = true
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (c *Ctx) SaveFile(formKey, dstPath string, opts ...FileSaveOption) error {
	config := NewFileSaveConfig(opts...)

	file, fileHeader, err := c.Request.FormFile(formKey)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	filename := fmt.Sprint(time.Now().Unix(), "-", filepath.Base(fileHeader.Filename))

	dstPath = filepath.Clean(dstPath)
	dstPath = filepath.Dir(dstPath)

	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return err
	}

	dstPath = filepath.Join(dstPath, filename)
	if !config.overwrite && fileExists(dstPath) {
		return errors.New("file already exists")
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	_, err = io.Copy(dst, file)
	return err
}

// Close closes the connection.
func (c *Ctx) Close() {
	log.Debug("Closing connection")
	c.conEnded = time.Now()
	c.conClosed <- true
}

// IsClosed returns true if the connection is closed.
func (c *Ctx) IsClosed() bool {
	return <-c.conClosed
}

// RouteOptionConfig holds the configuration for a route.
// TODO: Integrate RouteOptionConfig into the Router to allow setting default values for routes defined within that router, which can then be overridden by individual route options.
type RouteOptionConfig struct {
	openApiEnabled bool
	openApiConfig  struct {
		summary, description string
		tags                 []string
		deprecated           bool
		jwtAuth              bool
	}

	ObjIn       any
	ObjOut      any
	queryParams []struct {
		Name, Description string
		Required          bool
		Type              reflect.Type
		Example           any
	}
	middlewares []Middleware
}

// RouteOption is a function that configures a RouteOptionConfig.
type RouteOption func(*RouteOptionConfig)

// NewRouteOptionConfig creates a new RouteOptionConfig.
func NewRouteOptionConfig(opts ...RouteOption) *RouteOptionConfig {
	// TODO: Allow default RouteOptionConfig values to be set in AppConfig and passed through to NewRouteOptionConfig, enabling global defaults for route options.
	config := &RouteOptionConfig{
		openApiEnabled: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// Use adds a middleware to the route.
func Use(m Middleware) RouteOption {
	return func(c *RouteOptionConfig) {
		c.middlewares = append(c.middlewares, m)
	}
}

func WithQueryParam(name, description string, required bool, t reflect.Type) RouteOption {
	return func(c *RouteOptionConfig) {
		c.queryParams = append(c.queryParams, struct {
			Name, Description string
			Required          bool
			Type              reflect.Type
			Example           any
		}{
			Name:        name,
			Description: description,
			Required:    required,
			Type:        t,
		})
	}
}

// WithObjIn sets the input object for the route.
func WithObjIn(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjIn = obj
	}
}

// WithObjOut sets the output object for the route.
func WithObjOut(obj any) RouteOption {
	return func(c *RouteOptionConfig) {
		c.ObjOut = obj
	}
}

// WithOpenApiDisabled disables OpenAPI generation for the route.
func WithOpenApiDisabled() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = false
	}
}

// WithOpenApiEnabled enables OpenAPI generation for the route.
func WithOpenApiEnabled(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiEnabled = true
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

// WithOpenApiInfos sets the OpenAPI summary and description for the route.
func WithOpenApiInfos(summary, description string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.summary = summary
		c.openApiConfig.description = description
	}
}

// WithOpenApiDeprecated marks the route as deprecated in OpenAPI.
func WithOpenApiDeprecated() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.deprecated = true
	}
}

// WithOpenApiTags adds tags to the route in OpenAPI.
func WithOpenApiTags(tags ...string) RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.tags = tags
	}
}

// WithOpenApiJwtAuth enables JWT authentication for the route in OpenAPI.
func WithOpenApiJwtAuth() RouteOption {
	return func(c *RouteOptionConfig) {
		c.openApiConfig.jwtAuth = true
	}
}
