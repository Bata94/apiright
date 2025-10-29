package ar_templ

import (
	"testing"
)

func TestConfig(t *testing.T) {
	// Test that config is initialized
	if config.htmx != false {
		t.Errorf("Expected htmx to be false, got %v", config.htmx)
	}
	if config.htmxFetchUrl != "https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js" {
		t.Errorf("Expected htmxFetchUrl to be correct URL, got %s", config.htmxFetchUrl)
	}
	if config.htmxLinkUrl != "" {
		t.Errorf("Expected htmxLinkUrl to be empty, got %s", config.htmxLinkUrl)
	}
}
