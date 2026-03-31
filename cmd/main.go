package main

import (
	"fmt"
	"os"

	"github.com/bata94/apiright/cmd/apiright"
)

func main() {
	if err := apiright.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
