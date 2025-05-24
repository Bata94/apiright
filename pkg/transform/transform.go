package transform

import (
	"encoding/json"
	"reflect"
	"time"
)

// Transformer provides utilities for transforming between different model types
type Transformer struct {
	fieldMappings map[string]string
	excludeFields []string
	includeFields []string
	customMappers map[string]func(interface{}) interface{}
}

// NewTransformer creates a new transformer
func NewTransformer() *Transformer {
	return &Transformer{
		fieldMappings: make(map[string]string),
		customMappers: make(map[string]func(interface{}) interface{}),
	}
}

// MapField maps a source field to a destination field
func (t *Transformer) MapField(source, dest string) *Transformer {
	t.fieldMappings[source] = dest
	return t
}

// ExcludeFields excludes specific fields from transformation
func (t *Transformer) ExcludeFields(fields ...string) *Transformer {
	t.excludeFields = append(t.excludeFields, fields...)
	return t
}

// IncludeFields includes only specific fields in transformation
func (t *Transformer) IncludeFields(fields ...string) *Transformer {
	t.includeFields = append(t.includeFields, fields...)
	return t
}

// CustomMapper adds a custom transformation function for a field
func (t *Transformer) CustomMapper(field string, mapper func(interface{}) interface{}) *Transformer {
	t.customMappers[field] = mapper
	return t
}

// Transform transforms a source object to a destination type
func (t *Transformer) Transform(source interface{}, destType reflect.Type) (interface{}, error) {
	sourceValue := reflect.ValueOf(source)
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}
	
	destValue := reflect.New(destType).Elem()
	
	return t.transformValue(sourceValue, destValue)
}

// TransformSlice transforms a slice of objects
func (t *Transformer) TransformSlice(source interface{}, destElementType reflect.Type) (interface{}, error) {
	sourceValue := reflect.ValueOf(source)
	if sourceValue.Kind() != reflect.Slice {
		return nil, &TransformError{Message: "source is not a slice"}
	}
	
	destSliceType := reflect.SliceOf(destElementType)
	destSlice := reflect.MakeSlice(destSliceType, 0, sourceValue.Len())
	
	for i := 0; i < sourceValue.Len(); i++ {
		sourceItem := sourceValue.Index(i)
		destItem := reflect.New(destElementType).Elem()
		
		transformed, err := t.transformValue(sourceItem, destItem)
		if err != nil {
			return nil, err
		}
		
		destSlice = reflect.Append(destSlice, reflect.ValueOf(transformed))
	}
	
	return destSlice.Interface(), nil
}

// transformValue performs the actual transformation between two reflect.Values
func (t *Transformer) transformValue(source, dest reflect.Value) (interface{}, error) {
	sourceType := source.Type()
	destType := dest.Type()
	
	// Handle same types
	if sourceType == destType {
		return source.Interface(), nil
	}
	
	// Handle struct to struct transformation
	if sourceType.Kind() == reflect.Struct && destType.Kind() == reflect.Struct {
		return t.transformStruct(source, dest)
	}
	
	// Handle map to struct transformation
	if sourceType.Kind() == reflect.Map && destType.Kind() == reflect.Struct {
		return t.transformMapToStruct(source, dest)
	}
	
	// Handle struct to map transformation
	if sourceType.Kind() == reflect.Struct && destType.Kind() == reflect.Map {
		return t.transformStructToMap(source, dest)
	}
	
	return nil, &TransformError{
		Message: "unsupported transformation",
		Source:  sourceType.String(),
		Dest:    destType.String(),
	}
}

// transformStruct transforms between two structs
func (t *Transformer) transformStruct(source, dest reflect.Value) (interface{}, error) {
	destType := dest.Type()
	
	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := dest.Field(i)
		
		if !destFieldValue.CanSet() {
			continue
		}
		
		// Check if field should be included/excluded
		if !t.shouldIncludeField(destField.Name) {
			continue
		}
		
		// Find corresponding source field
		sourceFieldName := t.getSourceFieldName(destField.Name)
		sourceFieldValue, found := t.findSourceField(source, sourceFieldName)
		
		if !found {
			continue
		}
		
		// Apply custom mapper if exists
		if mapper, exists := t.customMappers[destField.Name]; exists {
			mappedValue := mapper(sourceFieldValue.Interface())
			if mappedValue != nil {
				destFieldValue.Set(reflect.ValueOf(mappedValue))
			}
			continue
		}
		
		// Handle type conversion
		if err := t.setFieldValue(destFieldValue, sourceFieldValue); err != nil {
			return nil, err
		}
	}
	
	return dest.Interface(), nil
}

// transformMapToStruct transforms a map to a struct
func (t *Transformer) transformMapToStruct(source, dest reflect.Value) (interface{}, error) {
	destType := dest.Type()
	
	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := dest.Field(i)
		
		if !destFieldValue.CanSet() {
			continue
		}
		
		if !t.shouldIncludeField(destField.Name) {
			continue
		}
		
		// Get field name (check for json tag)
		fieldName := t.getFieldName(destField)
		
		// Get value from map
		mapValue := source.MapIndex(reflect.ValueOf(fieldName))
		if !mapValue.IsValid() {
			continue
		}
		
		// Apply custom mapper if exists
		if mapper, exists := t.customMappers[destField.Name]; exists {
			mappedValue := mapper(mapValue.Interface())
			if mappedValue != nil {
				destFieldValue.Set(reflect.ValueOf(mappedValue))
			}
			continue
		}
		
		// Set field value
		if err := t.setFieldValue(destFieldValue, mapValue); err != nil {
			return nil, err
		}
	}
	
	return dest.Interface(), nil
}

// transformStructToMap transforms a struct to a map
func (t *Transformer) transformStructToMap(source, dest reflect.Value) (interface{}, error) {
	sourceType := source.Type()
	mapType := dest.Type()
	
	result := reflect.MakeMap(mapType)
	
	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		sourceFieldValue := source.Field(i)
		
		if !t.shouldIncludeField(sourceField.Name) {
			continue
		}
		
		fieldName := t.getFieldName(sourceField)
		
		// Apply custom mapper if exists
		var value interface{}
		if mapper, exists := t.customMappers[sourceField.Name]; exists {
			value = mapper(sourceFieldValue.Interface())
		} else {
			value = sourceFieldValue.Interface()
		}
		
		if value != nil {
			result.SetMapIndex(reflect.ValueOf(fieldName), reflect.ValueOf(value))
		}
	}
	
	return result.Interface(), nil
}

// Helper methods

func (t *Transformer) shouldIncludeField(fieldName string) bool {
	// Check exclude list
	for _, excluded := range t.excludeFields {
		if excluded == fieldName {
			return false
		}
	}
	
	// Check include list (if specified)
	if len(t.includeFields) > 0 {
		for _, included := range t.includeFields {
			if included == fieldName {
				return true
			}
		}
		return false
	}
	
	return true
}

func (t *Transformer) getSourceFieldName(destFieldName string) string {
	if mapped, exists := t.fieldMappings[destFieldName]; exists {
		return mapped
	}
	return destFieldName
}

func (t *Transformer) findSourceField(source reflect.Value, fieldName string) (reflect.Value, bool) {
	sourceType := source.Type()
	
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		if field.Name == fieldName {
			return source.Field(i), true
		}
	}
	
	return reflect.Value{}, false
}

func (t *Transformer) getFieldName(field reflect.StructField) string {
	// Check for json tag
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		return tag
	}
	return field.Name
}

func (t *Transformer) setFieldValue(dest, source reflect.Value) error {
	sourceType := source.Type()
	destType := dest.Type()
	
	// Direct assignment if types match
	if sourceType == destType {
		dest.Set(source)
		return nil
	}
	
	// Handle convertible types
	if sourceType.ConvertibleTo(destType) {
		dest.Set(source.Convert(destType))
		return nil
	}
	
	// Handle pointer types
	if destType.Kind() == reflect.Ptr && sourceType == destType.Elem() {
		ptr := reflect.New(sourceType)
		ptr.Elem().Set(source)
		dest.Set(ptr)
		return nil
	}
	
	if sourceType.Kind() == reflect.Ptr && destType == sourceType.Elem() {
		if !source.IsNil() {
			dest.Set(source.Elem())
		}
		return nil
	}
	
	// Handle time.Time special cases
	if destType == reflect.TypeOf(time.Time{}) {
		return t.setTimeValue(dest, source)
	}
	
	return &TransformError{
		Message: "cannot convert field",
		Source:  sourceType.String(),
		Dest:    destType.String(),
	}
}

func (t *Transformer) setTimeValue(dest, source reflect.Value) error {
	sourceInterface := source.Interface()
	
	switch v := sourceInterface.(type) {
	case string:
		if parsedTime, err := time.Parse(time.RFC3339, v); err == nil {
			dest.Set(reflect.ValueOf(parsedTime))
			return nil
		}
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			dest.Set(reflect.ValueOf(parsedTime))
			return nil
		}
	case int64:
		dest.Set(reflect.ValueOf(time.Unix(v, 0)))
		return nil
	case time.Time:
		dest.Set(reflect.ValueOf(v))
		return nil
	}
	
	return &TransformError{Message: "cannot convert to time.Time"}
}

// TransformError represents a transformation error
type TransformError struct {
	Message string
	Source  string
	Dest    string
}

func (e *TransformError) Error() string {
	if e.Source != "" && e.Dest != "" {
		return e.Message + ": " + e.Source + " -> " + e.Dest
	}
	return e.Message
}

// Common transformers

// TimeToString creates a transformer that converts time.Time to string
func TimeToString(format string) func(interface{}) interface{} {
	return func(value interface{}) interface{} {
		if t, ok := value.(time.Time); ok {
			return t.Format(format)
		}
		return value
	}
}

// StringToTime creates a transformer that converts string to time.Time
func StringToTime(format string) func(interface{}) interface{} {
	return func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			if t, err := time.Parse(format, s); err == nil {
				return t
			}
		}
		return value
	}
}

// JSONTransformer provides JSON-based transformation
type JSONTransformer struct{}

// ToAPI converts any struct to a map[string]interface{} via JSON
func (jt *JSONTransformer) ToAPI(dbModel interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(dbModel)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// FromAPI converts a map[string]interface{} to a struct via JSON
func (jt *JSONTransformer) FromAPI(apiModel interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(apiModel)
	if err != nil {
		return nil, err
	}
	
	// This would need the target type, which requires generics or reflection
	// For now, return the JSON data for manual unmarshaling
	return jsonData, nil
}