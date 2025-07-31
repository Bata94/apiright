package ar_templ

var (
	config Config
)

type Config struct {
	htmx bool
	htmxFetchUrl string
	htmxLinkUrl string

	// TODO: Add DaisyUI
	// TODO: Add TailwindCSS
	// TODO: Add AlpineJS
	// TODO: Add TemplUI
}

func init() {
	config.htmx = true;
	config.htmxFetchUrl = "https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js";
	config.htmxLinkUrl = "";
}
