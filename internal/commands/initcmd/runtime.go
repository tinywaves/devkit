package initcmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/yarlson/tap"

	runtimeconfig "github.com/tinywaves/devkit/internal/runtime"
)

const (
	packageJSONName = "package.json"
	customVersion   = "custom"
)

type nodeVersion = runtimeconfig.NodeVersion

var configuredRuntimeVersions = runtimeconfig.Versions()

func initializeJSToolchain(ctx context.Context, input io.Reader, output io.Writer, workdir string) error {
	packagePath := filepath.Join(workdir, packageJSONName)
	if _, err := os.Stat(packagePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(output, "package.json not found; skipping JavaScript / TypeScript initialization")
			return nil
		}
		return fmt.Errorf("check %s: %w", packageJSONName, err)
	}

	prompts := newPromptSession(input, output)
	nodeVersion, pnpmVersion, err := selectRuntimeVersions(ctx, prompts, configuredRuntimeVersions.NodeVersions, configuredRuntimeVersions.PnpmVersions)
	if err != nil {
		return err
	}
	if err := updatePackageJSON(packagePath, nodeVersion, pnpmVersion); err != nil {
		return fmt.Errorf("update %s: %w", packageJSONName, err)
	}

	fmt.Fprintf(output, "Configured Node.js %s and pnpm %s in %s\n", nodeVersion, pnpmVersion, packageJSONName)
	return nil
}

func initializeInteractiveInit(ctx context.Context, input io.Reader, output io.Writer, workdir string) error {
	prompts := newPromptSession(input, output)
	packagePath := filepath.Join(workdir, packageJSONName)
	_, packageErr := os.Stat(packagePath)
	hasPackageJSON := packageErr == nil
	if packageErr != nil && !os.IsNotExist(packageErr) {
		return fmt.Errorf("check %s: %w", packageJSONName, packageErr)
	}
	targetSelection, err := selectInitTarget(ctx, prompts)
	if err != nil {
		return err
	}
	target, err := findInitTarget(targetSelection)
	if err != nil {
		return err
	}
	if target.name != "js" {
		fmt.Fprintln(output, "Initializing", target.label, "(placeholder)")
		return nil
	}
	if !hasPackageJSON {
		fmt.Fprintln(output, "package.json not found; skipping JavaScript / TypeScript initialization")
		return nil
	}

	nodeVersion, pnpmVersion, err := selectRuntimeVersions(ctx, prompts, configuredRuntimeVersions.NodeVersions, configuredRuntimeVersions.PnpmVersions)
	if err != nil {
		return err
	}
	if err := updatePackageJSON(packagePath, nodeVersion, pnpmVersion); err != nil {
		return fmt.Errorf("update %s: %w", packageJSONName, err)
	}
	fmt.Fprintf(output, "Configured Node.js %s and pnpm %s in %s\n", nodeVersion, pnpmVersion, packageJSONName)
	return nil
}

func selectInitTarget(ctx context.Context, prompts *promptSession) (string, error) {
	options := make([]tap.SelectOption[string], 0, len(initTargets))
	for _, target := range initTargets {
		options = append(options, tap.SelectOption[string]{Label: target.label, Value: target.name})
	}
	selection, err := prompts.selectValue(ctx, "What do you want to initialize?", options)
	if err != nil {
		return "", fmt.Errorf("select initializer: %w", err)
	}
	return selection, nil
}

func selectRuntimeVersions(ctx context.Context, prompts *promptSession, nodeVersions []nodeVersion, pnpmVersions []string) (string, string, error) {
	nodeOptions := make([]tap.SelectOption[string], 0, len(nodeVersions)+1)
	for _, version := range nodeVersions {
		nodeOptions = append(nodeOptions, tap.SelectOption[string]{Label: nodeVersionLabel(version), Value: version.Version})
	}
	nodeOptions = append(nodeOptions, tap.SelectOption[string]{Label: "custom", Value: customVersion})

	nodeSelection, err := prompts.selectValue(ctx, "Which Node.js version do you want to use?", nodeOptions)
	if err != nil {
		return "", "", fmt.Errorf("select runtime versions: %w", err)
	}

	nodeVersion := nodeSelection
	if nodeSelection == customVersion {
		customNodeVersion, err := prompts.textValue(ctx, "Enter a Node.js version", nodeVersions[0].Version, validateStableVersion)
		if err != nil {
			return "", "", fmt.Errorf("select runtime versions: %w", err)
		}
		nodeVersion, err = normalizeStableVersion(customNodeVersion)
		if err != nil {
			return "", "", err
		}
	}

	pnpmOptions := make([]tap.SelectOption[string], 0, len(pnpmVersions))
	for _, version := range pnpmVersions {
		pnpmOptions = append(pnpmOptions, tap.SelectOption[string]{Label: version, Value: version})
	}
	pnpmVersion, err := prompts.selectValue(ctx, "Which pnpm version do you want to use?", pnpmOptions)
	if err != nil {
		return "", "", fmt.Errorf("select runtime versions: %w", err)
	}
	return nodeVersion, pnpmVersion, nil
}

func nodeVersionLabel(version nodeVersion) string {
	if version.LTS {
		return version.Version + " (LTS)"
	}
	return version.Version
}

type promptSession struct {
	native bool
	input  tap.Reader
	output tap.Writer
}

func newPromptSession(input io.Reader, output io.Writer) *promptSession {
	if isTerminalFile(input) && isTerminalFile(output) {
		return &promptSession{native: true}
	}

	return &promptSession{
		input:  &scriptedPromptReader{source: bufio.NewReader(input)},
		output: &promptWriter{target: output},
	}
}

func (session *promptSession) selectValue(ctx context.Context, message string, options []tap.SelectOption[string]) (string, error) {
	selectOptions := tap.SelectOptions[string]{Message: message, Options: options}
	if !session.native {
		selectOptions.Input = session.input
		selectOptions.Output = session.output
	}
	selection := tap.Select(ctx, selectOptions)
	if selection == "" {
		return "", errors.New("prompt canceled")
	}
	return selection, nil
}

func (session *promptSession) textValue(ctx context.Context, message, placeholder string, validate func(string) error) (string, error) {
	textOptions := tap.TextOptions{Message: message, Placeholder: placeholder, Validate: validate}
	if !session.native {
		textOptions.Input = session.input
		textOptions.Output = session.output
	}
	value := tap.Text(ctx, textOptions)
	if value == "" {
		return "", errors.New("prompt canceled")
	}
	return value, nil
}

func isTerminalFile(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

type scriptedPromptReader struct {
	source *bufio.Reader
}

func (reader *scriptedPromptReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (reader *scriptedPromptReader) On(event string, handler func(string, tap.Key)) {
	if event != "keypress" {
		return
	}
	line, err := reader.source.ReadString('\n')
	if len(line) == 0 && err != nil {
		return
	}
	for index := 0; index < len(line); index++ {
		if line[index] == '\x1b' && index+2 < len(line) && line[index+1] == '[' {
			keyName := map[byte]string{'A': "up", 'B': "down", 'C': "right", 'D': "left"}[line[index+2]]
			if keyName != "" {
				handler("", tap.Key{Name: keyName})
				index += 2
				continue
			}
		}
		switch line[index] {
		case '\r', '\n':
			handler("", tap.Key{Name: "return"})
		case '\x03':
			handler("\x03", tap.Key{Name: "c", Ctrl: true})
		case 127, '\b':
			handler("", tap.Key{Name: "backspace"})
		default:
			handler(string(line[index]), tap.Key{Rune: rune(line[index])})
		}
	}
}

type promptWriter struct {
	target io.Writer
}

func (writer *promptWriter) Write(data []byte) (int, error) {
	return writer.target.Write(data)
}

func (writer *promptWriter) On(string, func()) {}

func (writer *promptWriter) Emit(string) {}

func updatePackageJSON(path, nodeVersion, pnpmVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	packageJSON := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}
	if packageJSON == nil {
		return fmt.Errorf("parse JSON: expected an object")
	}

	devEngines := make(map[string]json.RawMessage)
	if raw, ok := packageJSON["devEngines"]; ok {
		if err := json.Unmarshal(raw, &devEngines); err != nil {
			return fmt.Errorf("parse devEngines: %w", err)
		}
		if devEngines == nil {
			return fmt.Errorf("parse devEngines: expected an object")
		}
	}

	packageManager, err := json.Marshal(map[string]string{
		"name":    "pnpm",
		"version": "^" + pnpmVersion,
		"onFail":  "download",
	})
	if err != nil {
		return fmt.Errorf("encode package manager: %w", err)
	}
	runtime, err := json.Marshal(map[string]string{
		"name":    "node",
		"version": "^" + nodeVersion,
		"onFail":  "download",
	})
	if err != nil {
		return fmt.Errorf("encode runtime: %w", err)
	}
	devEngines["packageManager"] = packageManager
	devEngines["runtime"] = runtime
	packageJSON["packageManager"], err = json.Marshal("pnpm@" + pnpmVersion)
	if err != nil {
		return fmt.Errorf("encode packageManager: %w", err)
	}
	packageJSON["devEngines"], err = json.Marshal(devEngines)
	if err != nil {
		return fmt.Errorf("encode devEngines: %w", err)
	}

	updated, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("encode package.json: %w", err)
	}
	updated = append(updated, '\n')
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, updated, fileInfo.Mode().Perm()); err != nil {
		return err
	}
	return verifyPackageJSON(path, nodeVersion, pnpmVersion)
}

func verifyPackageJSON(path, nodeVersion, pnpmVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	packageJSON := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	var packageManager string
	if err := json.Unmarshal(packageJSON["packageManager"], &packageManager); err != nil {
		return fmt.Errorf("read packageManager: %w", err)
	}
	if packageManager != "pnpm@"+pnpmVersion {
		return fmt.Errorf("packageManager is %q, want %q", packageManager, "pnpm@"+pnpmVersion)
	}

	var devEngines map[string]json.RawMessage
	if err := json.Unmarshal(packageJSON["devEngines"], &devEngines); err != nil {
		return fmt.Errorf("read devEngines: %w", err)
	}
	if devEngines == nil {
		return fmt.Errorf("devEngines is not an object")
	}
	if err := verifyEngine(devEngines["packageManager"], "pnpm", pnpmVersion); err != nil {
		return fmt.Errorf("verify devEngines.packageManager: %w", err)
	}
	if err := verifyEngine(devEngines["runtime"], "node", nodeVersion); err != nil {
		return fmt.Errorf("verify devEngines.runtime: %w", err)
	}
	return nil
}

func verifyEngine(raw json.RawMessage, name, version string) error {
	var engine map[string]string
	if err := json.Unmarshal(raw, &engine); err != nil {
		return fmt.Errorf("expected object: %w", err)
	}
	if engine["name"] != name || engine["version"] != "^"+version || engine["onFail"] != "download" {
		return fmt.Errorf("got name=%q version=%q onFail=%q", engine["name"], engine["version"], engine["onFail"])
	}
	return nil
}

func validateStableVersion(value string) error {
	if _, err := normalizeStableVersion(value); err != nil {
		return err
	}
	return nil
}

func normalizeStableVersion(value string) (string, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "v")
	version, err := semver.StrictNewVersion(trimmed)
	if err != nil || version.Prerelease() != "" || version.Metadata() != "" {
		return "", fmt.Errorf("version must be a stable semantic version such as 24.18.0")
	}
	return version.String(), nil
}
