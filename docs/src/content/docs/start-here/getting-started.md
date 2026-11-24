---
title: Getting Started
description: Install a lexiscope release, or build lexiscope from source and run its gates.
sidebar:
  order: 1
---

This page tells how to install a release of lexiscope. It also tells how to build the repository and run its checks.

## Install a release

The [v0.1.0 release](https://github.com/infrashift/lexiscope/releases/tag/v0.1.0) has static Linux binaries:

| File | Platform |
|---|---|
| `lexiscope-linux-amd64` | Linux on x86-64 |
| `lexiscope-linux-arm64` | Linux on ARM64 |
| `lexiscope-SHA256SUMS` | The SHA-256 sums of the binaries |

The release has no test of the arm64 binary on an arm64 host. Build lexiscope from source for other platforms.

1. Download the binary for the platform and `lexiscope-SHA256SUMS`.
2. Run `sha256sum -c --ignore-missing lexiscope-SHA256SUMS` to check the download.
3. Copy the binary into a directory on the `PATH` with the name `lexiscope`.
4. Make the link `lexis` to use the short alias.
5. Run `lexiscope --version`.

These commands install the amd64 binary into `~/.local/bin`:

```sh
curl -LO https://github.com/infrashift/lexiscope/releases/download/v0.1.0/lexiscope-linux-amd64
curl -LO https://github.com/infrashift/lexiscope/releases/download/v0.1.0/lexiscope-SHA256SUMS
sha256sum -c --ignore-missing lexiscope-SHA256SUMS
install -m 755 lexiscope-linux-amd64 ~/.local/bin/lexiscope
ln -s lexiscope ~/.local/bin/lexis
lexiscope --version
```

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
