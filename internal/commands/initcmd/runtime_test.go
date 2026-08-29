package initcmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitializeJSTSRuntimeUsesConfiguredVersions(t *testing.T) {
	workdir := t.TempDir()
	packagePath := filepath.Join(workdir, packageJSONName)
	if err := os.WriteFile(packagePath, []byte("{\"name\":\"example\",\"scripts\":{\"test\":\"go test\"}}\n"), 0o640); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	var output strings.Builder
	input := strings.NewReader("\n\n\n")
	if err := initializeInteractiveInit(context.Background(), input, &output, workdir); err != nil {
		t.Fatalf("initialize runtime: %v", err)
	}
	if strings.Contains(output.String(), "Fetching") {
		t.Fatalf("output = %q, should not fetch runtime versions", output.String())
	}

	data, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var packageJSON map[string]any
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		t.Fatalf("parse package.json: %v", err)
	}

	if got, want := packageJSON["packageManager"], "pnpm@12.1.0"; got != want {
		t.Fatalf("packageManager = %v, want %v", got, want)
	}
	devEngines, ok := packageJSON["devEngines"].(map[string]any)
	if !ok {
		t.Fatalf("devEngines = %T, want object", packageJSON["devEngines"])
	}
	packageManager, ok := devEngines["packageManager"].(map[string]any)
	if !ok {
		t.Fatalf("devEngines.packageManager = %T, want object", devEngines["packageManager"])
	}
	if got, want := packageManager["version"], "^12.1.0"; got != want {
		t.Fatalf("package manager version = %v, want %v", got, want)
	}
	runtime, ok := devEngines["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("devEngines.runtime = %T, want object", devEngines["runtime"])
	}
	if got, want := runtime["version"], "^26.8.1"; got != want {
		t.Fatalf("runtime version = %v, want %v", got, want)
	}

	fileInfo, err := os.Stat(packagePath)
	if err != nil {
		t.Fatalf("stat package.json: %v", err)
	}
	if got, want := fileInfo.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("package.json mode = %o, want %o", got, want)
	}
}

func TestConfiguredRuntimeVersions(t *testing.T) {
	versions := make([]string, 0, len(configuredRuntimeVersions.NodeVersions))
	for _, version := range configuredRuntimeVersions.NodeVersions {
		versions = append(versions, version.Version)
	}
	if got, want := strings.Join(versions, ","), "26.8.1,25.9.0,24.20.0,23.11.1,22.23.2,21.7.3,20.20.2,19.9.0,18.20.8"; got != want {
		t.Fatalf("Node.js versions = %q, want %q", got, want)
	}
	for _, index := range []int{2, 4, 6, 8} {
		if !configuredRuntimeVersions.NodeVersions[index].LTS {
			t.Fatalf("Node.js %s should be marked as LTS", configuredRuntimeVersions.NodeVersions[index].Version)
		}
	}
	if got, want := strings.Join(configuredRuntimeVersions.PnpmVersions, ","), "12.1.0,11.25.0,10.34.5"; got != want {
		t.Fatalf("pnpm versions = %q, want %q", got, want)
	}
}

func TestSelectRuntimeVersionsSupportsArrowKeysAndCustomNodeVersion(t *testing.T) {
	input := strings.NewReader("\x1b[B\x1b[B\n24.18.0\n\x1b[B\n")
	var output strings.Builder
	prompts := newPromptSession(input, &output)

	nodeVersion, pnpmVersion, err := selectRuntimeVersions(context.Background(), prompts, []nodeVersion{{Version: "24.18.0", LTS: true}, {Version: "22.14.0"}}, []string{"11.22.0", "10.2.0"})
	if err != nil {
		t.Fatalf("select runtime versions: %v", err)
	}

	if got, want := nodeVersion, "24.18.0"; got != want {
		t.Fatalf("Node.js version = %q, want %q", got, want)
	}
	if got, want := pnpmVersion, "10.2.0"; got != want {
		t.Fatalf("pnpm version = %q, want %q", got, want)
	}
	if !strings.Contains(output.String(), "24.18.0 (LTS)") {
		t.Fatalf("output = %q, want LTS label", output.String())
	}
}

func TestNormalizeStableVersion(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "24.18.0", want: "24.18.0"},
		{input: "v24.18.0", want: "24.18.0"},
	} {
		got, err := normalizeStableVersion(test.input)
		if err != nil {
			t.Fatalf("normalizeStableVersion(%q): %v", test.input, err)
		}
		if got != test.want {
			t.Errorf("normalizeStableVersion(%q) = %q, want %q", test.input, got, test.want)
		}
	}

	for _, input := range []string{"24.18", "24.18.0-beta.1", "latest"} {
		if err := validateStableVersion(input); err == nil {
			t.Errorf("validateStableVersion(%q) error = nil, want error", input)
		}
	}
}
