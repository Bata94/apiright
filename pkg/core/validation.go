package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

type ValidationRule interface {
	Validate(field string, value interface{}) error
}

type RequiredRule struct{}

func (r *RequiredRule) Validate(field string, value interface{}) error {
	if value == nil {
		return ValidationError{Field: field, Message: "is required"}
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		if rv.String() == "" {
			return ValidationError{Field: field, Message: "is required"}
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if rv.Len() == 0 {
			return ValidationError{Field: field, Message: "is required"}
		}
	}
	return nil
}

type MinLenRule struct {
	Min int
}

func (r *MinLenRule) Validate(field string, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.String {
		if rv.Len() < r.Min {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d characters", r.Min)}
		}
	}
	return nil
}

type MaxLenRule struct {
	Max int
}

func (r *MaxLenRule) Validate(field string, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.String {
		if rv.Len() > r.Max {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d characters", r.Max)}
		}
	}
	return nil
}

type EmailRule struct{}

func (r *EmailRule) Validate(field string, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.String {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(rv.String()) {
			return ValidationError{Field: field, Message: "must be a valid email address"}
		}
	}
	return nil
}

type MinValueRule struct {
	Min int64
}

func (r *MinValueRule) Validate(field string, value interface{}) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() < r.Min {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d", r.Min)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if int64(rv.Uint()) < r.Min {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d", r.Min)}
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() < float64(r.Min) {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d", r.Min)}
		}
	}
	return nil
}

type MaxValueRule struct {
	Max int64
}

func (r *MaxValueRule) Validate(field string, value interface{}) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() > r.Max {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d", r.Max)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if int64(rv.Uint()) > r.Max {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d", r.Max)}
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() > float64(r.Max) {
			return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d", r.Max)}
		}
	}
	return nil
}

type BasicValidator struct {
	rules map[string][]ValidationRule
}

func NewBasicValidator() *BasicValidator {
	return &BasicValidator{
		rules: make(map[string][]ValidationRule),
	}
}

func (v *BasicValidator) AddRule(field string, rule ValidationRule) {
	v.rules[field] = append(v.rules[field], rule)
}

func (v *BasicValidator) AddRequired(field string) {
	v.AddRule(field, &RequiredRule{})
}

func (v *BasicValidator) AddMinLen(field string, min int) {
	v.AddRule(field, &MinLenRule{Min: min})
}

func (v *BasicValidator) AddMaxLen(field string, max int) {
	v.AddRule(field, &MaxLenRule{Max: max})
}

func (v *BasicValidator) AddEmail(field string) {
	v.AddRule(field, &EmailRule{})
}

func (v *BasicValidator) AddMinValue(field string, min int64) {
	v.AddRule(field, &MinValueRule{Min: min})
}

func (v *BasicValidator) AddMaxValue(field string, max int64) {
	v.AddRule(field, &MaxValueRule{Max: max})
}

func (v *BasicValidator) Validate(data map[string]interface{}) error {
	var errors ValidationErrors
	for field, rules := range v.rules {
		value, exists := data[field]
		if !exists {
			value = nil
		}
		for _, rule := range rules {
			if err := rule.Validate(field, value); err != nil {
				errors = append(errors, err.(ValidationError))
			}
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

type ColumnValidator struct {
	columnRules map[string][]ValidationRule
}

func NewColumnValidator() *ColumnValidator {
	return &ColumnValidator{
		columnRules: make(map[string][]ValidationRule),
	}
}

func (cv *ColumnValidator) AddRule(column string, rule ValidationRule) {
	cv.columnRules[column] = append(cv.columnRules[column], rule)
}

func (cv *ColumnValidator) FromSchema(table string, schema *Schema) {
	for _, t := range schema.Tables {
		if t.Name == table {
			for _, col := range t.Columns {
				if !col.Nullable {
					cv.AddRule(col.Name, &RequiredRule{})
				}
				if col.Default == "" && !col.Nullable {
					cv.AddRule(col.Name, &RequiredRule{})
				}
			}
		}
	}
}

func (cv *ColumnValidator) Validate(data map[string]interface{}) error {
	var errors ValidationErrors
	for field, rules := range cv.columnRules {
		value, exists := data[field]
		if !exists {
			value = nil
		}
		for _, rule := range rules {
			if err := rule.Validate(field, value); err != nil {
				errors = append(errors, err.(ValidationError))
			}
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}
