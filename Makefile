# lexis development gates. Plain GNU Make syntax. The ZMakefile has the same targets.

GO ?= go
GOFMT ?= $(shell $(GO) env GOROOT)/bin/gofmt
PYTHON ?= python3
BUN ?= bun
ZIG ?= zig
# The ztree checkout. go.mod points its replace directive here (ADR-0002).
ZTREE_DIR ?= ../ztree
GOLANGCI_LINT ?= golangci-lint
# The pinned golangci-lint version (ADR-0003). The lint target checks it.
GOLANGCI_LINT_VERSION := 2.15.0
# Version metadata for -ldflags (internal/version). The date is the commit
# time, so two builds of one commit give the same binary.
VERSION ?= $(shell git describe --tags --match 'v[0-9]*' --dirty 2>/dev/null || echo v0.0.0-dev)
COMMIT ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)$(shell git diff --quiet HEAD 2>/dev/null || echo -dirty)
BUILD_DATE ?= $(shell TZ=UTC git log -1 --format=%cd --date=format-local:%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)
VERSION_PKG := github.com/infrashift/lexis/internal/version
LDFLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)
# gofmt walks directories, so it gets the Go directories only, not docs/.
GO_DIRS := cmd internal rules
# The self-check reads the files of scripts/ste-check.py and the Go comments.
SELF_CHECK_PATHS := PLAN.md CLAUDE.md docs/src/content/docs .okfw cmd internal rules
SELF_CHECK_EXCLUDES := --exclude .okfw/daily --exclude .okfw/ephemeral --exclude index.md --exclude log.md

.PHONY: check ztree build test lint self-check fmt wiki docs-style docs-dev docs-build

## check: run every gate (the default goal)
check: lint test wiki docs-style self-check

## ztree: build libztree.a in the ztree checkout (cgo links it)
ztree:
	cd $(ZTREE_DIR) && $(ZIG) build lib --release

## build: compile the lexis binary into bin/
build: ztree
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/lexis ./cmd/lexis

## test: run the unit tests with the race detector
test: ztree
	$(GO) test -race -count=1 ./...

## lint: check the formatting, run go vet and golangci-lint, and check the Power of 10 rules
lint: ztree
	$(GOFMT) -l $(GO_DIRS) | (! grep .)
	$(GO) vet ./...
	test "$$($(GOLANGCI_LINT) version --short)" = "$(GOLANGCI_LINT_VERSION)"
	$(GOLANGCI_LINT) run ./...
	$(GO) run ./cmd/nasa-lint

## self-check: scan the documentation and the comments of lexis with lexis (an error fails)
self-check: build
	bin/lexis scan $(SELF_CHECK_EXCLUDES) $(SELF_CHECK_PATHS)

## fmt: format the Go sources in place
fmt:
	$(GOFMT) -w $(GO_DIRS)

## wiki: validate the OKF wiki
wiki:
	okfw validate --all
	okfw validate --okf

## docs-style: check the documentation against ASD-STE100 rules
docs-style:
	$(PYTHON) -I scripts/ste-check.py

## docs-dev: run the Astro Starlight site locally with bun (default http://localhost:4321/lexis/)
docs-dev:
	cd docs && $(BUN) install --frozen-lockfile && $(BUN) --bun run dev

## docs-build: build the site into docs/dist (the same build as the Pages deploy)
docs-build:
	cd docs && $(BUN) install --frozen-lockfile && $(BUN) --bun run build
