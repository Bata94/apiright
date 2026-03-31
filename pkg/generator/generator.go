package generator

import (
	"context"
	"fmt"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/plugins"
)

// Generator orchestrates the complete code generation pipeline
type Generator struct {
	parser            *SchemaParser
	sqlGen            *SQLGenerator
	sqlcRunner        *SQLCRunner
	protoGen          *ProtoGenerator
	protoExtProcessor *ProtoExtensionProcessor
	serviceGen        *ServiceGenerator
	openapiGen        *OpenAPIGenerator
	cache             *Cache
	plugins           *plugins.PluginRegistry
	logger            core.Logger
}

// GenerateOptions controls generation behavior
type GenerateOptions struct {
	Force     bool // --force flag
	SQLOnly   bool // --sql-only flag
	GoOnly    bool // --go-only flag
	ProtoOnly bool // --proto-only flag
	Verbose   bool // --verbose flag
	DryRun    bool // --dry-run flag
}

// NewGenerator creates a new generator instance
func NewGenerator(projectDir string, options GenerateOptions, logger core.Logger, pluginRegistry *plugins.PluginRegistry) (*Generator, error) {
	// Load configuration
	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create components
	cache, err := NewCache(projectDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	parser := NewSchemaParser(cfg.Database.Type, logger)

	// Convert database type to dialect
	var dialect Dialect
	switch cfg.Database.Type {
	case "postgres", "postgresql":
		dialect = DialectPostgres
	case "mysql":
		dialect = DialectMySQL
	default:
		dialect = DialectSQLite
	}

	sqlGen := NewSQLGenerator(cfg.Generation.GenSuffix, dialect, logger)
	sqlcRunner := NewSQLCRunner(projectDir, options.Verbose, logger)
	protoGen := NewProtoGenerator(cfg.Generation.GenSuffix, logger)
	protoExtProcessor := NewProtoExtensionProcessor(cfg.Generation.GenSuffix, logger)
	serviceGen := NewServiceGenerator(cfg.Generation.GenSuffix, logger)
	openapiGen := NewOpenAPIGenerator(cfg.Generation.GenSuffix, logger)

	return &Generator{
		parser:            parser,
		sqlGen:            sqlGen,
		sqlcRunner:        sqlcRunner,
		protoGen:          protoGen,
		protoExtProcessor: protoExtProcessor,
		serviceGen:        serviceGen,
		openapiGen:        openapiGen,
		cache:             cache,
		plugins:           pluginRegistry,
		logger:            logger,
	}, nil
}

// Generate executes the complete generation pipeline
func (g *Generator) Generate(ctx *core.GenerationContext, options GenerateOptions) error {
	g.logger.Info("Starting code generation", "project", ctx.ProjectDir)

	if options.DryRun {
		g.logger.Info("Dry run mode - no files will be written")
	}

	spinner := core.NewSpinner("Generating code")
	spinner.Start()

	// 1. Check cache first (unless force regeneration)
	if !options.Force {
		shouldRegen, err := g.cache.ShouldRegenerate(
			ctx.Join(ctx.ProjectDir, "migrations"),
			ctx.Join(ctx.ProjectDir, "sqlc.yaml"),
		)
		if err != nil {
			g.logger.Warn("Cache check failed, proceeding with generation", "error", err)
		} else if !shouldRegen {
			if err := g.cache.RestoreFromCache(ctx); err == nil {
				g.logger.Info("Restored from cache - skipping generation")
				return nil
			}
			g.logger.Warn("Failed to restore from cache, proceeding with generation", "error", err)
		}
	}

	// 2. Execute before generation hooks
	if g.plugins != nil {
		if err := g.plugins.ExecuteHooks("before_generation", ctx); err != nil {
			return g.formatError("before_generation_hooks", err, "")
		}
	}

	// 3. Parse SQL schema
	spinner.SetMessage("Parsing SQL migrations")
	schema, err := g.parser.ParseMigrations(ctx.Join(ctx.ProjectDir, "migrations"))
	if err != nil {
		return g.formatError("schema_parsing", err, "migrations directory")
	}

	// Update context with schema
	ctx.WithSchema(schema)

	// 4. Generate SQL queries (unless go-only)
	if !options.GoOnly {
		spinner.SetMessage("Generating SQL queries")
		if err := g.sqlGen.GenerateQueries(schema, ctx); err != nil {
			return g.formatError("sql_generation", err, "")
		}
		g.logger.Info("Generated SQL queries", "tables", len(schema.Tables))
	}

	// 5. Execute plugin hooks before sqlc (unless sql-only)
	if g.plugins != nil && !options.SQLOnly {
		if err := g.plugins.ExecuteHooks("before_sqlc", ctx); err != nil {
			return g.formatError("before_sqlc_hooks", err, "")
		}
	}

	// 6. Run sqlc (unless sql-only)
	if !options.SQLOnly {
		spinner.SetMessage("Running sqlc code generation")
		if err := g.sqlcRunner.Generate(context.Background()); err != nil {
			return g.formatError("sqlc_execution", err, "sqlc.yaml")
		}
		g.logger.Info("sqlc generation completed")
	}

	// 7. Execute plugin hooks after sqlc (unless sql-only)
	if g.plugins != nil && !options.SQLOnly {
		if err := g.plugins.ExecuteHooks("after_sqlc", ctx); err != nil {
			return g.formatError("after_sqlc_hooks", err, "")
		}
	}

	// 8. Execute plugin hooks before protobuf
	if g.plugins != nil && !options.SQLOnly && !options.GoOnly {
		if err := g.plugins.ExecuteHooks("before_protobuf", ctx); err != nil {
			return g.formatError("before_protobuf_hooks", err, "")
		}
	}

	// 9. Generate protobuf (unless sql-only or go-only)
	if !options.SQLOnly && !options.GoOnly {
		spinner.SetMessage("Generating protobuf definitions")
		if err := g.protoGen.Generate(schema, ctx); err != nil {
			return g.formatError("protobuf_generation", err, "")
		}
		g.logger.Info("Generated protobuf definitions")
	}

	// 10. Execute plugin hooks after protobuf
	if g.plugins != nil && !options.SQLOnly && !options.GoOnly {
		if err := g.plugins.ExecuteHooks("after_protobuf", ctx); err != nil {
			return g.formatError("after_protobuf_hooks", err, "")
		}
	}

	// 10.5 Process proto extensions
	if !options.SQLOnly && !options.GoOnly {
		var extensions []core.ProtoExtension
		if g.plugins != nil {
			extensions = g.getProtoExtensions()
		}
		if err := g.protoExtProcessor.ProcessExtensions(ctx, extensions); err != nil {
			return g.formatError("proto_extension_processing", err, "")
		}
		if g.protoExtProcessor.HasExtensions() {
			if err := g.protoExtProcessor.CopyExtensionsToGen(ctx); err != nil {
				g.logger.Warn("Failed to copy proto extensions", "error", err)
			}
		}
		g.logger.Info("Processed proto extensions", "count", len(extensions))
	}

	// 11. Generate OpenAPI documentation (unless sql-only or go-only)
	if !options.SQLOnly && !options.GoOnly {
		spinner.SetMessage("Generating OpenAPI documentation")
		if err := g.openapiGen.Generate(schema, ctx); err != nil {
			return g.formatError("openapi_generation", err, "")
		}
		g.logger.Info("Generated OpenAPI documentation")
	}

	// 12. Generate service implementations (unless sql-only or proto-only)
	if !options.SQLOnly && !options.ProtoOnly {
		spinner.SetMessage("Generating service implementations")
		if err := g.serviceGen.GenerateServices(schema, ctx); err != nil {
			return g.formatError("service_generation", err, "")
		}
		g.logger.Info("Generated service implementations")
	}

	// 13. Save to cache (unless sql-only)
	if !options.SQLOnly {
		if err := g.cache.SaveToCache(ctx,
			ctx.Join(ctx.ProjectDir, "migrations"),
			ctx.Join(ctx.ProjectDir, "sqlc.yaml"),
		); err != nil {
			g.logger.Warn("Failed to save to cache", "error", err)
		}
	}

	// 14. Execute after generation hooks
	if g.plugins != nil {
		if err := g.plugins.ExecuteHooks("after_generation", ctx); err != nil {
			return g.formatError("after_generation_hooks", err, "")
		}
	}

	spinner.Success("Code generation complete")
	return nil
}

// formatError creates a formatted error with context information
func (g *Generator) formatError(errorType string, err error, context string) error {
	// Map error types to user-friendly messages
	errorMessages := map[string]string{
		"schema_parsing":          "Failed to parse SQL migration files",
		"sql_generation":          "Failed to generate CRUD SQL queries",
		"sqlc_execution":          "sqlc code generation failed",
		"protobuf_generation":     "Failed to generate protobuf definitions",
		"openapi_generation":      "Failed to generate OpenAPI documentation",
		"service_generation":      "Failed to generate service implementations",
		"before_generation_hooks": "Before-generation plugin hooks failed",
		"before_sqlc_hooks":       "Before-sqlc plugin hooks failed",
		"after_sqlc_hooks":        "After-sqlc plugin hooks failed",
		"before_protobuf_hooks":   "Before-protobuf plugin hooks failed",
		"after_protobuf_hooks":    "After-protobuf plugin hooks failed",
		"after_generation_hooks":  "After-generation plugin hooks failed",
		"cache_operation":         "Cache operation failed",
	}

	baseMessage := errorMessages[errorType]
	if baseMessage == "" {
		baseMessage = fmt.Sprintf("Generation error in %s", errorType)
	}

	if context != "" {
		baseMessage = fmt.Sprintf("%s: %s", baseMessage, context)
	}

	return fmt.Errorf("%s: %w", baseMessage, err)
}

// ProgressReporter reports generation progress to user
type ProgressReporter struct {
	logger  core.Logger
	current int
	total   int
	stage   string
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(total int, stage string, logger core.Logger) *ProgressReporter {
	return &ProgressReporter{
		current: 0,
		total:   total,
		stage:   stage,
		logger:  logger,
	}
}

// Update increments and reports progress
func (pr *ProgressReporter) Update(item string) {
	pr.current++

	if pr.total > 0 {
		percent := (pr.current * 100) / pr.total
		pr.logger.Info(fmt.Sprintf("%s progress: %d%% - %s", pr.stage, percent, item))
	} else {
		pr.logger.Info(fmt.Sprintf("%s: %s", pr.stage, item))
	}
}

// Complete marks the progress as complete
func (pr *ProgressReporter) Complete() {
	pr.logger.Info(fmt.Sprintf("%s completed: %d items processed", pr.stage, pr.current))
}

func (g *Generator) getProtoExtensions() []core.ProtoExtension {
	var extensions []core.ProtoExtension
	for _, p := range g.plugins.ListPlugins() {
		if ext, ok := p.(core.ProtoExtension); ok {
			extensions = append(extensions, ext)
		}
	}
	return extensions
}
