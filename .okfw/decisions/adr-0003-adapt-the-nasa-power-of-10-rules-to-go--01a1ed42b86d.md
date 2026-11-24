---
x-okfw-id: "01a1ed42-b86d-7603-9bd7-269009e66544"
title: "ADR-0003: Adapt the NASA Power of 10 rules to Go"
type: "ADR"
tags:
  - "adr"
  - "power-of-10"
x-okfw-created: "2026-11-18T00:05:35-06:00"
x-okfw-timestamp: "2026-11-18T00:05:35-06:00"
description: "How lexis applies each Power of 10 rule to Go code, which rules the Go port of nasa-lint checks, and the deviation for R3."
---

# ADR-0003: Adapt the NASA Power of 10 rules to Go

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, `PLAN.md`. Gerard J. Holzmann, "The Power of 10" (2006). zmake ADR-0003 and ztree ADR-0003.

## Context
The user requires the Power of 10 rules for all Go code in lexis. The rules target C. Go has a garbage collector, closures and panics. `scripts/nasa-lint.py` checks Zig files only, so it finds no functions in lexis. Some rules need changes for Go.

## Decision
| Rule | lexis adaptation | Check |
|---|---|---|
| R1 Simple control flow, no recursion | No direct or indirect recursion. No `goto`. Explicit bounded stacks for the Markdown tree and the comment list. | Go port: call graph |
| R2 Fixed loop bounds | Each loop ranges over a slice or counts toward a limit, for example `MaxFiles` or `MaxASTNodes`. A loop that reaches its limit returns an error. | Go port bans `for {}` |
| R3 No allocation after initialization | Deviation: Go allocates on the heap. lexis sets a hard cap on each buffer and allocates each buffer once per file. | Review |
| R4 Short functions | A function has 60 lines or fewer. | Go port, golangci-lint `funlen` |
| R5 Assertion density | An average of two or more `assert.` calls per function, with `internal/assert`. Assertions have no side effects. | Go port |
| R6 Smallest scope | Declare each variable in the smallest block. | Compiler, review |
| R7 Check return values | Do not use `_ =`. Do not discard an error with `, _ :=` on a call. | Go port, golangci-lint `errcheck` |
| R8 Limited code generation | Keep build tags and generated code to a minimum. | Review |
| R9 Restricted indirection | Do not declare function types or function variables. Each func literal that cobra needs has an allow comment. | Go port |
| R10 Zero warnings | `go vet`, golangci-lint and the Go port report zero warnings. | `zmake lint` |

**golangci-lint:** version 2.15.0, from the release archive with its SHA-256 sum (`9b7fe6a6…cebcea`). `zmake lint` fails when the installed version differs. `.golangci.yml` enables the `standard` set (errcheck, govet, ineffassign, staticcheck, unused) and funlen. funlen counts the lines between the braces, so its limit of 58 equals the R4 limit of 60. The `_test.go` files are exempt from funlen.

**R7 and errcheck:** errcheck rejects an unchecked write, and R7 rejects `_ =`. `internal/output.Writer` keeps the first write error, and each command checks it once ([ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)).

**Assertions:** `assert.True` panics with an `assert.Failure` value. A failed assertion is a programmer error. The CLI recovers the panic at its boundary and exits with status 2. lexis reports bad input with an error, not with an assertion.

**Exempt files:** test files (`_test.go`) are exempt from R4 and R5.

**Suppression:** `// nasa-lint: allow(Rn) <reason>` on the line or the line above. Each suppression needs a reason.

**Port:** `cmd/nasa-lint` and `internal/nasalint` are the Go port of `scripts/nasa-lint.py`. `zmake lint` runs the port. The Python script stays in `scripts/` without changes. The port reads the syntax tree with `go/ast` and resolves names with `go/types`. It uses the standard library only.

- **Scope:** each package of the module, without `_test.go` files. It skips hidden directories, `testdata`, `vendor`, `node_modules` and each directory with its own `go.mod`.
- **R1:** the call graph covers one package. An import cycle is a compile error in Go, so a recursion across packages needs a function value, and R9 reports each function value. A call through an interface of the package reaches each method of that name. The search for cycles uses an explicit stack.
- **R4:** the count starts at the `func` keyword and ends at the last brace, as in the Python script.
- **R5:** a call into a package whose path ends in `/internal/assert` is an assertion. Calls in function literals count for the function that contains them. The functions of the assert package do not count.
- **R7:** the port reports `_ = x`, `var _ = x` and each blank on the left side of a call, for example `x, _ := f()` and `_, err := f()`. A conversion is not a call. `v, _ := m[k]`, `v, _ := x.(T)` and `var _ I = T{}` stay legal.
- **R9:** the port reports each function type outside a declaration, a function literal or an interface method. It also reports each function literal. A function of the package outside call position is a function value, for example `f := helper`.
- **Limits:** 4096 directories, 8192 files, 2^20 syntax nodes per file and 2^16 functions per package.
- **Limit of the port:** imports resolve to empty packages. So the port misses a function of another package as a value, for example `strings.ToUpper`.
- **Exit status:** 0 without findings, 1 with findings, 2 on a failure.

**semscope:** `.semscope.cue` applies this ADR to the `nasa-p10` ruleset of semscope. `zmake lint` runs semscope when it is on the PATH ([ADR-0011](/decisions/adr-0011-semscope-checks-the-power-of-10-rules-in-the-lint-task--01a1f5140d5e.md)). It turns off R3 and R6 and excludes `scripts/`. Run `semscope scan -r nasa-p10 -r .semscope.cue .`. On 2026-11-19, semscope and the Go port gave 0 findings, 174 functions and 352 assertions. Without the overlay, semscope gives an R3 finding for each `append` and `make`. It also gives R6 findings for package variables and an R5 finding for the Python scripts.

`assert.Recover` is the boundary of each command. Defer it directly, so it can call `recover`. It recovers an `assert.Failure` only. Each other panic keeps its stack trace.

## Consequences
- The code is more verbose than common Go: explicit stacks, limits and assertions.
- A sort with `slices.SortFunc`, a `go/types` error callback and a cobra command need a function value. Each use has an allow comment with a reason.
- An input that exceeds a limit fails with an error.

## Alternatives considered
- **Run `nasa-lint.py` without a port.** Rejected: it reads Zig only and reports 0 functions on Go code.
- **Recursion with a depth counter.** Rejected: R1 forbids recursion.
