package app

import (
	"strings"
	"testing"
)

func TestNewRootCommandIncludesInit(t *testing.T) {
	var output strings.Builder
	command := NewRootCommand(strings.NewReader(""), &output, "test")
	command.SetArgs([]string{"init", "eslint"})

	if err := command.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}
	if got, want := output.String(), "Initializing eslint (placeholder)\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
