package plugins

import (
	"fmt"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

type CustomValidationPlugin struct {
	rules  map[string][]ValidationRule
	logger core.Logger
}

type ValidationRule struct {
	Field    string
	Type     string
	MinLen   int
	MaxLen   int
	Required bool
}

func NewCustomValidationPlugin() *CustomValidationPlugin {
	return &CustomValidationPlugin{
		rules: make(map[string][]ValidationRule),
	}
}

func (p *CustomValidationPlugin) Name() string {
	return "custom-validation"
}

func (p *CustomValidationPlugin) Version() string {
	return "1.0.0"
}

func (p *CustomValidationPlugin) Generate(ctx *core.GenerationContext) error {
	ctx.Log("info", "Custom validation plugin executed")

	for _, table := range ctx.Schema.Tables {
		tableName := strings.TrimSuffix(table.Name, "_ar_gen")

		if rules, ok := p.rules[tableName]; ok {
			for _, rule := range rules {
				ctx.Log("debug", "Applying validation rule",
					"table", tableName,
					"field", rule.Field,
					"type", rule.Type)
			}
		}
	}

	return nil
}

func (p *CustomValidationPlugin) Validate(schema *core.Schema) error {
	for tableName := range p.rules {
		found := false
		for _, table := range schema.Tables {
			if strings.TrimSuffix(table.Name, "_ar_gen") == tableName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("validation rule references unknown table: %s", tableName)
		}
	}
	return nil
}

func (p *CustomValidationPlugin) AddRule(table string, rule ValidationRule) {
	p.rules[table] = append(p.rules[table], rule)
}

func (p *CustomValidationPlugin) SetLogger(logger core.Logger) {
	p.logger = logger
}
