package plugins

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

// SourcePluginLoader handles loading of Go source file plugins
type SourcePluginLoader struct {
	registry *PluginRegistry
	logger   core.Logger
}

// NewSourcePluginLoader creates a new source plugin loader
func NewSourcePluginLoader(registry *PluginRegistry, logger core.Logger) *SourcePluginLoader {
	return &SourcePluginLoader{
		registry: registry,
		logger:   logger,
	}
}

// LoadFromFile loads a plugin from a Go source file
func (spl *SourcePluginLoader) LoadFromFile(path string) error {
	if spl.logger != nil {
		spl.logger.Info("Loading Go source plugin", "path", path)
	}

	// 1. Security validation
	if err := validatePluginPath(path); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	// 2. Read and validate source code
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}

	if err := validateGoSource(string(source)); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	// 3. Parse and instantiate plugin
	plugin, err := spl.parseAndInstantiatePlugin(string(source), path)
	if err != nil {
		return fmt.Errorf("failed to parse plugin: %w", err)
	}

	// Debug: check what plugin we got
	if spl.logger != nil {
		spl.logger.Info("Parsed plugin", "name", plugin.Name(), "version", plugin.Version())
	}

	// 4. Validate plugin interface
	if err := validatePluginInterface(plugin); err != nil {
		return fmt.Errorf("plugin interface validation failed: %w", err)
	}

	// 5. Register plugin
	if err := spl.registry.RegisterPlugin(plugin); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	if spl.logger != nil {
		spl.logger.Info("Successfully loaded plugin", "name", plugin.Name(), "path", path)
	}

	return nil
}

// parseAndInstantiatePlugin parses Go source and creates plugin instance
func (spl *SourcePluginLoader) parseAndInstantiatePlugin(source, path string) (core.Plugin, error) {
	// Parse the source file
	_, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go source: %w", err)
	}

	// Extract plugin information using simple regex patterns
	// This is a simplified approach for Phase 1
	pluginInfo, err := spl.extractPluginInfo(source, path)
	if err != nil {
		return nil, fmt.Errorf("failed to extract plugin info: %w", err)
	}

	// Create a configurable plugin instance with extracted info
	plugin := &ConfigurablePlugin{
		name:    pluginInfo.name,
		version: pluginInfo.version,
		config:  make(map[string]any),
	}

	return plugin, nil
}

// pluginInfo holds extracted plugin information
type pluginInfo struct {
	name    string
	version string
}

// extractPluginInfo extracts plugin name and version from source code
func (spl *SourcePluginLoader) extractPluginInfo(source, path string) (*pluginInfo, error) {
	info := &pluginInfo{
		name:    "unknown",
		version: "1.0.0",
	}

	// Extract package name
	pkgMatch := regexp.MustCompile(`package\s+(\w+)`).FindStringSubmatch(source)
	if len(pkgMatch) >= 2 {
		// Use filename as fallback for plugin name
		filename := filepath.Base(path)
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		info.name = toCamelCase(nameWithoutExt)
	}

	// Extract version from constant or comment
	versionPatterns := []string{
		`const\s+Version\s*=\s*"([^"]+)"`,
		`var\s+Version\s*=\s*"([^"]+)"`,
		`//\s*version:\s*([^\s]+)`,
		`//\s*@version\s+([^\s]+)`,
	}

	for _, pattern := range versionPatterns {
		if match := regexp.MustCompile(pattern).FindStringSubmatch(source); len(match) >= 2 {
			info.version = strings.TrimSpace(match[1])
			break
		}
	}

	// Look for plugin struct to get name
	structPattern := regexp.MustCompile(`type\s+(\w+Plugin)\s+struct`)
	if match := structPattern.FindStringSubmatch(source); len(match) >= 2 {
		info.name = match[1]
	}

	// Look for function that returns plugin
	funcPattern := regexp.MustCompile(`func\s+(\w+)\(\).*Plugin\s*{`)
	if match := funcPattern.FindStringSubmatch(source); len(match) >= 2 {
		info.name = match[1]
	}

	return info, nil
}

// toCamelCase converts string to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for i, part := range parts {
		if i > 0 && len(part) > 0 {
			part = strings.ToUpper(string(part[0])) + part[1:]
		}
		result += part
	}
	return result
}
