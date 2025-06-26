package openapi

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Writer handles writing OpenAPI documentation to various formats
type Writer struct {
	generator *Generator
}

// NewWriter creates a new documentation writer
func NewWriter(generator *Generator) *Writer {
	return &Writer{
		generator: generator,
	}
}

// WriteFiles generates and writes all configured output formats
func (w *Writer) WriteFiles() error {
	spec, err := w.generator.GenerateSpec()
	if err != nil {
		return fmt.Errorf("failed to generate spec: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(w.generator.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write JSON
	if w.generator.config.GenerateJSON {
		if err := w.WriteJSON(spec); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	}

	// Write YAML
	if w.generator.config.GenerateYAML {
		if err := w.WriteYAML(spec); err != nil {
			return fmt.Errorf("failed to write YAML: %w", err)
		}
	}

	// Write HTML
	if w.generator.config.GenerateHTML {
		if err := w.WriteHTML(spec); err != nil {
			return fmt.Errorf("failed to write HTML: %w", err)
		}
	}

	return nil
}

// WriteJSON writes the specification as JSON
func (w *Writer) WriteJSON(spec *OpenAPISpec) error {
	var data []byte
	var err error

	if w.generator.config.PrettyPrint {
		data, err = json.MarshalIndent(spec, "", "  ")
	} else {
		data, err = json.Marshal(spec)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := w.generator.GetOutputPath("openapi.json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// WriteYAML writes the specification as YAML
func (w *Writer) WriteYAML(spec *OpenAPISpec) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	filename := w.generator.GetOutputPath("openapi.yaml")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// WriteHTML writes the specification as HTML with Swagger UI
func (w *Writer) WriteHTML(spec *OpenAPISpec) error {
	// Generate the main HTML file
	if err := w.writeSwaggerUI(); err != nil {
		return fmt.Errorf("failed to write Swagger UI: %w", err)
	}

	// Write the spec as JSON for Swagger UI to load
	specJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec for HTML: %w", err)
	}

	filename := w.generator.GetOutputPath("spec.json")
	if err := os.WriteFile(filename, specJSON, 0644); err != nil {
		return fmt.Errorf("failed to write spec JSON for HTML: %w", err)
	}

	return nil
}

// writeSwaggerUI generates the Swagger UI HTML file
func (w *Writer) writeSwaggerUI() error {
	tmpl := template.Must(template.New("swagger").Parse(swaggerUITemplate))

	data := struct {
		Title       string
		SpecURL     string
		GeneratedAt string
		SpecJSON    string
	}{
		Title:       w.generator.spec.Info.Title,
		SpecURL:     "spec.json",
		GeneratedAt: GeneratedAt(),
		SpecJSON:    "",
	}

	filename := w.generator.GetOutputPath("index.html")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// WriteToFile writes the specification to a specific file
func (w *Writer) WriteToFile(filename string, format string) error {
	spec, err := w.generator.GenerateSpec()
	if err != nil {
		return fmt.Errorf("failed to generate spec: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	switch strings.ToLower(format) {
	case "json":
		return w.writeJSONToFile(spec, filename)
	case "yaml", "yml":
		return w.writeYAMLToFile(spec, filename)
	case "html":
		return w.writeHTMLToFile(spec, filename)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (w *Writer) writeJSONToFile(spec *OpenAPISpec, filename string) error {
	var data []byte
	var err error

	if w.generator.config.PrettyPrint {
		data, err = json.MarshalIndent(spec, "", "  ")
	} else {
		data, err = json.Marshal(spec)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func (w *Writer) writeYAMLToFile(spec *OpenAPISpec, filename string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func (w *Writer) writeHTMLToFile(spec *OpenAPISpec, filename string) error {
	tmpl := template.Must(template.New("swagger").Parse(swaggerUITemplate))

	data := struct {
		Title       string
		SpecURL     string
		GeneratedAt string
		SpecJSON    string
	}{
		Title:       w.generator.spec.Info.Title,
		SpecURL:     "",
		GeneratedAt: GeneratedAt(),
	}

	// Embed the spec directly in the HTML
	specJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}
	data.SpecJSON = string(specJSON)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return tmpl.Execute(file, data)
}

// GetGeneratedFiles returns a list of files that would be generated
func (w *Writer) GetGeneratedFiles() []string {
	var files []string

	if w.generator.config.GenerateJSON {
		files = append(files, w.generator.GetOutputPath("openapi.json"))
	}

	if w.generator.config.GenerateYAML {
		files = append(files, w.generator.GetOutputPath("openapi.yaml"))
	}

	if w.generator.config.GenerateHTML {
		files = append(files,
			w.generator.GetOutputPath("index.html"),
			w.generator.GetOutputPath("spec.json"),
		)
	}

	return files
}

// CleanOutputDir removes all generated files
func (w *Writer) CleanOutputDir() error {
	files := w.GetGeneratedFiles()

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove file %s: %w", file, err)
		}
	}

	return nil
}

// swaggerUITemplate is the HTML template for Swagger UI
const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}} - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #2c3e50;
        }
        .swagger-ui .topbar .download-url-wrapper {
            display: none;
        }
        .info {
            margin: 20px 0;
        }
        .info h1 {
            color: #2c3e50;
        }
        .generated-info {
            background: #ecf0f1;
            padding: 10px;
            border-radius: 5px;
            margin: 20px;
            font-size: 14px;
            color: #7f8c8d;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <div class="generated-info">
        Generated at: {{.GeneratedAt}}
    </div>

    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            {{if .SpecJSON}}
            // Embedded spec
            const spec = {{.SpecJSON}};
            {{else}}
            // External spec file
            const spec = "{{.SpecURL}}";
            {{end}}

            const ui = SwaggerUIBundle({
                {{if .SpecJSON}}
                spec: spec,
                {{else}}
                url: spec,
                {{end}}
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch', 'head', 'options'],
                onComplete: function() {
                    console.log("Swagger UI loaded successfully");
                },
                onFailure: function(data) {
                    console.error("Failed to load Swagger UI", data);
                }
            });

            window.ui = ui;
        };
    </script>
</body>
</html>`

// GenerateMarkdownDocs generates markdown documentation
func (w *Writer) GenerateMarkdownDocs() error {
	spec, err := w.generator.GenerateSpec()
	if err != nil {
		return fmt.Errorf("failed to generate spec: %w", err)
	}

	var md strings.Builder

	// Title and description
	md.WriteString(fmt.Sprintf("# %s\n\n", spec.Info.Title))
	if spec.Info.Description != "" {
		md.WriteString(fmt.Sprintf("%s\n\n", spec.Info.Description))
	}

	// Version and other info
	md.WriteString(fmt.Sprintf("**Version:** %s\n\n", spec.Info.Version))
	if spec.Info.Contact != nil {
		md.WriteString("## Contact\n\n")
		if spec.Info.Contact.Name != "" {
			md.WriteString(fmt.Sprintf("**Name:** %s\n", spec.Info.Contact.Name))
		}
		if spec.Info.Contact.Email != "" {
			md.WriteString(fmt.Sprintf("**Email:** %s\n", spec.Info.Contact.Email))
		}
		if spec.Info.Contact.URL != "" {
			md.WriteString(fmt.Sprintf("**URL:** %s\n", spec.Info.Contact.URL))
		}
		md.WriteString("\n")
	}

	// Servers
	if len(spec.Servers) > 0 {
		md.WriteString("## Servers\n\n")
		for _, server := range spec.Servers {
			md.WriteString(fmt.Sprintf("- **%s**", server.URL))
			if server.Description != "" {
				md.WriteString(fmt.Sprintf(" - %s", server.Description))
			}
			md.WriteString("\n")
		}
		md.WriteString("\n")
	}

	// Endpoints
	md.WriteString("## Endpoints\n\n")
	for path, pathItem := range spec.Paths {
		md.WriteString(fmt.Sprintf("### %s\n\n", path))

		operations := map[string]*Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"PATCH":   pathItem.Patch,
			"DELETE":  pathItem.Delete,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
			"TRACE":   pathItem.Trace,
		}

		for method, op := range operations {
			if op != nil {
				md.WriteString(fmt.Sprintf("#### %s\n\n", method))
				if op.Summary != "" {
					md.WriteString(fmt.Sprintf("**Summary:** %s\n\n", op.Summary))
				}
				if op.Description != "" {
					md.WriteString(fmt.Sprintf("**Description:** %s\n\n", op.Description))
				}

				// Parameters
				if len(op.Parameters) > 0 {
					md.WriteString("**Parameters:**\n\n")
					for _, param := range op.Parameters {
						md.WriteString(fmt.Sprintf("- `%s` (%s)", param.Name, param.In))
						if param.Required {
							md.WriteString(" *required*")
						}
						if param.Description != "" {
							md.WriteString(fmt.Sprintf(" - %s", param.Description))
						}
						md.WriteString("\n")
					}
					md.WriteString("\n")
				}

				// Responses
				if len(op.Responses) > 0 {
					md.WriteString("**Responses:**\n\n")
					for code, response := range op.Responses {
						md.WriteString(fmt.Sprintf("- `%s` - %s\n", code, response.Description))
					}
					md.WriteString("\n")
				}
			}
		}
	}

	// Write to file
	filename := w.generator.GetOutputPath("README.md")
	return os.WriteFile(filename, []byte(md.String()), 0644)
}

// WatchAndRegenerate watches for changes and regenerates documentation
func (w *Writer) WatchAndRegenerate(interval time.Duration, callback func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if callback != nil {
			if err := callback(); err != nil {
				fmt.Printf("Error in watch callback: %v\n", err)
				continue
			}
		}

		if err := w.WriteFiles(); err != nil {
			fmt.Printf("Error regenerating documentation: %v\n", err)
		} else {
			fmt.Printf("Documentation regenerated at %s\n", time.Now().Format(time.RFC3339))
		}
	}
}
