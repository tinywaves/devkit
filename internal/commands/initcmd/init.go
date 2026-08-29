package initcmd

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type initTarget struct {
	name    string
	label   string
	aliases []string
}

var initTargets = []initTarget{
	{
		name:    "js/ts-runtime",
		label:   "js/ts runtime",
		aliases: []string{"js/ts", "js-ts", "js-ts-runtime", "runtime"},
	},
	{
		name:  "eslint",
		label: "eslint",
	},
}

func NewCommand(input io.Reader) *cobra.Command {
	command := &cobra.Command{
		Use:   "init [js/ts-runtime|eslint]",
		Short: "Initialize project tooling",
		Args:  validateInitArgs,
	}
	command.RunE = func(command *cobra.Command, args []string) error {
		var target initTarget
		var err error
		if len(args) == 0 {
			workdir, workdirErr := os.Getwd()
			if workdirErr != nil {
				return fmt.Errorf("get current directory: %w", workdirErr)
			}
			return initializeInteractiveInit(command.Context(), input, command.OutOrStdout(), workdir)
		} else {
			target, err = findInitTarget(strings.Join(args, " "))
		}
		if err != nil {
			return err
		}

		if target.name == "js/ts-runtime" {
			workdir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get current directory: %w", err)
			}
			return initializeJSTSRuntime(command.Context(), input, command.OutOrStdout(), workdir)
		}

		fmt.Fprintln(command.OutOrStdout(), "Initializing", target.label, "(placeholder)")
		return nil
	}

	return command
}

func validateInitArgs(_ *cobra.Command, args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("accepts at most one initializer name")
	}
	if len(args) == 2 && !strings.EqualFold(strings.Join(args, " "), "js/ts runtime") {
		return fmt.Errorf("accepts at most one initializer name")
	}
	return nil
}

func findInitTarget(value string) (initTarget, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	for _, target := range initTargets {
		if normalized == target.name || normalized == target.label {
			return target, nil
		}
		if slices.Contains(target.aliases, normalized) {
			return target, nil
		}
	}

	return initTarget{}, fmt.Errorf("unknown initializer %q; choose js/ts-runtime or eslint", value)
}
