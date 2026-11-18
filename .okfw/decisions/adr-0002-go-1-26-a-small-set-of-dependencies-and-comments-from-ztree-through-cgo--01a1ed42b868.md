---
id: "01a1ed42-b868-7bb6-bc24-a408254ad25d"
title: "ADR-0002: Go 1.26, a small set of dependencies, and comments from ztree through cgo"
type: "ADR"
tags:
  - "adr"
  - "toolchain"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "lexis needs Go 1.26 or later, uses cobra, CUE and go-sarif as its only Go dependencies, and reads source comments with ztree through cgo."
---

# ADR-0002: Go 1.26, a small set of dependencies, and comments from ztree through cgo

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, `PLAN.md`. The ztree C ABI spec in the ztree wiki.

## Context
lexis needs a CLI, a CUE evaluator, a SARIF writer and a parser for the comments of five languages. `github.com/owenrumney/go-sarif/v3` needs Go 1.26. The user wants a Tree-sitter runtime that follows the Power of 10 rules, not lexers by hand and not the C runtime of Tree-sitter.

## Decision
- The module `github.com/infrashift/lexis` sets `go 1.26.10`, because go-sarif v3.3.2 needs it. The host has Go 1.27.3.
- The Go dependencies are `github.com/spf13/cobra`, `cuelang.org/go` and `github.com/owenrumney/go-sarif/v3`. Each phase adds a dependency when it needs it. Phase 2 added cobra v1.10.2, with pflag v1.0.9 and mousetrap v1.1.0. Phase 3 added `cuelang.org/go` v0.17.1, the newest stable release. It brings 10 indirect modules, and the `lexis` binary grows from about 4 MB to 25.8 MB. Phase 5 added go-sarif v3.3.2 ([ADR-0006](/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md)). It brings gojsonschema v1.2.0 and its two JSON pointer modules. Write an ADR before you add another one.
- The comments come from ztree, a Tree-sitter runtime in Zig. lexis imports `github.com/infrashift/ztree/bindings/go`. A `replace` directive in `go.mod` points to the ztree checkout next to lexis.
- Build ztree first: run `zig build lib --release` in the ztree root. The binding links `zig-out/lib/libztree.a`.
- cgo works with the default C compiler. With `CC="zig cc"`, add `-Wno-gcc-install-dir-libstdcxx` on a host with two GCC installations.

## Consequences
- `go test -race ./...` needs cgo and a C compiler.
- A build of lexis needs a ztree checkout and Zig 0.17.1. The `ztree` target of the ZMakefile builds `libztree.a` in `$(ZTREE_DIR)` (default `../ztree`). `build`, `test` and `lint` depend on it.
- The C library of ztree takes no callbacks and keeps no global state, so goroutines can call it at the same time.

## Alternatives considered
- **Lexers by hand for each language.** Rejected: five lexers to maintain, and string and raw-string rules differ in each language.
- **The C runtime of Tree-sitter through cgo.** Rejected by the user: it does not follow the Power of 10 rules.
