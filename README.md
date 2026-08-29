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
