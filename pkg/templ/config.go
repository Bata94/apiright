package ar_templ

var (
	config Config
)

type Config struct {
	htmx bool
	htmxFetchUrl string
	htmxLinkUrl string
}

func init() {
	config.htmx = true;
	config.htmxFetchUrl = "https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js";
	config.htmxLinkUrl = "";
}
