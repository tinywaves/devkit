package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCommand := &cobra.Command{
		Use:           "devkit",
		Short:         "Automate project setup and local environment preparation",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	rootCommand.SetVersionTemplate("devkit version {{.Version}}\n")
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "devkit: %v\n", err)
		os.Exit(1)
	}
}
