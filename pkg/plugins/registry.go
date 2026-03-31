package plugins

import (
	"fmt"
	"sync"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
)

// PluginRegistry manages plugin registration and lifecycle
type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]core.Plugin
	hooks   map[string][]PluginHook
}

// PluginHook represents a plugin hook point
type PluginHook struct {
	Name     string
	Priority int
	Func     func(ctx *core.GenerationContext) error
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]core.Plugin),
		hooks:   make(map[string][]PluginHook),
	}
}

// RegisterPlugin registers a plugin
func (pr *PluginRegistry) RegisterPlugin(plugin core.Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	name := plugin.Name()
	if _, exists := pr.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	pr.plugins[name] = plugin
	return nil
}

// GetPlugin gets a plugin by name
func (pr *PluginRegistry) GetPlugin(name string) (core.Plugin, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugin, exists := pr.plugins[name]
	return plugin, exists
}

// ListPlugins returns all registered plugins
func (pr *PluginRegistry) ListPlugins() []core.Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]core.Plugin, 0, len(pr.plugins))
	for _, plugin := range pr.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// UnregisterPlugin unregisters a plugin
func (pr *PluginRegistry) UnregisterPlugin(name string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	delete(pr.plugins, name)
	return nil
}

// RegisterHook registers a plugin hook
func (pr *PluginRegistry) RegisterHook(hookName string, hook PluginHook) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if pr.hooks[hookName] == nil {
		pr.hooks[hookName] = make([]PluginHook, 0)
	}

	pr.hooks[hookName] = append(pr.hooks[hookName], hook)
}

// ExecuteHooks executes all hooks for a given hook name
func (pr *PluginRegistry) ExecuteHooks(hookName string, ctx *core.GenerationContext) error {
	pr.mu.RLock()
	hooks := pr.hooks[hookName]
	pr.mu.RUnlock()

	// Sort hooks by priority (lower number = higher priority)
	sortedHooks := make([]PluginHook, len(hooks))
	copy(sortedHooks, hooks)

	// Simple bubble sort for now - could be optimized
	for i := 0; i < len(sortedHooks)-1; i++ {
		for j := 0; j < len(sortedHooks)-i-1; j++ {
			if sortedHooks[j].Priority > sortedHooks[j+1].Priority {
				sortedHooks[j], sortedHooks[j+1] = sortedHooks[j+1], sortedHooks[j]
			}
		}
	}

	for _, hook := range sortedHooks {
		if err := hook.Func(ctx); err != nil {
			return fmt.Errorf("plugin hook %s failed: %w", hook.Name, err)
		}
	}

	return nil
}

// ValidatePlugins validates all registered plugins against a schema
func (pr *PluginRegistry) ValidatePlugins(schema *core.Schema) error {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	for _, plugin := range pr.plugins {
		if err := plugin.Validate(schema); err != nil {
			return fmt.Errorf("plugin %s validation failed: %w", plugin.Name(), err)
		}
	}

	return nil
}

// GeneratePlugins runs the generation step for all registered plugins
func (pr *PluginRegistry) GeneratePlugins(ctx *core.GenerationContext) error {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	for _, plugin := range pr.plugins {
		if err := plugin.Generate(ctx); err != nil {
			return fmt.Errorf("plugin %s generation failed: %w", plugin.Name(), err)
		}
	}

	return nil
}

// GetProtoExtensions returns all plugins that implement ProtoExtension interface
func (pr *PluginRegistry) GetProtoExtensions() []core.ProtoExtension {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var extensions []core.ProtoExtension
	for _, plugin := range pr.plugins {
		if ext, ok := plugin.(core.ProtoExtension); ok {
			extensions = append(extensions, ext)
		}
	}
	return extensions
}

// GetMiddlewareProviders returns all plugins that implement MiddlewareProvider interface
func (pr *PluginRegistry) GetMiddlewareProviders() []core.MiddlewareProvider {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var providers []core.MiddlewareProvider
	for _, plugin := range pr.plugins {
		if mp, ok := plugin.(core.MiddlewareProvider); ok {
			providers = append(providers, mp)
		}
	}
	return providers
}

// PluginLoader handles loading plugins from various sources
type PluginLoader struct {
	registry *PluginRegistry
}

// simpleLogger provides basic logging for plugin loading
type simpleLogger struct {
	verbose bool
}

func (l *simpleLogger) Debug(msg string, fields ...interface{}) {
	if l.verbose {
		fmt.Printf("[PLUGIN-DEBUG] %s %v\n", msg, fields)
	}
}

func (l *simpleLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN] %s %v\n", msg, fields)
}

func (l *simpleLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN-WARN] %s %v\n", msg, fields)
}

func (l *simpleLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN-ERROR] %s %v\n", msg, fields)
}

func (l *simpleLogger) DPanic(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN-DPANIC] %s %v\n", msg, fields)
}

func (l *simpleLogger) Panic(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN-PANIC] %s %v\n", msg, fields)
}

func (l *simpleLogger) Fatal(msg string, fields ...interface{}) {
	fmt.Printf("[PLUGIN-FATAL] %s %v\n", msg, fields)
}

func (l *simpleLogger) With(fields ...interface{}) core.Logger {
	return l
}

func (l *simpleLogger) Sync() error {
	return nil
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(registry *PluginRegistry) *PluginLoader {
	return &PluginLoader{
		registry: registry,
	}
}

// LoadFromFile loads a plugin from a file
func (pl *PluginLoader) LoadFromFile(path string) error {
	sourceLoader := NewSourcePluginLoader(pl.registry, nil)
	return sourceLoader.LoadFromFile(path)
}

// LoadFromConfig loads plugins from configuration
func (pl *PluginLoader) LoadFromConfig(plugins []config.PluginConfig) error {
	for _, pluginConfig := range plugins {
		if !pluginConfig.Enabled {
			continue
		}

		var plugin core.Plugin
		var err error

		if pluginConfig.Path != "" {
			// Load from file path
			sourceLoader := NewSourcePluginLoader(pl.registry, &simpleLogger{verbose: false})
			err = sourceLoader.LoadFromFile(pluginConfig.Path)
			if err != nil {
				return fmt.Errorf("failed to load plugin from path %s: %w", pluginConfig.Path, err)
			}
			continue
		}

		// Fallback to configurable plugin
		plugin = &ConfigurablePlugin{
			name:    pluginConfig.Name,
			version: pluginConfig.Version,
			config:  pluginConfig.Config,
		}

		if err := pl.registry.RegisterPlugin(plugin); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", pluginConfig.Name, err)
		}
	}

	return nil
}

// ConfigurablePlugin is a plugin configured from configuration
type ConfigurablePlugin struct {
	name    string
	version string
	config  map[string]interface{}
}

// Name returns the plugin name
func (cp *ConfigurablePlugin) Name() string {
	return cp.name
}

// Version returns the plugin version
func (cp *ConfigurablePlugin) Version() string {
	return cp.version
}

// Generate runs the plugin generation
func (cp *ConfigurablePlugin) Generate(ctx *core.GenerationContext) error {
	// TODO: Implement configurable plugin generation
	ctx.Log("info", "Generating with plugin %s", cp.name)
	return nil
}

// Validate validates the plugin
func (cp *ConfigurablePlugin) Validate(schema *core.Schema) error {
	// TODO: Implement configurable plugin validation
	return nil
}

// Global registry instance
var globalRegistry = NewPluginRegistry()

// GetGlobalRegistry returns the global plugin registry
func GetGlobalRegistry() *PluginRegistry {
	return globalRegistry
}

// RegisterPlugin registers a plugin globally
func RegisterPlugin(plugin core.Plugin) error {
	return globalRegistry.RegisterPlugin(plugin)
}

// GetPlugin gets a plugin by name globally
func GetPlugin(name string) (core.Plugin, bool) {
	return globalRegistry.GetPlugin(name)
}

// ListPlugins returns all registered plugins globally
func ListPlugins() []core.Plugin {
	return globalRegistry.ListPlugins()
}

// UnregisterPlugin unregisters a plugin globally
func UnregisterPlugin(name string) error {
	return globalRegistry.UnregisterPlugin(name)
}

// RegisterHook registers a plugin hook globally
func RegisterHook(hookName string, hook PluginHook) {
	globalRegistry.RegisterHook(hookName, hook)
}

// ExecuteHooks executes all hooks for a given hook name globally
func ExecuteHooks(hookName string, ctx *core.GenerationContext) error {
	return globalRegistry.ExecuteHooks(hookName, ctx)
}

// ValidatePlugins validates all registered plugins globally
func ValidatePlugins(schema *core.Schema) error {
	return globalRegistry.ValidatePlugins(schema)
}

// GeneratePlugins runs generation for all registered plugins globally
func GeneratePlugins(ctx *core.GenerationContext) error {
	return globalRegistry.GeneratePlugins(ctx)
}

// GetProtoExtensions returns all plugins that implement ProtoExtension interface
func GetProtoExtensions() []core.ProtoExtension {
	var extensions []core.ProtoExtension
	for _, plugin := range globalRegistry.ListPlugins() {
		if ext, ok := plugin.(core.ProtoExtension); ok {
			extensions = append(extensions, ext)
		}
	}
	return extensions
}

// GetMiddlewareProviders returns all plugins that implement MiddlewareProvider interface
func GetMiddlewareProviders() []core.MiddlewareProvider {
	var providers []core.MiddlewareProvider
	for _, plugin := range globalRegistry.ListPlugins() {
		if mp, ok := plugin.(core.MiddlewareProvider); ok {
			providers = append(providers, mp)
		}
	}
	return providers
}

// HasProtoExtensions checks if any registered plugin provides proto extensions
func HasProtoExtensions() bool {
	return len(GetProtoExtensions()) > 0
}

// Hook points for plugins
const (
	HookBeforeGeneration = "before_generation"
	HookAfterGeneration  = "after_generation"
	HookBeforeSQLC       = "before_sqlc"
	HookAfterSQLC        = "after_sqlc"
	HookBeforeProtobuf   = "before_protobuf"
	HookAfterProtobuf    = "after_protobuf"
	HookBeforeServer     = "before_server"
	HookAfterServer      = "after_server"
)

// BasePlugin provides common functionality for plugins
type BasePlugin struct {
	NameVal    string
	VersionVal string
	Config     map[string]interface{}
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version string) *BasePlugin {
	return &BasePlugin{
		NameVal:    name,
		VersionVal: version,
		Config:     make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (bp *BasePlugin) Name() string {
	return bp.NameVal
}

// Version returns the plugin version
func (bp *BasePlugin) Version() string {
	return bp.VersionVal
}

// SetConfig sets the plugin configuration
func (bp *BasePlugin) SetConfig(config map[string]interface{}) {
	bp.Config = config
}

// GetConfig gets a configuration value
func (bp *BasePlugin) GetConfig(key string) (interface{}, bool) {
	value, exists := bp.Config[key]
	return value, exists
}

// GetConfigString gets a string configuration value
func (bp *BasePlugin) GetConfigString(key string) string {
	if value, exists := bp.GetConfig(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetConfigBool gets a boolean configuration value
func (bp *BasePlugin) GetConfigBool(key string) bool {
	if value, exists := bp.GetConfig(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}
