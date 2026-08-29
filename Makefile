APP_NAME := devkit
VERSION ?= $(shell git describe --tags --match 'v*' --abbrev=0 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: help fmt lint test test-race vet mod-verify check prek install-hooks build release-patch release-minor release-major

help:
	@printf '%-28s%s\n' \
		'make fmt' 'Check gofmt formatting' \
		'make lint' 'Format and lint Go files' \
		'make test' 'Run Go tests' \
		'make test-race' 'Run Go tests with the race detector' \
		'make vet' 'Run go vet' \
		'make mod-verify' 'Verify module checksums' \
		'make check' 'Run formatting, tests, vet, and module checks' \
		'make prek' 'Run prek hooks against all files' \
		'make install-hooks' 'Install prek Git hooks' \
		'make build' 'Build the CLI' \
		'make release-patch' 'Create and push the next patch release' \
		'make release-minor' 'Create and push the next minor release' \
		'make release-major' 'Create and push the next major release'

fmt:
	@test -z "$$(find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -l {} +)"

lint:
	@command -v golangci-lint >/dev/null 2>&1 || (printf '%s\n' 'golangci-lint is required for make lint' >&2; exit 1)
	golangci-lint fmt
	git diff --exit-code -- '*.go'
	golangci-lint run

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

mod-verify:
	go mod verify

check: fmt lint test test-race vet mod-verify

prek:
	prek run --all-files

install-hooks:
	prek install

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$(APP_NAME)" ./cmd/devkit

define release
	@test -z "$$(git status --porcelain)" || (printf '%s\n' 'Working tree is not clean; commit changes before releasing' >&2; exit 1)
	@version="$$(go tool svu $(1))" && git tag -a "$$version" -m "Release $$version" && git push origin "$$version"
endef

release-patch:
	$(call release,patch)

release-minor:
	$(call release,minor)

release-major:
	$(call release,major)
