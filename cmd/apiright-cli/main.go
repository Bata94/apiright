package main

import (
	"flag"

	ar_templ "github.com/bata94/apiright/pkg/templ"
)

var (
	inputDir    string
	outputFile  string
	packageName string
)

const (
	defaultOutputFileName = "routes_gen.go"
)

func main() {
	flag.StringVar(&inputDir, "input", "", "Input directory containing .templ files (required)")
	flag.StringVar(&outputFile, "output", defaultOutputFileName, "Output file name for generated routes.go")
	flag.StringVar(&packageName, "package", "uirouter", "Package name for the generated routes.go file")

	flag.Parse()

	ar_templ.GeneratorRun(inputDir, outputFile, packageName)
}
