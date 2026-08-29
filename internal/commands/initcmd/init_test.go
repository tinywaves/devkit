package initcmd

import (
	"strings"
	"testing"
)

func TestInitCommandWithArgument(t *testing.T) {
	var output strings.Builder
	command := NewCommand(strings.NewReader(""))
	command.SetOut(&output)
	command.SetArgs([]string{"eslint"})

	if err := command.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}

	if got, want := output.String(), "Initializing eslint (placeholder)\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestInitCommandWithPrompt(t *testing.T) {
	var output strings.Builder
	command := NewCommand(strings.NewReader("\n"))
	command.SetOut(&output)

	if err := command.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}

	if !strings.Contains(output.String(), "package.json not found; skipping JavaScript / TypeScript initialization\n") {
		t.Fatalf("output = %q, want package.json skip message", output.String())
	}
}

func TestFindInitTargetRejectsUnknownValue(t *testing.T) {
	if _, err := findInitTarget("python"); err == nil {
		t.Fatal("findInitTarget() error = nil, want error")
	}
}

func TestInitCommandAcceptsJSTarget(t *testing.T) {
	var output strings.Builder
	command := NewCommand(strings.NewReader(""))
	command.SetOut(&output)
	command.SetArgs([]string{"js"})

	if err := command.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}

	if !strings.Contains(output.String(), "package.json not found; skipping JavaScript / TypeScript initialization\n") {
		t.Fatalf("output = %q, want package.json skip message", output.String())
	}
}

func TestInitCommandRejectsAliases(t *testing.T) {
	for _, value := range []string{"javascript", "typescript", "node", "js/ts-runtime"} {
		if _, err := findInitTarget(value); err == nil {
			t.Fatalf("findInitTarget(%q) error = nil, want error", value)
		}
	}
}
