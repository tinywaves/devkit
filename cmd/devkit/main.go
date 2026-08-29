package main

import (
	"fmt"
	"os"

	"github.com/tinywaves/devkit/internal/app"
)

var version = "dev"

func main() {
	rootCommand := app.NewRootCommand(os.Stdin, os.Stdout, version)
	rootCommand.SetErr(os.Stderr)
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "devkit: %v\n", err)
		os.Exit(1)
	}
}
