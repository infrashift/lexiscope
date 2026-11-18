---
title: Make Targets
description: The targets of the lexis ZMakefile and Makefile.
sidebar:
  order: 2
---

The `ZMakefile` and the `Makefile` have the same targets. zmake reads the `ZMakefile`. GNU Make reads the `Makefile`.

| Target | Command | Part of `check` |
|---|---|---|
| `check` | Runs `lint`, `test`, `wiki` and `docs-style` | default goal |
| `build` | `go build -o bin/lexis ./cmd/lexis` | no |
| `test` | `go test -race -count=1 ./...` | yes |
| `lint` | `gofmt -l`, `go vet ./...` and `scripts/nasa-lint.py` | yes |
| `fmt` | `gofmt -w` on the Go directories | no |
| `wiki` | `okfw validate --all` and `okfw validate --okf` | yes |
| `docs-style` | `scripts/ste-check.py` | yes |
| `docs-dev` | Runs this site locally, by default at `http://localhost:4321/lexis/` | no |
| `docs-build` | Builds this site into `docs/dist/` | no |

The `check` target does not build this site, so the Go gates do not need bun.
The `docs-dev` and `docs-build` targets install packages with `--frozen-lockfile`. To change a package version, edit `docs/package.json` and run `bun install` in `docs/`.

Until Phase 2, `scripts/nasa-lint.py` reads Zig files only, so it finds no functions in lexis. The Go port in Phase 2 replaces it.
