---
id: "01a1ed42-b86d-7603-9bd7-269009e66544"
title: "ADR-0003: Adapt the NASA Power of 10 rules to Go"
type: "ADR"
tags:
  - "adr"
  - "power-of-10"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "How lexis applies each Power of 10 rule to Go code, which rules the Go port of nasa-lint checks, and the deviation for R3."
---

# ADR-0003: Adapt the NASA Power of 10 rules to Go

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, `PLAN.md`. Gerard J. Holzmann, "The Power of 10: Rules for Developing Safety-Critical Code" (2006). zmake ADR-0003 and ztree ADR-0003.

## Context
The user requires the Power of 10 rules for all Go code in lexis. The rules target C. Go has a garbage collector, closures and panics. `scripts/nasa-lint.py` checks Zig files only, so it finds no functions in lexis. Some rules need an adaptation for Go.

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

**Assertions:** `assert.True` panics with an `assert.Failure` value. A failed assertion is a programmer error. The CLI recovers the panic at its boundary and exits with status 2. lexis reports bad input with an error, not with an assertion.

**Exempt files:** test files (`_test.go`) are exempt from R4 and R5.

**Suppression:** `// nasa-lint: allow(Rn) <reason>` on the line or the line above. Each suppression needs a reason.

**Port:** Phase 2 ports `scripts/nasa-lint.py` to Go in a separate commit. Until then `zmake lint` runs the Python script, and it checks nothing in lexis. `gofmt` and `go vet` run from Phase 1.

## Consequences
- The code is more verbose than common Go: explicit stacks, limits and assertions.
- An input that exceeds a limit fails with an error.

## Alternatives considered
- **Run `nasa-lint.py` without a port.** Rejected: it reads Zig only and reports 0 functions on Go code.
- **Recursion with a depth counter.** Rejected: R1 forbids recursion.
