---
title: Make Targets
description: The targets of the lexiscope ZMakefile and Makefile.
sidebar:
  order: 2
---

The `ZMakefile` and the `Makefile` have the same targets. zmake reads the `ZMakefile`. GNU Make reads the `Makefile`.

| Target | Command | Part of `check` |
|---|---|---|
| `check` | Runs `lint`, `test`, `wiki`, `docs-style` and `self-check` | default goal |
| `ztree` | Builds `libztree.a` in `../ztree`, or in `ZTREE_DIR` | through `build`, `test` and `lint` |
| `build` | `go build` with the version from git, into `bin/lexiscope`. It also makes the link `bin/lexis`, the short alias. | no |
| `install` | Runs `build`, then copies `bin/lexiscope` into `$(PREFIX)/bin` and makes the link `lexis` there. `PREFIX` is `~/.local` by default. | no |
| `test` | `go test -race -count=1 ./...` | yes |
| `lint` | `gofmt -l`, `go vet ./...`, golangci-lint 2.15.0 and `go run ./cmd/nasa-lint` | yes |
| `fmt` | `gofmt -w` on the Go directories | no |
| `wiki` | `okfw validate --all` and `okfw validate --okf` | yes |
| `docs-style` | `scripts/ste-check.py` | yes |
| `self-check` | `bin/lexiscope scan` on the documents and the Go comments of lexiscope, with `vocabulary.cue`. It also writes `lexiscope-self.sarif` and `lexiscope-self-v22.sarif`. Each finding fails it. | yes |
| `docs-dev` | Runs this site locally, by default at `http://localhost:4321/lexiscope/` | no |
| `docs-build` | Builds this site into `docs/dist/` | no |

The `check` target does not build this site, so the Go gates do not need bun.
The `docs-dev` and `docs-build` targets install packages with `--frozen-lockfile`. To change a package version, edit `docs/package.json` and run `bun install` in `docs/`.

`scripts/nasa-lint.py` reads Zig files only. `cmd/nasa-lint` is its Go port.
