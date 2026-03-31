package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

// validatePluginPath performs security validation on plugin file paths
func validatePluginPath(path string) error {
	// Prevent directory traversal attacks
	if strings.Contains(path, "..") {
		return errors.New("path traversal detected")
	}

	// Check file extension
	ext := filepath.Ext(path)
	if ext != ".go" {
		return fmt.Errorf("only .go files are supported, got: %s", ext)
	}

	// Check if file exists and is a regular file
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat plugin file: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("plugin path must be a regular file: %s", path)
	}

	return nil
}

// validateGoSource performs basic security validation on Go source code
func validateGoSource(source string) error {
	// Define patterns that could be dangerous in plugins
	dangerousPatterns := []struct {
		pattern string
		reason  string
	}{
		{`os\.Exec`, "use of os.Exec (potentially dangerous)"},
		{`exec\.Command`, "use of exec.Command (potentially dangerous)"},
		{`syscall\.Exec`, "use of syscall.Exec (potentially dangerous)"},
		{`http\.ListenAndServe`, "use of http.ListenAndServe (should be handled by framework)"},
		{`net\.Dial`, "use of net.Dial (uncontrolled network access)"},
		{`os\.Remove`, "use of os.Remove (file deletion)"},
		{`os\.RemoveAll`, "use of os.RemoveAll (directory deletion)"},
		{`plugin\.Open`, "recursive plugin loading (security risk)"},
		{`runtime\.GC`, "manual garbage collection manipulation"},
		{`debug\.SetGCPercent`, "garbage collection manipulation"},
		{`os\.Exit`, "use of os.Exit (can crash application)"},
		{`log\.Fatal`, "use of log.Fatal (can crash application)"},
	}

	// Check for dangerous patterns
	for _, dp := range dangerousPatterns {
		matched, err := regexp.MatchString(dp.pattern, source)
		if err != nil {
			return fmt.Errorf("invalid pattern in security validation: %w", err)
		}
		if matched {
			return fmt.Errorf("dangerous code pattern detected: %s (%s)", dp.pattern, dp.reason)
		}
	}

	return nil
}

// validatePluginInterface ensures that plugin implements required interface methods
func validatePluginInterface(plugin core.Plugin) error {
	// Since we're using ConfigurablePlugin, it already implements the interface
	// Just basic checks for Phase 1
	if plugin.Name() == "" {
		return errors.New("plugin name cannot be empty")
	}

	if plugin.Version() == "" {
		return errors.New("plugin version cannot be empty")
	}

	return nil
}
