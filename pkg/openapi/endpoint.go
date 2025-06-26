package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// EndpointOptions contains all the options for documenting an endpoint
type EndpointOptions struct {
	// Basic information
	Summary     string
	Description string
	Tags        []string
	OperationID string
	Deprecated  bool

	// Request/Response types
	RequestType  reflect.Type
	ResponseType reflect.Type

	// Parameters
	PathParams   []ParameterInfo
	QueryParams  []ParameterInfo
	HeaderParams []ParameterInfo

	// Request body
	RequestBody *RequestBodyInfo

	// Responses
	Responses map[int]ResponseInfo

	// Security
	Security []SecurityRequirement

	// Examples
	RequestExample  interface{}
	ResponseExample interface{}
}

// ParameterInfo contains information about a parameter
type ParameterInfo struct {
	Name        string
	Description string
	Required    bool
	Type        reflect.Type
	Example     interface{}
	Schema      *Schema
}

// RequestBodyInfo contains information about request body
type RequestBodyInfo struct {
	Description string
	Required    bool
	ContentType string
	Type        reflect.Type
	Example     interface{}
	Schema      *Schema
}

// ResponseInfo contains information about a response
type ResponseInfo struct {
	Description string
	ContentType string
	Type        reflect.Type
	Example     interface{}
	Schema      *Schema
	Headers     map[string]HeaderInfo
}

// HeaderInfo contains information about a response header
type HeaderInfo struct {
	Description string
	Type        reflect.Type
	Example     interface{}
	Schema      *Schema
}

// EndpointBuilder helps build endpoint documentation
type EndpointBuilder struct {
	options EndpointOptions
	sg      *SchemaGenerator
}

// NewEndpointBuilder creates a new endpoint builder
func NewEndpointBuilder() *EndpointBuilder {
	return &EndpointBuilder{
		options: EndpointOptions{
			Responses: make(map[int]ResponseInfo),
		},
		sg: NewSchemaGenerator(),
	}
}

// Summary sets the endpoint summary
func (eb *EndpointBuilder) Summary(summary string) *EndpointBuilder {
	eb.options.Summary = summary
	return eb
}

// Description sets the endpoint description
func (eb *EndpointBuilder) Description(description string) *EndpointBuilder {
	eb.options.Description = description
	return eb
}

// Tags sets the endpoint tags
func (eb *EndpointBuilder) Tags(tags ...string) *EndpointBuilder {
	eb.options.Tags = tags
	return eb
}

// OperationID sets the operation ID
func (eb *EndpointBuilder) OperationID(id string) *EndpointBuilder {
	eb.options.OperationID = id
	return eb
}

// Deprecated marks the endpoint as deprecated
func (eb *EndpointBuilder) Deprecated() *EndpointBuilder {
	eb.options.Deprecated = true
	return eb
}

// RequestType sets the request body type
func (eb *EndpointBuilder) RequestType(t reflect.Type) *EndpointBuilder {
	eb.options.RequestType = t
	return eb
}

// ResponseType sets the response type
func (eb *EndpointBuilder) ResponseType(t reflect.Type) *EndpointBuilder {
	eb.options.ResponseType = t
	return eb
}

// PathParam adds a path parameter
func (eb *EndpointBuilder) PathParam(name, description string, t reflect.Type) *EndpointBuilder {
	eb.options.PathParams = append(eb.options.PathParams, ParameterInfo{
		Name:        name,
		Description: description,
		Required:    true,
		Type:        t,
	})
	return eb
}

// QueryParam adds a query parameter
func (eb *EndpointBuilder) QueryParam(name, description string, required bool, t reflect.Type) *EndpointBuilder {
	eb.options.QueryParams = append(eb.options.QueryParams, ParameterInfo{
		Name:        name,
		Description: description,
		Required:    required,
		Type:        t,
	})
	return eb
}

// HeaderParam adds a header parameter
func (eb *EndpointBuilder) HeaderParam(name, description string, required bool, t reflect.Type) *EndpointBuilder {
	eb.options.HeaderParams = append(eb.options.HeaderParams, ParameterInfo{
		Name:        name,
		Description: description,
		Required:    required,
		Type:        t,
	})
	return eb
}

// RequestBody sets the request body information
func (eb *EndpointBuilder) RequestBody(description string, required bool, contentType string, t reflect.Type) *EndpointBuilder {
	eb.options.RequestBody = &RequestBodyInfo{
		Description: description,
		Required:    required,
		ContentType: contentType,
		Type:        t,
	}
	return eb
}

// Response adds a response definition
func (eb *EndpointBuilder) Response(statusCode int, description, contentType string, t reflect.Type) *EndpointBuilder {
	eb.options.Responses[statusCode] = ResponseInfo{
		Description: description,
		ContentType: contentType,
		Type:        t,
		Headers:     make(map[string]HeaderInfo),
	}
	return eb
}

// ResponseHeader adds a header to a response
func (eb *EndpointBuilder) ResponseHeader(statusCode int, name, description string, t reflect.Type) *EndpointBuilder {
	if response, exists := eb.options.Responses[statusCode]; exists {
		response.Headers[name] = HeaderInfo{
			Description: description,
			Type:        t,
		}
		eb.options.Responses[statusCode] = response
	}
	return eb
}

// Security adds security requirements
func (eb *EndpointBuilder) Security(requirements ...SecurityRequirement) *EndpointBuilder {
	eb.options.Security = append(eb.options.Security, requirements...)
	return eb
}

// RequestExample sets the request example
func (eb *EndpointBuilder) RequestExample(example interface{}) *EndpointBuilder {
	eb.options.RequestExample = example
	return eb
}

// ResponseExample sets the response example
func (eb *EndpointBuilder) ResponseExample(example interface{}) *EndpointBuilder {
	eb.options.ResponseExample = example
	return eb
}

// Build builds the endpoint options
func (eb *EndpointBuilder) Build() EndpointOptions {
	return eb.options
}

// ConvertToOperation converts endpoint options to an OpenAPI operation
func (eb *EndpointBuilder) ConvertToOperation() Operation {
	operation := Operation{
		Summary:     eb.options.Summary,
		Description: eb.options.Description,
		Tags:        eb.options.Tags,
		OperationID: eb.options.OperationID,
		Deprecated:  eb.options.Deprecated,
		Security:    eb.options.Security,
		Parameters:  []Parameter{},
		Responses:   make(map[string]Response),
	}

	// Add parameters
	for _, param := range eb.options.PathParams {
		operation.Parameters = append(operation.Parameters, eb.convertParameter(param, "path"))
	}
	for _, param := range eb.options.QueryParams {
		operation.Parameters = append(operation.Parameters, eb.convertParameter(param, "query"))
	}
	for _, param := range eb.options.HeaderParams {
		operation.Parameters = append(operation.Parameters, eb.convertParameter(param, "header"))
	}

	// Add request body
	if eb.options.RequestBody != nil {
		operation.RequestBody = eb.convertRequestBody(*eb.options.RequestBody)
	}

	// Add responses
	for statusCode, response := range eb.options.Responses {
		operation.Responses[fmt.Sprintf("%d", statusCode)] = eb.convertResponse(response)
	}

	// Add default responses if none specified
	if len(operation.Responses) == 0 {
		operation.Responses["200"] = Response{
			Description: "Successful response",
		}
	}

	return operation
}

func (eb *EndpointBuilder) convertParameter(param ParameterInfo, in string) Parameter {
	var schema *Schema
	if param.Schema != nil {
		schema = param.Schema
	} else if param.Type != nil {
		generatedSchema := eb.sg.GenerateSchema(param.Type)
		schema = &generatedSchema
	}

	return Parameter{
		Name:        param.Name,
		In:          in,
		Description: param.Description,
		Required:    param.Required,
		Schema:      schema,
		Example:     param.Example,
	}
}

func (eb *EndpointBuilder) convertRequestBody(reqBody RequestBodyInfo) *RequestBody {
	contentType := reqBody.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	var schema *Schema
	if reqBody.Schema != nil {
		schema = reqBody.Schema
	} else if reqBody.Type != nil {
		generatedSchema := eb.sg.GenerateSchema(reqBody.Type)
		schema = &generatedSchema
	}

	content := make(map[string]MediaType)
	if schema != nil {
		mediaType := MediaType{
			Schema: schema,
		}
		if reqBody.Example != nil {
			mediaType.Example = reqBody.Example
		}
		content[contentType] = mediaType
	}

	return &RequestBody{
		Description: reqBody.Description,
		Required:    reqBody.Required,
		Content:     content,
	}
}

func (eb *EndpointBuilder) convertResponse(resp ResponseInfo) Response {
	response := Response{
		Description: resp.Description,
		Headers:     make(map[string]Header),
	}

	// Add headers
	for name, header := range resp.Headers {
		var schema *Schema
		if header.Schema != nil {
			schema = header.Schema
		} else if header.Type != nil {
			generatedSchema := eb.sg.GenerateSchema(header.Type)
			schema = &generatedSchema
		}

		response.Headers[name] = Header{
			Description: header.Description,
			Schema:      schema,
			Example:     header.Example,
		}
	}

	// Add content
	if resp.Type != nil || resp.Schema != nil {
		contentType := resp.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		var schema *Schema
		if resp.Schema != nil {
			schema = resp.Schema
		} else {
			generatedSchema := eb.sg.GenerateSchema(resp.Type)
			schema = &generatedSchema
		}

		mediaType := MediaType{
			Schema: schema,
		}
		if resp.Example != nil {
			mediaType.Example = resp.Example
		}

		response.Content = map[string]MediaType{
			contentType: mediaType,
		}
	}

	return response
}

// Common endpoint builders for standard HTTP responses

// StandardResponses returns common HTTP response definitions
func StandardResponses() map[int]ResponseInfo {
	return map[int]ResponseInfo{
		200: {
			Description: "Successful response",
			ContentType: "application/json",
		},
		400: {
			Description: "Bad request",
			ContentType: "application/json",
		},
		401: {
			Description: "Unauthorized",
			ContentType: "application/json",
		},
		403: {
			Description: "Forbidden",
			ContentType: "application/json",
		},
		404: {
			Description: "Not found",
			ContentType: "application/json",
		},
		500: {
			Description: "Internal server error",
			ContentType: "application/json",
		},
	}
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error" description:"Error message"`
	Code    int    `json:"code" description:"Error code"`
	Details string `json:"details,omitempty" description:"Additional error details"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message" description:"Success message"`
	Data    interface{} `json:"data,omitempty" description:"Response data"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data" description:"Response data"`
	Page       int         `json:"page" description:"Current page number"`
	PerPage    int         `json:"per_page" description:"Items per page"`
	Total      int         `json:"total" description:"Total number of items"`
	TotalPages int         `json:"total_pages" description:"Total number of pages"`
}

// HTTPMethodFromString converts a string to an HTTP method
func HTTPMethodFromString(method string) string {
	return strings.ToUpper(method)
}

// IsValidHTTPMethod checks if a method is a valid HTTP method
func IsValidHTTPMethod(method string) bool {
	validMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodTrace,
		http.MethodConnect,
	}

	method = strings.ToUpper(method)
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}
