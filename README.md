# devkit

Devkit is a Go-based CLI for automating project setup and local development environment preparation.

It provides a consistent entry point for tasks that would otherwise require repeatedly copying templates, downloading tools, or editing configuration files by hand.

## Installation

Install the latest release on macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/tinywaves/devkit/main/scripts/install.sh | bash
```

The installer detects the operating system and CPU architecture, downloads the matching GoReleaser archive, verifies it with `checksums.txt`, and installs `devkit` to `~/.local/bin/devkit` by default.

Install a specific version by setting `DEVKIT_VERSION`:

```sh
curl -fsSL https://raw.githubusercontent.com/tinywaves/devkit/main/scripts/install.sh | DEVKIT_VERSION=v0.0.0 bash
```

Override the installation directory with `DEVKIT_PREFIX`:

```sh
curl -fsSL https://raw.githubusercontent.com/tinywaves/devkit/main/scripts/install.sh | DEVKIT_PREFIX="$HOME/.local" bash
```

## What Devkit Does

- Initialize projects and generate project scaffolding.
- Create files from reusable code templates.
- Download tools and device-related dependencies.
- Generate or update project configuration files.
- Prepare consistent local development environments.

## Goals

- **Repeatable:** Produce consistent results across projects and machines.
- **Practical:** Automate common setup work without hiding important behavior.
- **Extensible:** Make it straightforward to add new initializers, templates, downloads, and configuration tasks.
- **Cross-platform:** Keep commands and generated output portable whenever possible.
- **Safe:** Provide clear output and actionable errors, especially when changing files or installing tools.

## Technology

Devkit is implemented in Go and targets a simple, portable command-line experience.

## Project Structure

```text
cmd/
├── devkit/                 # Devkit executable entry point
└── update-runtime/         # Runtime version snapshot updater
internal/
├── app/                    # Root command assembly
├── commands/
│   └── initcmd/            # devkit init command and initializers
└── runtime/
    ├── config.go           # Embedded version data loader
    └── data/versions.json  # Node.js and pnpm version snapshot
```

Executable entry points under `cmd` stay small. User-facing commands live in dedicated packages under `internal/commands`, while reusable domain logic and embedded data live in focused packages under `internal`.

## Development

The repository uses the conventional root-level `Makefile` for local development tasks. The lowercase name `makefile` is also recognized by GNU Make, but this project keeps only `Makefile` to follow common Go project conventions and avoid case-sensitive filesystem ambiguity.

Git hooks are managed by [`prek`](https://prek.j178.dev/) through the native `prek.toml` configuration. Prek only provides the Git hook integration; the project checks themselves live in `make check`. Install `prek` using the official instructions, then install the repository hook:

```sh
prek install
```

The pre-commit hook runs `make check`, which includes formatting, `golangci-lint`, tests, race tests, `go vet`, and module verification.

The `lint` and `check` targets require `golangci-lint` to be available on `PATH`. CI installs the pinned version from `.golangci-lint-version` before running `make check`.

Run the CLI directly:

```sh
go run ./cmd/devkit
go run ./cmd/devkit --help
go run ./cmd/devkit --version
```

Initialize project tooling interactively:

```sh
devkit init
```

Or select an initializer directly:

```sh
devkit init js
devkit init eslint
```

The `js` initializer configures the JavaScript/TypeScript toolchain for any project containing `package.json`, including frontend and Node.js projects. It uses the configured Node.js and pnpm version snapshots, then writes the selected versions to `packageManager` and `devEngines` without requiring network access.

### Runtime Version Snapshots

The version data is checked into `internal/runtime/data/versions.json` and embedded into the Devkit binary. The current snapshot contains the highest stable release for each Node.js major version from 18 onward and each pnpm major version from 10 onward. Node.js LTS labels are maintained alongside the Node.js versions.

The snapshot is based on these upstream sources:

- Node.js release index: `https://nodejs.org/dist/index.json`
- Node.js release schedule: `https://raw.githubusercontent.com/nodejs/Release/main/schedule.json`
- pnpm package metadata: `https://registry.npmjs.org/pnpm`

The two Node.js requests are intentional: the release index provides available versions, while the release schedule identifies LTS release lines. The pnpm request provides all published versions, from which the highest stable version in each major line is selected.

Normal `devkit init` does not query these endpoints. Because a static snapshot cannot detect upstream changes while offline, refreshing it belongs to repository maintenance rather than project initialization. The Go updater downloads the complete JSON responses, extracts the required fields, and rewrites the normalized `versions.json`. If pnpm 12 or any other release line receives a newer stable version, the rewritten JSON produces a Git diff; no upstream changes produce no diff. Releasing a new Devkit binary then distributes the refreshed choices.

Refresh the snapshot locally with:

```sh
make update-runtime
```

This command is intended for repository maintenance. It updates the checked-in data snapshot; regular users do not need to run it and normal initialization remains offline.

Build a local binary:

```sh
go build -o devkit ./cmd/devkit
```

Release builds should inject the release tag into the binary:

```sh
go build -ldflags "-X main.version=v1.0.0" -o devkit ./cmd/devkit
./devkit --version
```

The development fallback remains `dev` when no version is injected.

## CI/CD

GitHub Actions runs the same `make check` quality gate used by the local pre-commit hook, along with `govulncheck`, GoReleaser configuration validation, and a build for pushes to `main` and pull requests targeting `main`.

The Makefile provides the same local commands:

```sh
make check
make lint
make prek
make install-hooks
make update-runtime
make test
make vet
make build
```

`svu` is tracked as a Go tool dependency, so release targets run it through `go tool` and do not require a global installation. These targets calculate the next version, create an annotated tag, and push it to `origin`:

```sh
make release-patch
make release-minor
make release-major
```

Pushing a version tag builds archives for Linux, macOS, and Windows, then publishes a GitHub Release through GoReleaser:

The release workflow uses the open-source GoReleaser distribution. It verifies Go modules before running GoReleaser and does not require a GoReleaser Pro license.

The release contains:

```text
devkit_linux_x86_64.tar.gz
devkit_linux_arm64.tar.gz
devkit_macOS_x86_64.tar.gz
devkit_macOS_arm64.tar.gz
devkit_windows_x86_64.zip
devkit_windows_arm64.zip
checksums.txt
```
