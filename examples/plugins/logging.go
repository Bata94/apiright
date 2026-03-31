package plugins

import (
	"fmt"

	"github.com/bata94/apiright/pkg/core"
)

type CustomLoggingPlugin struct {
	level  string
	logger core.Logger
}

func NewCustomLoggingPlugin(level string) *CustomLoggingPlugin {
	return &CustomLoggingPlugin{
		level: level,
	}
}

func (p *CustomLoggingPlugin) Name() string {
	return "custom-logging"
}

func (p *CustomLoggingPlugin) Version() string {
	return "1.0.0"
}

func (p *CustomLoggingPlugin) Generate(ctx *core.GenerationContext) error {
	ctx.Log("info", "Custom logging plugin executed", "level", p.level)

	for _, table := range ctx.Schema.Tables {
		ctx.Log("debug", "Processing table", "table", table.Name)
	}

	return nil
}

func (p *CustomLoggingPlugin) Validate(schema *core.Schema) error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[p.level] {
		return fmt.Errorf("invalid log level: %s", p.level)
	}
	return nil
}

func (p *CustomLoggingPlugin) SetLogger(logger core.Logger) {
	p.logger = logger
}
