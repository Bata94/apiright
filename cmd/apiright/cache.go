package apiright

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage generation cache",
		Long:  `Clear or manage the generation cache to force regeneration of code.`,
		Example: `  apiright cache clean
  apiright cache status`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "clean",
		Short:   "Clear the generation cache",
		Long:    `Removes the generation cache directory to force full regeneration.`,
		Example: `  apiright cache clean`,
		RunE:    runCacheClean,
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show cache status",
		Long:    `Displays the current status of the generation cache.`,
		Example: `  apiright cache status`,
		RunE:    runCacheStatus,
	})

	return cmd
}

func runCacheClean(cmd *cobra.Command, args []string) error {
	projectDir, err := GetProjectDir(cmd)
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	cacheDir := filepath.Join(projectDir, ".apiright_cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Println("Cache directory does not exist, nothing to clean")
		return nil
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Printf("Cache cleared successfully: %s\n", cacheDir)
	return nil
}

func runCacheStatus(cmd *cobra.Command, args []string) error {
	projectDir, err := GetProjectDir(cmd)
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	cacheDir := filepath.Join(projectDir, ".apiright_cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Println("Cache directory does not exist")
		fmt.Println("Status: Empty (no cache)")
		return nil
	}

	metadataPath := filepath.Join(cacheDir, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		fmt.Println("Cache directory exists but no metadata found")
		fmt.Println("Status: Empty (no cached data)")
		return nil
	}

	fmt.Println("Cache Status: Active")
	fmt.Printf("Cache Directory: %s\n", cacheDir)

	return nil
}
