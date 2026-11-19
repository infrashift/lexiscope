---
title: Getting Started
description: Build lexiscope from source and run its gates.
sidebar:
  order: 1
---

lexiscope has no binary release. This page tells how to build the repository and run its checks.

## Requirements

| Tool | Version | Use |
|---|---|---|
| Go | 1.26.10 or later | Build and test lexiscope |
| Python | 3.10 or later | Run `scripts/ste-check.py` and `scripts/nasa-lint.py` |
| zmake or GNU Make | any | Run the targets of the `ZMakefile` or the `Makefile` |
| golangci-lint | 2.15.0 exactly | Run the static checks of `zmake lint` |
| okfw | 2026.10 or later | Validate the wiki in `.okfw/` |
| bun | 1.3.9 | Build this site |

lexiscope reads source comments with [ztree](https://github.com/infrashift/ztree), a Tree-sitter runtime in Zig. You also need Zig 0.17.1, a C compiler, and a ztree checkout next to the lexiscope checkout. `zmake build` builds the ztree library first.

## Build and check

1. Clone the repository.
2. Run `zmake build` to write `bin/lexiscope`.
3. Run `bin/lexiscope --version`.
4. Run `zmake check` to run every gate.
5. Run `zmake install` to copy `bin/lexiscope` into `~/.local/bin`. The target also makes the link `lexis`, so both names start the CLI. Set `PREFIX` for a different directory, for example `zmake install PREFIX=/usr/local`.

GNU Make reads the `Makefile`, which has the same targets. Run `make check` if zmake is not on the `PATH`.

## Run this site locally

1. Run `zmake docs-dev`.
2. Open the address that the server prints, for example `http://localhost:4321/lexiscope/`.

Astro uses the next free port when port 4321 is in use.

The target installs the packages from `docs/bun.lock` and does not change the lock file.
