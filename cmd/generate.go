package cmd

import (
	ar_templ "github.com/bata94/apiright/pkg/templ"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	inputDir    string
	outputFile  string
	packageName string
)

const (
	defaultInputDir       = "./ui/pages"
	defaultOutputFileName = "uirouter/routes_gen.go"
	defaultPackageName    = "uirouter"
)

var genCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate go code files",
	Long:  "Generate go code files",
	Run: func(cmd *cobra.Command, args []string) {
		ar_templ.GeneratorRun(inputDir, outputFile, packageName)
	},
}

func init() {
	var err error

	genCmd.PersistentFlags().StringVarP(&inputDir, "input", "i", defaultInputDir, "Input directory containing .templ files (required)")
	err = viper.BindPFlag("input", genCmd.PersistentFlags().Lookup("input"))
	if err != nil {
		log.Fatal("Error binding input flag: ", err)
		return
	}

	genCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", defaultOutputFileName, "Output file name for generated routes.go")
	err = viper.BindPFlag("output", genCmd.PersistentFlags().Lookup("output"))
	if err != nil {
		log.Fatal("Error binding output flag: ", err)
		return
	}

	genCmd.PersistentFlags().StringVarP(&packageName, "package", "p", defaultPackageName, "Package name for the generated routes.go file")
	err = viper.BindPFlag("package", genCmd.PersistentFlags().Lookup("package"))
	if err != nil {
		log.Fatal("Error binding package flag: ", err)
		return
	}

	rootCmd.AddCommand(genCmd)
}
