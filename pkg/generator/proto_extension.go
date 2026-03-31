package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bata94/apiright/pkg/core"
)

type ProtoExtensionProcessor struct {
	logger     core.Logger
	genSuffix  string
	extensions []ProtoExtensionFile
}

type ProtoExtensionFile struct {
	Name     string
	Path     string
	Content  string
	Imports  []string
	Services []ProtoExtensionService
}

type ProtoExtensionService struct {
	Name    string
	Methods []ProtoExtensionMethod
}

type ProtoExtensionMethod struct {
	Name         string
	RequestType  string
	ResponseType string
	HTTPMethod   string
	HTTPPath     string
}

func NewProtoExtensionProcessor(genSuffix string, logger core.Logger) *ProtoExtensionProcessor {
	return &ProtoExtensionProcessor{
		genSuffix:  genSuffix,
		logger:     logger,
		extensions: make([]ProtoExtensionFile, 0),
	}
}

func (pep *ProtoExtensionProcessor) ProcessExtensions(ctx *core.GenerationContext, extensions []core.ProtoExtension) error {
	pep.extensions = make([]ProtoExtensionFile, 0)

	for _, ext := range extensions {
		for _, protoFile := range ext.ProtoFiles() {
			if err := pep.processProtoFile(protoFile, ext.Imports()); err != nil {
				pep.logger.Warn("Failed to process proto extension", "file", protoFile, "error", err)
				continue
			}
		}
	}

	if err := pep.processUserProtoDirectory(ctx); err != nil {
		pep.logger.Warn("Failed to process user proto directory", "error", err)
	}

	pep.logger.Info("Processed proto extensions", "count", len(pep.extensions))
	return nil
}

func (pep *ProtoExtensionProcessor) processProtoFile(path string, additionalImports []string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read proto file %s: %w", path, err)
	}

	ext := ProtoExtensionFile{
		Name:    filepath.Base(path),
		Path:    path,
		Content: string(content),
		Imports: append(pep.extractImports(string(content)), additionalImports...),
	}

	ext.Services = pep.extractServices(string(content))
	pep.extensions = append(pep.extensions, ext)

	pep.logger.Debug("Processed proto file", "path", path, "services", len(ext.Services))
	return nil
}

func (pep *ProtoExtensionProcessor) processUserProtoDirectory(ctx *core.GenerationContext) error {
	protoDir := ctx.Join(ctx.ProjectDir, "proto")
	if _, err := os.Stat(protoDir); os.IsNotExist(err) {
		return nil
	}

	files, err := os.ReadDir(protoDir)
	if err != nil {
		return fmt.Errorf("failed to read proto directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".proto") {
			continue
		}

		protoPath := filepath.Join(protoDir, file.Name())
		if err := pep.processProtoFile(protoPath, nil); err != nil {
			pep.logger.Warn("Failed to process user proto", "file", protoPath, "error", err)
		}
	}

	return nil
}

func (pep *ProtoExtensionProcessor) extractImports(content string) []string {
	importRegex := regexp.MustCompile(`import\s+"([^"]+)";`)
	matches := importRegex.FindAllStringSubmatch(content, -1)

	imports := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	return imports
}

func (pep *ProtoExtensionProcessor) extractServices(content string) []ProtoExtensionService {
	serviceRegex := regexp.MustCompile(`service\s+(\w+)\s*\{([^}]+)\}`)
	matches := serviceRegex.FindAllStringSubmatch(content, -1)

	services := make([]ProtoExtensionService, 0)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		svc := ProtoExtensionService{
			Name:    match[1],
			Methods: pep.extractMethods(match[2]),
		}
		services = append(services, svc)
	}

	return services
}

func (pep *ProtoExtensionProcessor) extractMethods(serviceBody string) []ProtoExtensionMethod {
	rpcRegex := regexp.MustCompile(`rpc\s+(\w+)\s*\(([^)]+)\)\s*returns\s*\(([^)]+)\)`)
	matches := rpcRegex.FindAllStringSubmatch(serviceBody, -1)

	methods := make([]ProtoExtensionMethod, 0, len(matches))
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		method := ProtoExtensionMethod{
			Name:         match[1],
			RequestType:  strings.TrimSpace(match[2]),
			ResponseType: strings.TrimSpace(match[3]),
		}
		methods = append(methods, method)
	}

	return methods
}

func (pep *ProtoExtensionProcessor) ValidateImports(ctx *core.GenerationContext) error {
	generatedProtoName := "db" + pep.genSuffix + ".proto"

	for _, ext := range pep.extensions {
		for _, imp := range ext.Imports {
			if strings.HasSuffix(imp, generatedProtoName) {
				continue
			}
			if strings.HasPrefix(imp, "google/protobuf/") {
				continue
			}
			if !pep.importExists(imp) {
				pep.logger.Warn("Proto import not found",
					"file", ext.Name,
					"import", imp)
			}
		}
	}

	return nil
}

func (pep *ProtoExtensionProcessor) importExists(importPath string) bool {
	for _, ext := range pep.extensions {
		if strings.HasSuffix(ext.Path, importPath) {
			return true
		}
	}

	genProtoDir := filepath.Join("gen", "proto")
	if _, err := os.Stat(filepath.Join(genProtoDir, importPath)); err == nil {
		return true
	}

	return false
}

func (pep *ProtoExtensionProcessor) GetExtensionServices() []ProtoExtensionService {
	var allServices []ProtoExtensionService
	for _, ext := range pep.extensions {
		allServices = append(allServices, ext.Services...)
	}
	return allServices
}

func (pep *ProtoExtensionProcessor) CopyExtensionsToGen(ctx *core.GenerationContext) error {
	genProtoDir := ctx.Join(ctx.ProjectDir, "gen", "proto")
	if err := os.MkdirAll(genProtoDir, 0755); err != nil {
		return fmt.Errorf("failed to create gen/proto directory: %w", err)
	}

	for _, ext := range pep.extensions {
		extName := filepath.Base(ext.Path)
		destPath := filepath.Join(genProtoDir, "ext_"+extName)

		if err := os.WriteFile(destPath, []byte(ext.Content), 0644); err != nil {
			pep.logger.Warn("Failed to copy proto extension", "source", ext.Path, "dest", destPath)
			continue
		}

		pep.logger.Debug("Copied proto extension", "source", ext.Path, "dest", destPath)
	}

	return nil
}

func (pep *ProtoExtensionProcessor) GetGeneratedProtoPath(ctx *core.GenerationContext) string {
	return ctx.Join(ctx.ProjectDir, "gen", "proto", "db"+pep.genSuffix+".proto")
}

func (pep *ProtoExtensionProcessor) GetAPIGenProtoPath(ctx *core.GenerationContext) string {
	return ctx.Join(ctx.ProjectDir, "gen", "proto", "api"+pep.genSuffix+".proto")
}

func (pep *ProtoExtensionProcessor) HasExtensions() bool {
	return len(pep.extensions) > 0
}

func (pep *ProtoExtensionProcessor) GetExtensions() []ProtoExtensionFile {
	return pep.extensions
}
