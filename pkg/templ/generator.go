package ar_templ

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TODO: By Default all Routes are registerd to OpenApi Docs, maybe disable this by default and add a flag to enable it.
// TODO: Subdirectories are not supported yet.

// Route represents a single route with its path and component name.
type Route struct {
	Path          string
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

func GeneratorRun(inputDir, outputFile, packageName string) {
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

	// Ensure Output directory exists
	outputDir := filepath.Dir(outputFile)
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating output directory %s: %v", outputDir, err)
	}

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
		if after, ok := strings.CutPrefix(line, "module "); ok {
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

		goFunctionName := strings.ReplaceAll(baseName, "-", " ")
		goFunctionName = cases.Title(language.English, cases.NoLower).String(goFunctionName)
		goFunctionName = strings.ReplaceAll(goFunctionName, " ", "")

		routes = append(routes, Route{
			Path:          routePath,
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
	defer func() {
		if err := f.Close(); err != nil {
			log.Panicf("Error closing output file %s: %v", outputFilePath, err)
		}
	}()

	return tmpl.Execute(f, data)
}

const routesTemplate = `// Code generated by go generate; DO NOT EDIT.
// This file was generated by ApiRight
{{$importAlias := .ImportAlias}}
package {{.PackageName}}

import (
	"log"

	ar "github.com/bata94/apiright/pkg/core"
	ar_templ "github.com/bata94/apiright/pkg/templ"
	{{.ImportAlias}} "{{.ModulePath}}/{{.RelativeInputPath}}"
)

// RegisterUIRoutes registers the UI routes with the given http.ServeMux.
func RegisterUIRoutes(router *ar.Router) {
	log.Println("Registering UI routes...")
	{{range .Routes}}
	router.GET(ar_templ.SimpleRenderer("{{.Path}}", {{$importAlias}}.{{.ComponentName}}()))
	{{end}}
	log.Println("UI routes registered.")
}
`
