package core

import (
	"io"
	"net/http"
)

// RequestUnmarshaller is an interface for unmarshalling request bodies.
type RequestUnmarshaller interface {
	Unmarshal(data []byte, v any) error
	ContentType() string
}

// ResponseMarshaller is an interface for marshalling response bodies.
type ResponseMarshaller interface {
	Marshal(v any) ([]byte, error)
	ContentType() string
}

// JSONUnmarshaller implements RequestUnmarshaller for JSON.
type JSONUnmarshaller struct{}

func (j JSONUnmarshaller) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (j JSONUnmarshaller) ContentType() string {
	return MIMETYPE_JSON.toString()
}

// JSONMarshaller implements ResponseMarshaller for JSON.
type JSONMarshaller struct{}

func (j JSONMarshaller) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (j JSONMarshaller) ContentType() string {
	return MIMETYPE_JSON.toString()
}

// XMLUnmarshaller implements RequestUnmarshaller for XML.
type XMLUnmarshaller struct{}

func (x XMLUnmarshaller) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

func (x XMLUnmarshaller) ContentType() string {
	return MIMETYPE_XML.toString()
}

// XMLMarshaller implements ResponseMarshaller for XML.
type XMLMarshaller struct{}

func (x XMLMarshaller) Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}

func (x XMLMarshaller) ContentType() string {
	return MIMETYPE_XML.toString()
}

// YAMLUnmarshaller implements RequestUnmarshaller for YAML.
type YAMLUnmarshaller struct{}

func (y YAMLUnmarshaller) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func (y YAMLUnmarshaller) ContentType() string {
	return MIMETYPE_YAML.toString()
}

// YAMLMarshaller implements ResponseMarshaller for YAML.
type YAMLMarshaller struct{}

func (y YAMLMarshaller) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y YAMLMarshaller) ContentType() string {
	return MIMETYPE_YAML.toString()
}
