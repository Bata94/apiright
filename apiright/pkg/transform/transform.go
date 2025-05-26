package transform

import (
	"reflect"
	"strings"
)

// Transformer handles conversion between different model types
type Transformer[TSource any, TTarget any] struct {
	fieldMapping map[string]string
}

// NewTransformer creates a new transformer between source and target types
func NewTransformer[TSource any, TTarget any]() *Transformer[TSource, TTarget] {
	return &Transformer[TSource, TTarget]{
		fieldMapping: make(map[string]string),
	}
}

// WithFieldMapping adds a custom field mapping from target field to source field
func (t *Transformer[TSource, TTarget]) WithFieldMapping(sourceField, targetField string) *Transformer[TSource, TTarget] {
	t.fieldMapping[targetField] = sourceField
	return t
}

// Transform converts a source object to target type
func (t *Transformer[TSource, TTarget]) Transform(source TSource) (TTarget, error) {
	var target TTarget
	
	sourceVal := reflect.ValueOf(source)
	targetVal := reflect.ValueOf(&target).Elem()
	
	// Handle pointer types for source
	if sourceVal.Kind() == reflect.Ptr {
		if sourceVal.IsNil() {
			return target, nil
		}
		sourceVal = sourceVal.Elem()
	}
	
	// Handle pointer types for target
	if targetVal.Kind() == reflect.Ptr {
		// Create a new instance of the target type
		targetType := targetVal.Type().Elem()
		newTarget := reflect.New(targetType)
		targetVal.Set(newTarget)
		targetVal = newTarget.Elem()
	}
	
	sourceType := sourceVal.Type()
	targetType := targetVal.Type()
	
	// Copy fields from source to target
	for i := 0; i < targetType.NumField(); i++ {
		targetField := targetType.Field(i)
		targetFieldVal := targetVal.Field(i)
		
		if !targetFieldVal.CanSet() {
			continue
		}
		
		// Find corresponding source field
		sourceFieldName := t.getSourceFieldName(targetField.Name)
		sourceFieldVal, found := t.findSourceField(sourceVal, sourceType, sourceFieldName)
		
		if !found {
			continue
		}
		
		// Convert and set the value
		if err := t.setFieldValue(targetFieldVal, sourceFieldVal); err != nil {
			return target, err
		}
	}
	
	return target, nil
}

// TransformSlice converts a slice of source objects to target type
func (t *Transformer[TSource, TTarget]) TransformSlice(sources []TSource) ([]TTarget, error) {
	targets := make([]TTarget, len(sources))
	
	for i, source := range sources {
		target, err := t.Transform(source)
		if err != nil {
			return nil, err
		}
		targets[i] = target
	}
	
	return targets, nil
}

// getSourceFieldName returns the source field name for a target field
func (t *Transformer[TSource, TTarget]) getSourceFieldName(targetFieldName string) string {
	if mappedName, exists := t.fieldMapping[targetFieldName]; exists {
		return mappedName
	}
	return targetFieldName
}

// findSourceField finds a field in the source struct by name
func (t *Transformer[TSource, TTarget]) findSourceField(sourceVal reflect.Value, sourceType reflect.Type, fieldName string) (reflect.Value, bool) {
	// Try exact match first
	if field := sourceVal.FieldByName(fieldName); field.IsValid() {
		return field, true
	}
	
	// Try case-insensitive match
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		if strings.EqualFold(field.Name, fieldName) {
			return sourceVal.Field(i), true
		}
		
		// Check JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if strings.EqualFold(tagName, fieldName) {
				return sourceVal.Field(i), true
			}
		}
		
		// Check db tag
		if dbTag := field.Tag.Get("db"); dbTag != "" {
			if strings.EqualFold(dbTag, fieldName) {
				return sourceVal.Field(i), true
			}
		}
	}
	
	return reflect.Value{}, false
}

// setFieldValue sets a target field value from a source field value
func (t *Transformer[TSource, TTarget]) setFieldValue(targetField, sourceField reflect.Value) error {
	if !sourceField.IsValid() {
		return nil
	}
	
	sourceType := sourceField.Type()
	targetType := targetField.Type()
	
	// Direct assignment if types match
	if sourceType.AssignableTo(targetType) {
		targetField.Set(sourceField)
		return nil
	}
	
	// Handle convertible types
	if sourceType.ConvertibleTo(targetType) {
		targetField.Set(sourceField.Convert(targetType))
		return nil
	}
	
	// Handle pointer to value conversion
	if sourceType.Kind() == reflect.Ptr && sourceType.Elem().AssignableTo(targetType) {
		if !sourceField.IsNil() {
			targetField.Set(sourceField.Elem())
		}
		return nil
	}
	
	// Handle value to pointer conversion
	if targetType.Kind() == reflect.Ptr && sourceType.AssignableTo(targetType.Elem()) {
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(sourceField)
		targetField.Set(ptr)
		return nil
	}
	
	return nil
}

// BiDirectionalTransformer handles transformation in both directions
type BiDirectionalTransformer[T1 any, T2 any] struct {
	forward  *Transformer[T1, T2]
	backward *Transformer[T2, T1]
}

// NewBiDirectionalTransformer creates a transformer that works in both directions
func NewBiDirectionalTransformer[T1 any, T2 any]() *BiDirectionalTransformer[T1, T2] {
	return &BiDirectionalTransformer[T1, T2]{
		forward:  NewTransformer[T1, T2](),
		backward: NewTransformer[T2, T1](),
	}
}

// WithFieldMapping adds field mapping for both directions
func (bt *BiDirectionalTransformer[T1, T2]) WithFieldMapping(field1, field2 string) *BiDirectionalTransformer[T1, T2] {
	bt.forward.WithFieldMapping(field1, field2)
	bt.backward.WithFieldMapping(field2, field1)
	return bt
}

// Forward transforms from T1 to T2
func (bt *BiDirectionalTransformer[T1, T2]) Forward(source T1) (T2, error) {
	return bt.forward.Transform(source)
}

// Backward transforms from T2 to T1
func (bt *BiDirectionalTransformer[T1, T2]) Backward(source T2) (T1, error) {
	return bt.backward.Transform(source)
}

// ForwardSlice transforms a slice from T1 to T2
func (bt *BiDirectionalTransformer[T1, T2]) ForwardSlice(sources []T1) ([]T2, error) {
	return bt.forward.TransformSlice(sources)
}

// BackwardSlice transforms a slice from T2 to T1
func (bt *BiDirectionalTransformer[T1, T2]) BackwardSlice(sources []T2) ([]T1, error) {
	return bt.backward.TransformSlice(sources)
}

// Quick transformation functions for common use cases

// Transform converts a source object to target type using automatic field mapping
func Transform[TSource any, TTarget any](source TSource) (TTarget, error) {
	transformer := NewTransformer[TSource, TTarget]()
	return transformer.Transform(source)
}

// TransformSlice converts a slice of source objects to target type
func TransformSlice[TSource any, TTarget any](sources []TSource) ([]TTarget, error) {
	transformer := NewTransformer[TSource, TTarget]()
	return transformer.TransformSlice(sources)
}