package apiright

import (
	"fmt"
	"os"

	"github.com/bata94/apiright/pkg/core"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apiright",
	Short: "APIRight - Auto-generate APIs from SQL schemas",
	Long: `APIRight is a modular Go framework that auto-generates v0 APIs
from sqlc schemas, providing rapid development capabilities for startup teams.`,
	Example: `  apiright init myproject
  apiright gen
  apiright serve
  apiright migrate up
  apiright --help`,
}

func Execute() error {
	rootCmd.Version = core.Version
	fmt.Fprintf(os.Stderr, "APIRight version %s\n", core.Version)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("project", "p", "", "project directory")
	rootCmd.PersistentFlags().BoolVar(&devMode, "dev", false, "enable development mode")

	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewGenCommand())
	rootCmd.AddCommand(NewServeCommand())
	rootCmd.AddCommand(NewMigrateCommand())
	rootCmd.AddCommand(NewDBCCommand())
	rootCmd.AddCommand(NewCacheCommand())
	rootCmd.AddCommand(NewDoctorCommand())
	rootCmd.AddCommand(NewVersionCommand())
}
