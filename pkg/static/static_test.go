package static

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bata94/apiright/pkg/core"
)

func TestNewStaticServeFileConfig(t *testing.T) {
	tests := []struct {
		name     string
		opts     []StaticServFileOption
		expected *StaticSevFileConfig
	}{
		{
			name: "default config",
			opts: []StaticServFileOption{},
			expected: &StaticSevFileConfig{
				preLoad:         true,
				preCache:        true,
				maxPreCacheSize: 1024 * 1024 * 10, // 10 MB
				forcePreCache:   false,
				contentType:     "",
				indexFile:       "index.html",
			},
		},
		{
			name: "without preload",
			opts: []StaticServFileOption{WithoutPreLoad()},
			expected: &StaticSevFileConfig{
				preLoad:         false,
				preCache:        false,
				maxPreCacheSize: 1024 * 1024 * 10,
				forcePreCache:   false,
				contentType:     "",
				indexFile:       "index.html",
			},
		},
		{
			name: "with custom content type",
			opts: []StaticServFileOption{WithContentType("application/json")},
			expected: &StaticSevFileConfig{
				preLoad:         true,
				preCache:        true,
				maxPreCacheSize: 1024 * 1024 * 10,
				forcePreCache:   false,
				contentType:     "application/json",
				indexFile:       "index.html",
			},
		},
		{
			name: "with custom index file",
			opts: []StaticServFileOption{WithIndexFile("default.htm")},
			expected: &StaticSevFileConfig{
				preLoad:         true,
				preCache:        true,
				maxPreCacheSize: 1024 * 1024 * 10,
				forcePreCache:   false,
				contentType:     "",
				indexFile:       "default.htm",
			},
		},
		{
			name: "with precache size",
			opts: []StaticServFileOption{WithPreCacheSize(1024)},
			expected: &StaticSevFileConfig{
				preLoad:         true,
				preCache:        true,
				maxPreCacheSize: 1024,
				forcePreCache:   false,
				contentType:     "",
				indexFile:       "index.html",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStaticServeFileConfig(tt.opts...)
			if config.preLoad != tt.expected.preLoad {
				t.Errorf("preLoad = %v, want %v", config.preLoad, tt.expected.preLoad)
			}
			if config.preCache != tt.expected.preCache {
				t.Errorf("preCache = %v, want %v", config.preCache, tt.expected.preCache)
			}
			if config.maxPreCacheSize != tt.expected.maxPreCacheSize {
				t.Errorf("maxPreCacheSize = %v, want %v", config.maxPreCacheSize, tt.expected.maxPreCacheSize)
			}
			if config.forcePreCache != tt.expected.forcePreCache {
				t.Errorf("forcePreCache = %v, want %v", config.forcePreCache, tt.expected.forcePreCache)
			}
			if config.contentType != tt.expected.contentType {
				t.Errorf("contentType = %v, want %v", config.contentType, tt.expected.contentType)
			}
			if config.indexFile != tt.expected.indexFile {
				t.Errorf("indexFile = %v, want %v", config.indexFile, tt.expected.indexFile)
			}
		})
	}
}

func TestServeStaticFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name           string
		urlPath        string
		filePath       string
		opts           []StaticServFileOption
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "serve existing file with preload",
			urlPath:        "/test",
			filePath:       testFile,
			opts:           []StaticServFileOption{},
			expectedStatus: 200,
			expectedBody:   testContent,
		},
		{
			name:           "serve existing file without preload",
			urlPath:        "/test",
			filePath:       testFile,
			opts:           []StaticServFileOption{WithoutPreLoad()},
			expectedStatus: 200,
			expectedBody:   testContent,
		},
		{
			name:           "serve non-existing file",
			urlPath:        "/nonexistent",
			filePath:       filepath.Join(tempDir, "nonexistent.txt"),
			opts:           []StaticServFileOption{WithoutPreLoad()},
			expectedStatus: 404,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ServeStaticFile(tt.urlPath, tt.filePath, tt.opts...)

			req := httptest.NewRequest("GET", tt.urlPath, nil)
			w := httptest.NewRecorder()

			ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

			err := handler(ctx)
			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}

			// Write the response to the recorder
			ctx.SendingReturn(w, nil)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}
		})
	}
}

func TestServeStaticDir(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create test files
	indexFile := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(indexFile, []byte("<html><body>Index</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	err = os.WriteFile(subFile, []byte("Sub content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	tests := []struct {
		name           string
		urlPath        string
		dirPath        string
		requestPath    string
		opts           []StaticServFileOption
		expectedStatus int
		checkBody      func(t *testing.T, body string)
	}{
		{
			name:           "serve directory with index file",
			urlPath:        "/static",
			dirPath:        tempDir,
			requestPath:    "/static/",
			opts:           []StaticServFileOption{},
			expectedStatus: 200,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "<html><body>Index</body></html>") {
					t.Errorf("Expected index file content, got: %s", body)
				}
			},
		},
		{
			name:           "serve directory without index file",
			urlPath:        "/static",
			dirPath:        subDir,
			requestPath:    "/static/",
			opts:           []StaticServFileOption{},
			expectedStatus: 200,
			checkBody: func(t *testing.T, body string) {
				if !strings.Contains(body, "File Explorer") {
					t.Errorf("Expected directory listing, got: %s", body)
				}
			},
		},
		{
			name:           "serve specific file",
			urlPath:        "/static",
			dirPath:        tempDir,
			requestPath:    "/static/test.txt",
			opts:           []StaticServFileOption{WithoutPreLoad()},
			expectedStatus: 200,
			checkBody: func(t *testing.T, body string) {
				if body != "Test content" {
					t.Errorf("Expected 'Test content', got: %s", body)
				}
			},
		},
		{
			name:           "serve non-existing file",
			urlPath:        "/static",
			dirPath:        tempDir,
			requestPath:    "/static/nonexistent.txt",
			opts:           []StaticServFileOption{WithoutPreLoad()},
			expectedStatus: 404,
			checkBody:      func(t *testing.T, body string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := ServeStaticDir(tt.urlPath, tt.dirPath, tt.opts...)

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

			err := handler(ctx)
			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}

			// Write the response to the recorder
			ctx.SendingReturn(w, nil)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.checkBody != nil {
				body := w.Body.String()
				tt.checkBody(t, body)
			}
		})
	}
}

func TestGetStaticDirData(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create test files
	indexFile := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(indexFile, []byte("index content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := NewStaticServeFileConfig()

	dirData, err := getStaticDirData(tempDir, tempDir, "/static/", config)
	if err != nil {
		t.Fatalf("getStaticDirData returned error: %v", err)
	}

	if !dirData.IndexFileExists {
		t.Error("Expected IndexFileExists to be true")
	}

	if len(dirData.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(dirData.Files))
	}

	if len(dirData.Dirs) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(dirData.Dirs))
	}

	// Check that both files are present
	fileNames := make(map[string]bool)
	for _, file := range dirData.Files {
		fileNames[file.Name] = true
	}

	if !fileNames["index.html"] {
		t.Error("Expected to find 'index.html' in files")
	}

	if !fileNames["test.txt"] {
		t.Error("Expected to find 'test.txt' in files")
	}

	if dirData.Dirs[0].Name != "subdir" {
		t.Errorf("Expected dir name 'subdir', got '%s'", dirData.Dirs[0].Name)
	}
}

func TestSetupTemplateFuncs(t *testing.T) {
	funcs := setupTemplateFuncs()

	if funcs["hasSuffix"] == nil {
		t.Error("hasSuffix function not found in template funcs")
	}

	if funcs["byteFormat"] == nil {
		t.Error("byteFormat function not found in template funcs")
	}

	// Test hasSuffix function
	hasSuffix := funcs["hasSuffix"].(func(string, string) bool)
	if !hasSuffix("test.txt", ".txt") {
		t.Error("hasSuffix should return true for 'test.txt' and '.txt'")
	}
	if hasSuffix("test.txt", ".md") {
		t.Error("hasSuffix should return false for 'test.txt' and '.md'")
	}

	// Test byteFormat function
	byteFormat := funcs["byteFormat"].(func(int64) string)
	if byteFormat(512) != "512 B" {
		t.Errorf("byteFormat(512) = %s, want '512 B'", byteFormat(512))
	}
	if byteFormat(1024) != "1.0 KB" {
		t.Errorf("byteFormat(1024) = %s, want '1.0 KB'", byteFormat(1024))
	}
	if byteFormat(1024*1024) != "1.0 MB" {
		t.Errorf("byteFormat(1024*1024) = %s, want '1.0 MB'", byteFormat(1024*1024))
	}
}

func TestDecideStaticIndexFile(t *testing.T) {
	tempDir := t.TempDir()
	config := NewStaticServeFileConfig()

	// Test with index file present
	indexFile := filepath.Join(tempDir, "index.html")
	indexContent := "<html>Index</html>"
	err := os.WriteFile(indexFile, []byte(indexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	dirData := DirTemplateData{
		IndexFileExists: true,
	}

	templ, err := template.New("test").Funcs(setupTemplateFuncs()).Parse(dirExplorerTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result := decideStaticIndexFile(dirData, tempDir, templ, config)
	if string(result) != indexContent {
		t.Errorf("Expected index content, got: %s", string(result))
	}

	// Test without index file
	dirData.IndexFileExists = false
	dirData.Title = "Test"
	dirData.BaseUrl = "/test/"
	dirData.Files = []FileData{{Name: "test.txt", Size: 100}}
	dirData.Dirs = []DirData{{Name: "subdir"}}

	result = decideStaticIndexFile(dirData, tempDir, templ, config)
	if !bytes.Contains(result, []byte("File Explorer")) {
		t.Error("Expected directory listing template to be used")
	}
}

func TestLoadFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading existing file
	f, fInfo, err := loadFile(testFile)
	if err != nil {
		t.Errorf("loadFile returned error: %v", err)
	}
	if f == nil {
		t.Error("Expected file to be returned")
	}
	if fInfo == nil {
		t.Error("Expected file info to be returned")
	}
	closeFile(f)

	// Test loading non-existing file
	_, _, err = loadFile(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error for non-existing file")
	}
}
