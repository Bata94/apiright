package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bata94/apiright/pkg/core"
)

// Cache manages file-based caching for generated code
type Cache struct {
	dir    string
	logger core.Logger
}

// CacheMetadata stores metadata about cached generation
type CacheMetadata struct {
	SchemaHash  string            `json:"schema_hash"`
	ConfigHash  string            `json:"config_hash"`
	SQLCVersion string            `json:"sqlc_version"`
	GeneratedAt time.Time         `json:"generated_at"`
	Files       map[string]string `json:"files"` // file path -> hash
	Plugins     []string          `json:"plugins"`
}

// NewCache creates a new cache instance
func NewCache(projectDir string, logger core.Logger) (*Cache, error) {
	cacheDir := filepath.Join(projectDir, ".apiright_cache")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"files/sql", "files/go", "files/proto", "temp"}
	for _, subdir := range subdirs {
		fullPath := filepath.Join(cacheDir, subdir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache subdirectory %s: %w", subdir, err)
		}
	}

	return &Cache{
		dir:    cacheDir,
		logger: logger,
	}, nil
}

// ShouldRegenerate determines if regeneration is needed based on cache invalidation
func (c *Cache) ShouldRegenerate(migrationDir, configDir string) (bool, error) {
	// Load current cache metadata
	metadata, err := c.loadMetadata()
	if err != nil {
		c.logger.Debug("No cache metadata found, regeneration needed")
		return true, nil
	}

	// Calculate current hashes
	currentSchemaHash, err := c.calculateSchemaHash(migrationDir)
	if err != nil {
		c.logger.Warn("Failed to calculate schema hash", "error", err)
		return true, nil
	}

	currentConfigHash, err := c.calculateConfigHash(configDir)
	if err != nil {
		c.logger.Warn("Failed to calculate config hash", "error", err)
		return true, nil
	}

	currentSQLCVersion, err := c.getSQLCVersion()
	if err != nil {
		c.logger.Warn("Failed to get sqlc version", "error", err)
		return true, nil
	}

	// Check if anything changed
	if metadata.SchemaHash != currentSchemaHash {
		c.logger.Debug("Schema hash changed, regeneration needed")
		return true, nil
	}

	if metadata.ConfigHash != currentConfigHash {
		c.logger.Debug("Config hash changed, regeneration needed")
		return true, nil
	}

	if metadata.SQLCVersion != currentSQLCVersion {
		c.logger.Debug("sqlc version changed, regeneration needed")
		return true, nil
	}

	c.logger.Debug("Cache is valid, no regeneration needed")
	return false, nil
}

// RestoreFromCache restores files from cache
func (c *Cache) RestoreFromCache(ctx *core.GenerationContext) error {
	metadata, err := c.loadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load cache metadata: %w", err)
	}

	// Restore cached files to their destinations
	sourceBase := filepath.Join(c.dir, "files")

	for _, subDir := range []string{"sql", "go", "proto"} {
		sourceDir := filepath.Join(sourceBase, subDir)
		destDir := ctx.Join(ctx.ProjectDir, "gen", subDir)

		if err := c.restoreDirectory(sourceDir, destDir); err != nil {
			return fmt.Errorf("failed to restore %s files from cache: %w", subDir, err)
		}
	}

	c.logger.Info("Restored files from cache", "files", len(metadata.Files), "generated_at", metadata.GeneratedAt)
	return nil
}

// SaveToCache saves generated files to cache
func (c *Cache) SaveToCache(ctx *core.GenerationContext, migrationDir, configDir string) error {
	// Calculate current metadata
	metadata := &CacheMetadata{
		GeneratedAt: time.Now(),
		Files:       make(map[string]string),
		Plugins:     []string{}, // TODO: Get from context
	}

	var err error
	metadata.SchemaHash, err = c.calculateSchemaHash(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to calculate schema hash: %w", err)
	}

	metadata.ConfigHash, err = c.calculateConfigHash(configDir)
	if err != nil {
		return fmt.Errorf("failed to calculate config hash: %w", err)
	}

	metadata.SQLCVersion, err = c.getSQLCVersion()
	if err != nil {
		return fmt.Errorf("failed to get sqlc version: %w", err)
	}

	// Save generated files to cache
	genDir := ctx.Join(ctx.ProjectDir, "gen")

	for _, subDir := range []string{"sql", "go", "proto"} {
		sourceDir := filepath.Join(genDir, subDir)
		destDir := filepath.Join(c.dir, "files", subDir)

		if err := c.saveDirectory(sourceDir, destDir, metadata); err != nil {
			return fmt.Errorf("failed to cache %s files: %w", subDir, err)
		}
	}

	// Save metadata
	if err := c.saveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to save cache metadata: %w", err)
	}

	c.logger.Info("Saved files to cache", "files", len(metadata.Files))
	return nil
}

// Invalidate clears the cache
func (c *Cache) Invalidate() error {
	if err := os.RemoveAll(c.dir); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	// Recreate cache structure
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("failed to recreate cache directory: %w", err)
	}

	c.logger.Info("Cache invalidated")
	return nil
}

// loadMetadata loads cache metadata from file
func (c *Cache) loadMetadata() (*CacheMetadata, error) {
	metadataPath := filepath.Join(c.dir, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err // File not found is expected for first run
	}

	var metadata CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse cache metadata: %w", err)
	}

	return &metadata, nil
}

// saveMetadata saves cache metadata to file
func (c *Cache) saveMetadata(metadata *CacheMetadata) error {
	metadataPath := filepath.Join(c.dir, "metadata.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

// calculateSchemaHash calculates hash of all migration files
func (c *Cache) calculateSchemaHash(migrationDir string) (string, error) {
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return "", err
	}

	hash := sha256.New()

	// Sort files for consistent hashing
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			fileNames = append(fileNames, file.Name())
		}
	}

	for _, fileName := range fileNames {
		filePath := filepath.Join(migrationDir, fileName)
		fileHash, err := c.calculateFileHash(filePath)
		if err != nil {
			return "", err
		}

		hash.Write([]byte(fileName))
		hash.Write([]byte(fileHash))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateConfigHash calculates hash of configuration files
func (c *Cache) calculateConfigHash(configDir string) (string, error) {
	configFiles := []string{"sqlc.yaml", "apiright.yaml"}

	hash := sha256.New()

	for _, configFile := range configFiles {
		filePath := filepath.Join(configDir, configFile)
		fileHash, err := c.calculateFileHash(filePath)
		if err != nil {
			// Config files might not exist, that's ok
			continue
		}

		hash.Write([]byte(configFile))
		hash.Write([]byte(fileHash))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateFileHash calculates SHA-256 hash of a file
func (c *Cache) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer core.Close("file", file, c.logger)

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getSQLCVersion gets the current sqlc version
func (c *Cache) getSQLCVersion() (string, error) {
	// For now, return a simple version identifier
	// In the future, we could execute "sqlc version" command
	return "1.0.0", nil
}

// restoreDirectory restores files from cache to destination
func (c *Cache) restoreDirectory(sourceDir, destDir string) error {
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Source directory might not exist
		}
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		sourcePath := filepath.Join(sourceDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		if err := c.copyFile(sourcePath, destPath); err != nil {
			return err
		}
	}

	return nil
}

// saveDirectory saves files from source to cache
func (c *Cache) saveDirectory(sourceDir, destDir string, metadata *CacheMetadata) error {
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Source directory might not exist
		}
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		sourcePath := filepath.Join(sourceDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		// Copy file to cache
		if err := c.copyFile(sourcePath, destPath); err != nil {
			return err
		}

		// Add to metadata
		fileHash, err := c.calculateFileHash(sourcePath)
		if err != nil {
			return err
		}
		metadata.Files[filepath.Join(filepath.Base(destDir), file.Name())] = fileHash
	}

	return nil
}

// copyFile copies a file from source to destination
func (c *Cache) copyFile(sourcePath, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer core.Close("source", source, c.logger)

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer core.Close("dest", dest, c.logger)

	_, err = io.Copy(dest, source)
	return err
}
