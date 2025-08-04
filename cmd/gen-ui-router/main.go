package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Route represents a single route with its path and component name.
type Route struct {
	Path        string
	ComponentName string
}

// TemplateData holds the data for the Go template.
type TemplateData struct {
	PackageName       string
	Routes            []Route
	ModulePath        string
	RelativeInputPath string
	ImportAlias       string
}

const (
	defaultOutputFileName = "routes_gen.go"
)

var (
	inputDir    string
	outputFile  string
	packageName string
	routeRegex  = regexp.MustCompile(`(?m)^//\s*ui-router:\s*(.*)$`)
)

func init() {
	flag.StringVar(&inputDir, "input", "", "Input directory containing .templ files (required)")
	flag.StringVar(&outputFile, "output", defaultOutputFileName, "Output file name for generated routes.go")
	flag.StringVar(&packageName, "package", "uirouter", "Package name for the generated routes.go file")
}

func main() {
	flag.Parse()

	if inputDir == "" {
		log.Fatal("Error: --input directory is required.")
	}

	// Determine the Go module path
	modulePath, err := getGoModulePath()
	if err != nil {
		log.Fatalf("Error determining Go module path: %v", err)
	}

	// Ensure inputDir is absolute
	absInputDir, err := filepath.Abs(inputDir)
	if err != nil {
		log.Fatalf("Error resolving absolute path for input directory %s: %v", inputDir, err)
	}
	inputDir = absInputDir

	projectRoot, err := getProjectRoot()
	if err != nil {
		log.Fatalf("Error determining project root: %v", err)
	}
	relInputDir, err := filepath.Rel(projectRoot, inputDir)
	if err != nil {
		log.Fatalf("Error getting relative path for input directory %s: %v", inputDir, err)
	}
	// Ensure relative path uses forward slashes for Go import paths
	relInputDir = filepath.ToSlash(relInputDir)

	importAlias := filepath.Base(relInputDir)
	importAlias = strings.ReplaceAll(importAlias, "-", "_") // Ensure valid Go identifier

	routes, err := findRoutes(inputDir)
	if err != nil {
		log.Fatalf("Error finding routes: %v", err)
	}

	if len(routes) == 0 {
		log.Printf("No routes found in %s. Skipping generation of %s.", inputDir, outputFile)
		return
	}

	data := TemplateData{
		PackageName:       packageName,
		Routes:            routes,
		ModulePath:        modulePath,
		RelativeInputPath: relInputDir,
		ImportAlias:       importAlias,
	}

	err = generateRoutesFile(outputFile, data)
	if err != nil {
		log.Fatalf("Error generating routes file: %v", err)
	}

	log.Printf("Successfully generated %s with %d routes.", outputFile, len(routes))
}

func getGoModulePath() (string, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return "", err
	}
	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}
	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		if after, ok :=strings.CutPrefix(line, "module "); ok  {
			return strings.TrimSpace(after), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}

func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	currentDir := cwd
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root directory
			break
		}
		currentDir = parentDir
	}
	return "", fmt.Errorf("go.mod not found in current directory or any parent directories")
}

func findRoutes(dir string) ([]Route, error) {
	var routes []Route
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".templ") {
			return nil
		}

		baseName := strings.TrimSuffix(filepath.Base(path), ".templ")
		routePath := "/" + baseName
		switch baseName {
		case "":
			log.Printf("Warning: Found empty component name in %s. Skipping.", path)
			return nil
		case "index", "root":
			routePath = "/"
		}

		if routePath == "" {
			log.Printf("Warning: Found empty route path in %s. Skipping.", path)
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		goPackageDir := filepath.Dir(relPath)
		goPackageName := ""
		if goPackageDir != "." {
			goPackageName = filepath.Base(goPackageDir)
			goPackageName = strings.ReplaceAll(goPackageName, "-", "_")
		}

		goFunctionName := strings.ReplaceAll(baseName, "-", "_")
		goFunctionName = strings.Title(goFunctionName)

		routes = append(routes, Route{
			Path:        routePath,
			ComponentName: goFunctionName,
		})
		return nil
	})
	return routes, err
}

func generateRoutesFile(outputFilePath string, data TemplateData) error {
	tmpl := template.Must(template.New("routes").Parse(routesTemplate))

	f, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputFilePath, err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

const routesTemplate = `// Code generated by go generate; DO NOT EDIT.
// This file was generated by cmd/gen-ui-router

package {{.PackageName}}

import (
	"net/http"
	"log"
	"context"

	"github.com/a-h/templ"
	{{.ImportAlias}} "{{.ModulePath}}/{{.RelativeInputPath}}"
)

var (
	_ context.Context = nil // Dummy usage to satisfy Go linter
	_ templ.Component = nil // Dummy usage to satisfy Go linter
)

// RegisterUIRoutes registers the UI routes with the given http.ServeMux.
func RegisterUIRoutes(mux *http.ServeMux) {
	log.Println("Registering UI routes...")
	{{$importAlias := .ImportAlias}}
	{{range .Routes}}
	mux.HandleFunc("{{.Path}}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		comp := {{$importAlias}}.{{.ComponentName}}()
		comp.Render(ctx, w)
	})
	{{end}}
	log.Println("UI routes registered.")
}
`
