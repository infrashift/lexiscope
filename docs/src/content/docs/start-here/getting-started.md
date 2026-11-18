---
title: Getting Started
description: Build lexis from source and run its gates.
sidebar:
  order: 1
---

lexis has no binary release and no commands yet. Phase 2 of the plan adds the CLI. This page tells how to build the repository and run its checks.

## Requirements

| Tool | Version | Use |
|---|---|---|
| Go | 1.26 or later | Build and test lexis |
| Python | 3.10 or later | Run `scripts/ste-check.py` and `scripts/nasa-lint.py` |
| zmake or GNU Make | any | Run the targets of the `ZMakefile` or the `Makefile` |
| okfw | 2026.10 or later | Validate the wiki in `.okfw/` |
| bun | 1.3.9 | Build this site |

From Phase 4, lexis reads source comments with [ztree](https://github.com/infrashift/ztree), a Tree-sitter runtime in Zig. You then also need Zig 0.17.1 and a ztree checkout next to the lexis checkout.

## Build and check

1. Clone the repository.
2. Run `zmake build` to write `bin/lexis`.
3. Run `zmake check` to run every gate.

GNU Make reads the `Makefile`, which has the same targets. Run `make check` if zmake is not on the `PATH`.

## Run this site locally

1. Run `zmake docs-dev`.
2. Open the address that the server prints, for example `http://localhost:4321/lexis/`.

Astro uses the next free port when port 4321 is in use.

The target installs the packages from `docs/bun.lock` and does not change the lock file.
