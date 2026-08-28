# Devkit Development Guide

## Project Scope

Devkit is a CLI application written in Go. It is designed to:

- Initialize projects and generate project scaffolding.
- Provide and apply reusable code templates.
- Download, install, or prepare device-related dependencies.
- Create, update, and configure files required by projects or local development environments.

All implementations should be designed for CLI use. Commands must be clear, behavior must be predictable, errors must be actionable, and cross-platform compatibility should be considered.

## Go Development Guidelines

- This is a Go project. New code must follow official Go conventions and be formatted with `gofmt`.
- Prefer the Go standard library. Evaluate third-party packages only when the standard library cannot reasonably satisfy the requirement.
- Avoid reinventing existing solutions. When a mature implementation already exists for a general-purpose feature, use a reliable package instead of building a custom replacement.
- Do not introduce a large package, a complex dependency tree, or a high maintenance burden for a small feature. The benefit of a dependency must clearly outweigh its cost.
- Keep implementations simple, maintain clear module boundaries, and provide useful error messages for user-visible failures.

## Third-Party Package Selection

Conduct thorough research before introducing any third-party Go package. Do not select a package based on assumptions or a single metric.

Research must consider at least:

- GitHub stars, usage, import counts, and general community adoption.
- Recent commits and releases, issue and pull request responsiveness, and maintainer activity.
- Compatibility with the Go version used by this project and its primary target platforms.
- API stability, documentation quality, license, dependency tree size, and known security risks.
- Newer, lighter, or more actively maintained alternatives that provide similar functionality.

Selection principles:

- Prefer widely adopted packages with strong usage, high GitHub stars, good documentation, and active maintenance.
- High usage and star counts are not sufficient on their own. Do not select a package solely because of its historical popularity if it is unmaintained, outdated, or superseded by a modern alternative.
- When candidates have similar maturity, prefer the package with more recent development, regular releases, and compatibility with modern Go tooling.
- Briefly document the evaluated candidates, selection rationale, important tradeoffs, and links to the official documentation or source repository.

## Dependency Installation Approval

The user must run every `go get` command manually. Agents and automated tools must never run `go get` themselves.

When a task requires a new third-party dependency:

1. Complete the requirement analysis, package research, and package selection first.
2. Stop immediately when the dependency needs to be installed. Do not run `go get`, and do not bypass this process by manually editing `go.mod` or `go.sum`.
3. Tell the user the exact command to run, such as `go get example.com/module@version`.
4. Provide links to the selected package's official documentation and source repository, together with a short explanation of the selection rationale.
5. Wait for the user to confirm that the command has been completed.
6. Continue implementing code that depends on the package and perform subsequent verification only after receiving confirmation.

If a task does not require a new dependency, implementation may continue without triggering this approval process.
