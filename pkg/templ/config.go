package ar_templ

var (
	config Config
)

type Config struct {
	htmx         bool
	htmxFetchUrl string
	htmxLinkUrl  string

	// TODO: Add DaisyUI
	// TODO: Add TailwindCSS
	// TODO: Add AlpineJS
	// TODO: Add TemplUI
}

// TODO: make sure this is not overwritten in the lifetime of the app
func init() {
	config.htmx = false
	config.htmxFetchUrl = "https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js"
	config.htmxLinkUrl = ""
}
