package ar_templ

var (
	config Config
)

type Config struct {
	htmx         bool
	htmxFetchUrl string
	htmxLinkUrl  string

	// TODO: Add configuration options and integration for popular UI frameworks like DaisyUI, TailwindCSS, AlpineJS, and TemplUI to streamline their usage within the application.
}

// TODO: Ensure that the 'config' variable, once initialized, is not inadvertently overwritten or modified during the application's lifetime to maintain consistent configuration.
func init() {
	config.htmx = false
	config.htmxFetchUrl = "https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js"
	config.htmxLinkUrl = ""
}
