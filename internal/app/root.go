package app

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/tinywaves/devkit/internal/commands/initcmd"
)

func NewRootCommand(input io.Reader, output io.Writer, version string) *cobra.Command {
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
	rootCommand.SetOut(output)
	rootCommand.SetVersionTemplate("devkit version {{.Version}}\n")
	rootCommand.AddCommand(initcmd.NewCommand(input))
	return rootCommand
}
