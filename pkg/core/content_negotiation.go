package core

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

// ContentNegotiatorImpl implements the ContentNegotiator interface
type ContentNegotiatorImpl struct {
	supportedTypes []string
	defaultType    string
}

// XMLResponse represents a generic XML response with proper structure
type XMLResponse struct {
	XMLName xml.Name `xml:"response"`
	Data    any      `xml:",any"`
}

// XMLElement represents a generic XML element that can hold any data type
type XMLElement struct {
	XMLName  xml.Name
	Value    any          `xml:",chardata"`
	Elements []XMLElement `xml:",any"`
	Attrs    []xml.Attr   `xml:",attr"`
}

// NewContentNegotiator creates a new content negotiator
func NewContentNegotiator() *ContentNegotiatorImpl {
	return &ContentNegotiatorImpl{
		supportedTypes: []string{
			"application/json",
			"application/xml",
			"application/yaml",
			"application/protobuf",
			"text/plain",
		},
		defaultType: "application/json",
	}
}

// SupportedTypes returns the list of supported content types
func (cn *ContentNegotiatorImpl) SupportedTypes() []string {
	return cn.supportedTypes
}

// SerializeResponse serializes data to the specified content type
func (cn *ContentNegotiatorImpl) SerializeResponse(data any, contentType string) ([]byte, error) {
	switch contentType {
	case "application/json":
		return json.Marshal(data)
	case "application/xml":
		return cn.serializeToXML(data)
	case "application/yaml":
		return yaml.Marshal(data)
	case "text/plain":
		return []byte(fmt.Sprintf("%v", data)), nil
	case "application/protobuf":
		return cn.serializeProtobuf(data)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// DeserializeRequest deserializes data from the specified content type
func (cn *ContentNegotiatorImpl) DeserializeRequest(data []byte, contentType string, target any) error {
	switch contentType {
	case "application/json":
		return json.Unmarshal(data, target)
	case "application/xml":
		return xml.Unmarshal(data, target)
	case "application/yaml":
		return yaml.Unmarshal(data, target)
	case "text/plain":
		// For plain text, try to convert to string and set if target is string pointer
		if strPtr, ok := target.(*string); ok {
			*strPtr = string(data)
			return nil
		}
		return fmt.Errorf("plain text deserialization only supports *string target")
	case "application/protobuf":
		return cn.deserializeProtobuf(data, target)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// DetectContentType detects content type from Accept header
func (cn *ContentNegotiatorImpl) DetectContentType(header string) string {
	if header == "" {
		return cn.defaultType
	}

	// Parse Accept header with q-value support
	type acceptOption struct {
		typeName string
		quality  float64
	}

	var options []acceptOption

	// Split by comma to handle multiple types
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Extract type and q-value
		typeAndParams := strings.SplitN(part, ";", 2)
		typeName := strings.TrimSpace(typeAndParams[0])

		// Default quality is 1.0
		quality := 1.0

		// Parse q-value if present
		if len(typeAndParams) > 1 {
			qPart := strings.TrimSpace(typeAndParams[1])
			if strings.HasPrefix(qPart, "q=") {
				qValue := strings.TrimPrefix(qPart, "q=")
				if parsedQ, err := strconv.ParseFloat(qValue, 64); err == nil {
					quality = parsedQ
				}
			}
		}

		// Skip q=0 options (not acceptable)
		if quality > 0 {
			options = append(options, acceptOption{typeName: typeName, quality: quality})
		}
	}

	// Sort by quality (highest first)
	for i := 0; i < len(options)-1; i++ {
		for j := i + 1; j < len(options); j++ {
			if options[j].quality > options[i].quality {
				options[i], options[j] = options[j], options[i]
			}
		}
	}

	// Find best match
	for _, opt := range options {
		acceptedType := opt.typeName

		// Check for wildcards
		if acceptedType == "*/*" {
			return cn.defaultType
		}

		// Check for exact match
		for _, supportedType := range cn.supportedTypes {
			if acceptedType == supportedType {
				return supportedType
			}
		}

		// Check for type wildcard (e.g., "application/*")
		if strings.HasSuffix(acceptedType, "/*") {
			prefix := strings.TrimSuffix(acceptedType, "/*")
			for _, supportedType := range cn.supportedTypes {
				if strings.HasPrefix(supportedType, prefix+"/") {
					return supportedType
				}
			}
		}
	}

	// Fallback to default
	return cn.defaultType
}

// Constants for content types
const (
	ContentTypeJSON     = "application/json"
	ContentTypeXML      = "application/xml"
	ContentTypeYAML     = "application/yaml"
	ContentTypeProtobuf = "application/protobuf"
	ContentTypeText     = "text/plain"
	ContentTypeHTML     = "text/html"
	ContentTypeForm     = "application/x-www-form-urlencoded"
)

// ContentInfo holds information about content negotiation
type ContentInfo struct {
	ContentType   string
	ContentLength int
	Charset       string
	Encoding      string
}

// ParseContentHeader parses a content-related header (Content-Type or Accept)
func ParseContentHeader(header string) *ContentInfo {
	if header == "" {
		return &ContentInfo{}
	}

	info := &ContentInfo{}
	parts := strings.Split(header, ";")

	// Main content type
	if len(parts) > 0 {
		info.ContentType = strings.TrimSpace(parts[0])
	}

	// Parameters
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "charset=") {
			info.Charset = strings.TrimPrefix(part, "charset=")
		} else if strings.HasPrefix(part, "encoding=") {
			info.Encoding = strings.TrimPrefix(part, "encoding=")
		}
	}

	return info
}

// IsContentTypeSupported checks if a content type is supported
func (cn *ContentNegotiatorImpl) IsContentTypeSupported(contentType string) bool {
	for _, supportedType := range cn.supportedTypes {
		if supportedType == contentType {
			return true
		}
	}
	return false
}

// SetSupportedTypes sets the list of supported content types
func (cn *ContentNegotiatorImpl) SetSupportedTypes(types []string) {
	cn.supportedTypes = types
}

// SetDefaultType sets the default content type
func (cn *ContentNegotiatorImpl) SetDefaultType(defaultType string) {
	cn.defaultType = defaultType
}

// serializeProtobuf serializes data to protobuf format
func (cn *ContentNegotiatorImpl) serializeProtobuf(data any) ([]byte, error) {
	// Check if data is already a proto.Message
	if protoMsg, ok := data.(proto.Message); ok {
		return proto.Marshal(protoMsg)
	}

	// Try to handle common cases where we need to convert to protobuf
	// This is a simple implementation - in a real scenario, you'd want to
	// have proper mapping between Go types and protobuf messages

	// For now, fall back to JSON if we can't handle it as protobuf
	return nil, fmt.Errorf("data is not a protobuf message and no conversion available")
}

// deserializeProtobuf deserializes data from protobuf format
func (cn *ContentNegotiatorImpl) deserializeProtobuf(data []byte, target any) error {
	// Check if target is a pointer to a proto.Message
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	// Try to cast to proto.Message
	if protoMsg, ok := target.(proto.Message); ok {
		return proto.Unmarshal(data, protoMsg)
	}

	return fmt.Errorf("target is not a protobuf message pointer")
}

// serializeToXML converts map[string]any to XML-serializable format
func (cn *ContentNegotiatorImpl) serializeToXML(data any) ([]byte, error) {
	// If data is already XML-serializable, use it directly
	switch v := data.(type) {
	case map[string]any:
		// Convert map to struct-based approach for XML compatibility
		return cn.mapToXML(v)
	default:
		// Try to marshal directly
		return xml.Marshal(data)
	}
}

// mapToXML converts map[string]any to proper XML with full feature support
func (cn *ContentNegotiatorImpl) mapToXML(data map[string]any) ([]byte, error) {
	if data == nil {
		return []byte("<response/>"), nil
	}

	// Convert map to XML-serializable structure
	xmlData := cn.convertToXMLStructure(data)

	// Wrap in response element
	response := XMLResponse{
		Data: xmlData,
	}

	// Marshal to XML
	return xml.Marshal(response)
}

// convertToXMLStructure recursively converts data to XML-serializable format
func (cn *ContentNegotiatorImpl) convertToXMLStructure(data any) any {
	if data == nil {
		return ""
	}

	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Map:
		if v.Len() == 0 {
			return ""
		}

		// Convert map to slice of XMLElements
		elements := make([]XMLElement, 0, v.Len())
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			value := v.MapIndex(key).Interface()

			// Skip nil values
			if value == nil {
				continue
			}

			convertedValue := cn.convertToXMLStructure(value)

			element := XMLElement{
				XMLName: xml.Name{Local: cn.sanitizeXMLName(keyStr)},
			}

			// Handle the converted value based on its type
			switch val := convertedValue.(type) {
			case []XMLElement:
				// Nested structure
				element.Elements = val
			case string:
				// Simple text content
				element.Value = val
			default:
				// Other types, convert to string
				element.Value = fmt.Sprintf("%v", val)
			}

			elements = append(elements, element)
		}
		return elements

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return ""
		}

		elements := make([]XMLElement, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			value := v.Index(i).Interface()
			if value == nil {
				continue
			}

			convertedValue := cn.convertToXMLStructure(value)

			element := XMLElement{
				XMLName: xml.Name{Local: "item"},
			}

			switch val := convertedValue.(type) {
			case []XMLElement:
				element.Elements = val
			case string:
				element.Value = val
			default:
				element.Value = fmt.Sprintf("%v", val)
			}

			elements = append(elements, element)
		}
		return elements

	case reflect.String:
		return cn.escapeXMLContent(v.String())

	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", data)

	default:
		// For other types, try to convert to string
		return cn.escapeXMLContent(fmt.Sprintf("%v", data))
	}
}

// sanitizeXMLName ensures XML element names are valid
func (cn *ContentNegotiatorImpl) sanitizeXMLName(name string) string {
	// XML names must start with letter or underscore, and can contain letters, digits, hyphens, underscores, and periods
	var sanitized strings.Builder

	// Replace invalid characters with underscore
	for i, r := range name {
		var valid bool
		if i == 0 {
			valid = (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
		} else {
			valid = (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
				r == '-' || r == '_' || r == '.'
		}
		if !valid {
			sanitized.WriteRune('_')
			continue
		}
		sanitized.WriteRune(r)
	}

	result := sanitized.String()
	if result == "" {
		return "_empty"
	}
	return result
}

// escapeXMLContent properly escapes XML character data
func (cn *ContentNegotiatorImpl) escapeXMLContent(content string) string {
	var buf strings.Builder
	for _, r := range content {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&apos;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
