package static

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
)

var log logger.Logger = logger.GetLogger()

// StaticSevFileConfig holds the configuration for serving a static file.
type StaticSevFileConfig struct {
	preLoad         bool
	preCache        bool
	maxPreCacheSize int64
	forcePreCache   bool
	contentType     string
	indexFile       string
	Title           string
	// TODO: Add more ConfigOptions
	// disableFileExplorer bool
	// includeHiddenFiles bool
	// excludeFiles []strings // as regex
	// excludeDirs []strings // as regex
	// customCssFile string
	// TODO: add compression on startup or on the fly based on "preLoaded"
	// compress bool
	// compressLevel int
	// compressType string
}

// StaticServFileOption is a function that configures a StaticSevFileConfig.
type StaticServFileOption func(*StaticSevFileConfig)

// NewStaticServeFileConfig creates a new StaticSevFileConfig.
func NewStaticServeFileConfig(opts ...StaticServFileOption) *StaticSevFileConfig {
	c := &StaticSevFileConfig{
		preLoad:         true,
		preCache:        true,
		maxPreCacheSize: 1024 * 1024 * 10, // 10 MB
		forcePreCache:   false,
		contentType:     "",
		indexFile:       "index.html",
		Title:           "ApiRight",
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.preCache && !c.preLoad {
		log.Fatal("Cannot set PreCache without PreLoad!")
	}

	if c.forcePreCache {
		log.Debug("Pre-caching files on startup.")
		c.preCache = true
	}

	return c
}

// WithoutPreLoad it will not look if files exits on startup! Be careful if using this option with a Directory, this function will expose all files in the directory, even if created after Server start!
// This will also disable pre-caching.
func WithoutPreLoad() StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.preLoad = false
		c.preCache = false
	}
}

// WithoutPreCache it will not cache files on startup.
func WithoutPreCache() StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.preCache = false
	}
}

// WithPreCache it will load files into memory on startup, even if bigger than the maxPreCacheSize.
// Caching will mainly reduce latency. With NVME drives, the speed difference is negligible.
func WithPreCache() StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.preCache = true
		c.forcePreCache = true
	}
}

// WithPreCacheSize sets the maximum size of the pre-cache.
func WithPreCacheSize(size int64) StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.preCache = true
		c.maxPreCacheSize = size
	}
}

// WithContentType sets the content type of the file.
func WithContentType(contentType string) StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.contentType = contentType
	}
}

// WithIndexFile sets the index file for the directory.
func WithIndexFile(indexFile string) StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.indexFile = indexFile
	}
}

// WithTitle sets the title for the directory listing page.
func WithTitle(title string) StaticServFileOption {
	return func(c *StaticSevFileConfig) {
		c.Title = title
	}
}

// loadFile loads a file from the given path.
// It returns the file, the file info, and an error if any.
// Remember to close the file after use, best use the closeFile function.
func loadFile(p string) (*os.File, os.FileInfo, error) {
	var (
		f     *os.File
		fInfo os.FileInfo
		err   error
	)
	fInfo, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = fmt.Errorf("static file '%s' does not exist. Please ensure the file exists", p)
		goto Return
	}

	f, err = os.Open(p)
	if err != nil {
		err = fmt.Errorf("static file '%s' exists, but is not readable: %w", p, err)
		goto Return
	}

Return:
	return f, fInfo, err
}

func closeFile(f *os.File) {
	if f == nil {
		return
	}
	err := f.Close()
	if err != nil {
		log.Panic("Failed to close file: ", err)
	}
}

// ServeStaticFile serves a static file.
func ServeStaticFile(urlPath, filePath string, opts ...StaticServFileOption) core.Handler {
	var h core.Handler
	config := NewStaticServeFileConfig(opts...)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("Error resolving absolute path for static file %s: %v", filePath, err)
	}

	if config.preLoad {
		log.Debugf("Attempting to serve static file from absolute path: %s", absFilePath)

		f, fInfo, err := loadFile(absFilePath)
		defer closeFile(f)
		if err != nil {
			log.Fatal("static file '%s' does not exist. Please ensure the file exists", absFilePath)
		}

		if config.preCache {
			if config.forcePreCache || fInfo.Size() <= config.maxPreCacheSize {
				log.Debug("Pre-caching file: ", absFilePath)
				content, err := io.ReadAll(f)
				if err != nil {
					log.Error("static file '%s' exists, but is not readable: %w", absFilePath, err)
				}

				h = func(c *core.Ctx) error {
					c.Response.SetStatus(200)
					c.Response.SetData(content)
					c.Response.AddHeader("Content-Type", config.contentType)
					return nil
				}
			}
		}

		if h == nil {
			h = func(c *core.Ctx) error {
				f, err := os.Open(absFilePath)
				defer closeFile(f)
				if err != nil {
					c.Response.SetStatus(404)
					c.Response.SetMessage("File not found... Please double-check the URL and try again.")
					return nil
				}

				content, err := io.ReadAll(f)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						c.Response.SetStatus(404)
						c.Response.SetMessage("File not found... Please double-check the URL and try again.")
						return nil
					}
					err = fmt.Errorf("static file '%s' exists, but is not readable: %w", absFilePath, err)
					log.Error(err)

					c.Response.SetStatus(500)
					c.Response.SetMessage("File was found, but is not readable... Please try again later, if it persists, contact the administrator.")
					return nil
				}

				c.Response.SetStatus(200)
				c.Response.SetData(content)
				c.Response.AddHeader("Content-Type", config.contentType)
				return nil
			}
		}
	} else {
		h = func(c *core.Ctx) error {
			f, _, err := loadFile(absFilePath)
			defer closeFile(f)
			if err != nil {
				c.Response.SetStatus(404)
				c.Response.SetMessage("File not found... Please double-check the URL and try again.")
				return nil
			}

			content, err := io.ReadAll(f)
			if err != nil {
				err = fmt.Errorf("static file '%s' exists, but is not readable: %w", absFilePath, err)
				log.Error(err)

				c.Response.SetStatus(500)
				c.Response.SetMessage("File was found, but is not readable... Please try again later, if it persists, contact the administrator.")
				return nil
			}

			c.Response.SetStatus(200)
			c.Response.SetData(content)
			c.Response.AddHeader("Content-Type", config.contentType)
			return nil
		}
	}
	return h
}

// DirTemplateData holds the data for the directory template
type DirTemplateData struct {
	Title           string
	BaseUrl         string
	IndexFileExists bool
	Files           []FileData
	Dirs            []DirData
}

// FileData holds the data for a file
type FileData struct {
	DirPath string
	Name    string
	Size    int64
}

// DirData holds the data for a directory
type DirData struct {
	Name string
}

// setupTemplateFuncs creates and returns a template.FuncMap
// containing our custom functions.
func setupTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"hasSuffix": func(s, suffix string) bool {
			return strings.HasSuffix(s, suffix)
		},
		"byteFormat": func(b int64) string {
			const unit = 1024
			if b < unit {
				return fmt.Sprintf("%d B", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			// KMGTPE represents Kilo, Mega, Giga, Tera, Peta, Exa
			return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
		},
	}
}

// ServeStaticDir serves a directory at the given URL path.
// If the directory contains a "index.html" file, it will be served as "/" route. If not a FileExplorer will be shown, by default.
// It is highly recommended to leave the PreLoad option enabled.
func ServeStaticDir(urlPath, dirPath string, opts ...StaticServFileOption) (core.Handler, core.Handler) {
	var h core.Handler
	var redirectHandler core.Handler
	log.Debug("📁 Serving static directory: ", dirPath, " at: ", urlPath)
	config := NewStaticServeFileConfig(opts...)

	dirTempl, err := template.New(urlPath).Funcs(setupTemplateFuncs()).Parse(dirExplorerTemplate)
	if err != nil {
		log.Fatal(err)
	}

	pattern := urlPath
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	if config.preLoad {
		dirData, err := getStaticDirData(dirPath, dirPath, pattern, config)
		if err != nil {
			log.Fatal(err)
		}

		if config.forcePreCache || config.preCache {
			content := decideStaticIndexFile(dirData, dirPath, dirTempl, config)
			h = func(c *core.Ctx) error {
				c.Response.SetStatus(200)
				c.Response.SetData(content)
				c.Response.AddHeader("Content-Type", "text/html")

				return nil
			}
		} else {
			h = func(c *core.Ctx) error {
				content := decideStaticIndexFile(dirData, dirPath, dirTempl, config)

				c.Response.SetStatus(200)
				c.Response.SetData(content)
				c.Response.AddHeader("Content-Type", "text/html")

				return nil
			}
		}
	} else {
		h = func(c *core.Ctx) error {
			var (
				content []byte
				err     error
			)

			if strings.HasSuffix(c.Request.URL.Path, "/") {
				var dirData DirTemplateData
				dirData, err = getStaticDirData(dirPath, dirPath+strings.TrimPrefix(c.Request.URL.Path, urlPath+"/"), pattern, config)
				if err != nil {
					log.Error(err)
					return err
				}
				content = decideStaticIndexFile(dirData, dirPath+strings.TrimPrefix(c.Request.URL.Path, urlPath+"/"), dirTempl, config)
			} else {
				log.Debug("📁 Serving static directory: ", dirPath, " at: ", urlPath)
				log.Debugf("Current URLPath %s", c.Request.URL.Path)

				content, err = os.ReadFile(filepath.Join(dirPath, strings.TrimPrefix(c.Request.URL.Path, urlPath+"/")))
				if err != nil {
					log.Error(err)
					if errors.Is(err, os.ErrNotExist) {
						c.Response.SetStatus(404)
						c.Response.SetMessage("File not found... Please double-check the URL and try again.")
						return nil
					}
					return err
				}
			}

			c.Response.SetStatus(200)
			c.Response.SetData(content)
			c.Response.AddHeader("Content-Type", "text/html")

			return err
		}
	}

	// Create redirect handler for urlPath to urlPath+"/"
	redirectHandler = func(c *core.Ctx) error {
		c.Response.Redirect(urlPath+"/", http.StatusMovedPermanently)
		return nil
	}

	return h, redirectHandler
}

func getStaticDirData(baseDirPath, dirPath, baseUrl string, config *StaticSevFileConfig) (DirTemplateData, error) {
	dirData := DirTemplateData{
		Title:           config.Title,
		BaseUrl:         baseUrl,
		IndexFileExists: false,
		Files:           []FileData{},
		Dirs:            []DirData{},
	}
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return dirData, err
	}

	if len(files) == 0 {
		return dirData, fmt.Errorf("directory is empty: %s", dirPath)
	}

	for _, file := range files {
		if file.Name() == config.indexFile {
			dirData.IndexFileExists = true
			break
		}
	}

	for _, file := range files {
		if file.IsDir() {
			dirData.Dirs = append(dirData.Dirs, DirData{
				Name: file.Name(),
			})
			continue
		}
		fInfo, err := file.Info()
		if err != nil {
			return dirData, err
		}
		size := fInfo.Size()
		fd := FileData{
			DirPath: strings.TrimPrefix(dirPath, baseDirPath),
			Name:    file.Name(),
			Size:    size,
		}

		if !strings.HasSuffix(fd.DirPath, "/") && fd.DirPath != "" {
			fd.DirPath += "/"
		}
		dirData.Files = append(dirData.Files, fd)
	}

	return dirData, nil
}

func decideStaticIndexFile(dirData DirTemplateData, dirPath string, dirTempl *template.Template, config *StaticSevFileConfig) []byte {
	buf := new(bytes.Buffer)
	if dirData.IndexFileExists {
		indexFileContent, err := os.ReadFile(filepath.Join(dirPath, config.indexFile))
		if err != nil {
			log.Fatal(err)
		}
		buf.Write(indexFileContent)
	} else {
		if err := dirTempl.Execute(buf, dirData); err != nil {
			log.Fatal(err)
		}
	}
	return buf.Bytes()
}

// TODO: Better Breadcrumbs
// TODO: Move out css to a separate file
const dirExplorerTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>{{.Title}} - File Explorer</title>
    <link rel="icon" href="./favicon.ico" type="image/x-icon" />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0-beta3/css/all.min.css"
    />
    <style>
      :root {
        --primary-color: #007bff;
        --secondary-color: #6c757d;
        --background-color: #f8f9fa;
        --card-background: #ffffff;
        --border-color: #dee2e6;
        --text-color: #343a40;
        --link-color: #007bff;
        --link-hover-color: #0056b3;
        --header-bg: #e9ecef;
      }

      body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
          "Helvetica Neue", Arial, sans-serif, "Apple Color Emoji",
          "Segoe UI Emoji", "Segoe UI Symbol";
        line-height: 1.6;
        color: var(--text-color);
        background-color: var(--background-color);
        margin: 0;
        padding: 0;
      }

      .container {
        max-width: 960px;
        margin: 20px auto;
        padding: 0 15px;
      }

      header {
        background-color: var(--header-bg);
        padding: 20px 0;
        border-bottom: 1px solid var(--border-color);
        text-align: center;
      }

      header h1 {
        margin: 0;
        color: var(--primary-color);
      }

      .file-list {
        list-style: none;
        padding: 0;
        margin-top: 20px;
        background-color: var(--card-background);
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
      }

      .file-item {
        display: flex;
        align-items: center;
        padding: 12px 20px;
        border-bottom: 1px solid var(--border-color);
        transition: background-color 0.2s ease;
      }

      .file-item:last-child {
        border-bottom: none;
      }

      .file-item:hover {
        background-color: #f1f3f5;
      }

      .file-item a {
        text-decoration: none;
        color: var(--link-color);
        display: flex;
        align-items: center;
        flex-grow: 1;
      }

      .file-item a:hover {
        color: var(--link-hover-color);
      }

      .file-icon {
        width: 24px;
        text-align: center;
        margin-right: 15px;
        color: var(--secondary-color);
      }

      .file-icon.folder {
        color: #ffc107; /* Yellow for folders */
      }

      .file-name {
        flex-grow: 1;
        font-weight: 500;
      }

      .file-size {
        font-size: 0.9em;
        color: var(--secondary-color);
        margin-left: 15px;
      }

      .path-breadcrumb {
        margin-top: 15px;
        padding: 10px 20px;
        background-color: var(--header-bg);
        border-radius: 5px;
        font-size: 0.9em;
      }

      .path-breadcrumb a {
        text-decoration: none;
        color: var(--primary-color);
      }

      .path-breadcrumb span {
        color: var(--secondary-color);
      }

      .path-breadcrumb a:hover {
        text-decoration: underline;
      }

      footer {
        text-align: center;
        margin-top: 40px;
        padding: 20px;
        color: var(--secondary-color);
        font-size: 0.8em;
      }

      @media (max-width: 768px) {
        .file-item {
          flex-wrap: wrap;
        }
        .file-size {
          margin-left: 0;
          width: 100%;
          text-align: right;
          font-size: 0.85em;
          padding-top: 5px;
        }
        .file-name {
          flex-basis: calc(100% - 39px); /* icon width + margin */
        }
      }
    </style>
  </head>
  <body>
    <header>
      <div class="container">
        <h1><i class="fas fa-folder-open"></i> {{.Title}}</h1>
      </div>
    </header>
    <main class="container">
      <nav class="path-breadcrumb">
        <!-- Example Breadcrumb (You'll need to pass this data from Go) -->
        <!-- For now, assuming Title is the current path -->
        You are in:
        <a href="{{.BaseUrl}}">Home</a>
        {{- if ne .Title "/"}}
          <span> / </span>
          <span>{{.Title}}</span>
        {{- end}}
      </nav>

      <ul class="file-list">
				{{range .Dirs}}
					<li class="file-item">
						<a href="{{$.BaseUrl}}{{.Name}}/">
							<span class="file-icon">
								<i class="fas fa-folder folder"></i>
							</span>
							<span class="file-name">{{.Name}}</span>
						</a>
					</li>
				{{end}}

        {{range .Files}}
          <li class="file-item">
            <a href="{{$.BaseUrl}}{{.DirPath}}{{.Name}}">
              <span class="file-icon">
                {{if (hasSuffix .Name ".pdf")}}
                  <i class="fas fa-file-pdf"></i>
                {{else if (or (hasSuffix .Name ".png") (hasSuffix .Name ".jpg") (hasSuffix .Name ".jpeg") (hasSuffix .Name ".gif"))}}
                  <i class="fas fa-file-image"></i>
                {{else if (or (hasSuffix .Name ".zip") (hasSuffix .Name ".rar") (hasSuffix .Name ".7z"))}}
                  <i class="fas fa-file-archive"></i>
                {{else if (or (hasSuffix .Name ".doc") (hasSuffix .Name ".docx"))}}
                  <i class="fas fa-file-word"></i>
                {{else if (or (hasSuffix .Name ".xls") (hasSuffix .Name ".xlsx"))}}
                  <i class="fas fa-file-excel"></i>
                {{else if (or (hasSuffix .Name ".ppt") (hasSuffix .Name ".pptx"))}}
                  <i class="fas fa-file-powerpoint"></i>
                {{else if (or (hasSuffix .Name ".txt") (hasSuffix .Name ".md") (hasSuffix .Name ".go") (hasSuffix .Name ".html") (hasSuffix .Name ".css") (hasSuffix .Name ".js") (hasSuffix .Name ".json") (hasSuffix .Name ".xml"))}}
                  <i class="fas fa-file-code"></i>
                {{else if (or (hasSuffix .Name ".mp3") (hasSuffix .Name ".wav"))}}
                  <i class="fas fa-file-audio"></i>
                {{else if (or (hasSuffix .Name ".mp4") (hasSuffix .Name ".mov") (hasSuffix .Name ".avi"))}}
                  <i class="fas fa-file-video"></i>
                {{else}}
                  <i class="fas fa-file"></i>
                {{end}}
              </span>
              <span class="file-name">{{.Name}}</span>
              <span class="file-size">
								{{.Size | byteFormat}}
              </span>
            </a>
          </li>
        {{end}}
      </ul>
			{{if and (eq (len .Files) 0) (eq (len .Dirs) 0)}}
        <p style="text-align: center; color: var(--secondary-color);">
          This directory is empty.
        </p>
      {{end}}
    </main>
    <footer>
      <p>ApiRight File Explorer</p>
    </footer>
  </body>
</html>
`
