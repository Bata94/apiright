package apiright

import (
	"fmt"

	"github.com/bata94/apiright/pkg/core"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Show version information",
		Long:    `Display APIRight version and build information.`,
		Example: `  apiright version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("APIRight - Auto-generate APIs from SQL schemas")
			fmt.Println()
			fmt.Printf("Version: %s\n", core.Version)
			fmt.Println("Go: 1.25+")
			fmt.Println("License: MIT")
			fmt.Println()
			fmt.Println("Features:")
			fmt.Println("  - Auto-generate CRUD operations from SQL schemas")
			fmt.Println("  - Content negotiation: JSON, XML, YAML, Protobuf, Plain Text")
			fmt.Println("  - HTTP and gRPC endpoints")
			fmt.Println("  - Plugin system for extensibility")
			fmt.Println("  - Database migrations")
		},
	}

	return cmd
}
