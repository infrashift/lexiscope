---
id: "01a1ed42-b88b-7eca-a21d-cb5916646909"
title: "ISSUE-002: CLI, version and ruleset overrides, and the Go port of nasa-lint"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-2"
  - "cli"
  - "closed"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T03:55:25-06:00"
description: "Add the cobra CLI, the version package, golangci-lint and a Go port of nasa-lint.py."
---

# ISSUE-002: CLI, version and ruleset overrides, and the Go port of nasa-lint

**Status:** closed · **Phase:** 2

## Description
Add the commands of `PLAN.md` Phase 2 and make the Power of 10 check read Go code.

## Acceptance Criteria
- [x] A Go port of `scripts/nasa-lint.py` checks R1, R2, R4, R5, R7 and R9, and `zmake lint` runs it.
- [x] golangci-lint v2 runs at a pinned version with errcheck, funlen and govet.
- [x] `lexis --version`, `lexis -v` and `lexis version` print the version line.
- [x] `lexis scan` accepts the flags of `PLAN.md`.
- [x] `lexis rules list` and `lexis rules inspect` show the resolved ruleset.
- [x] The docs site has a CLI guide and a CLI reference.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
- `cmd/nasa-lint` and `internal/nasalint` port the Python script to Go with `go/ast` and `go/types` (commit 04c24b7). [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) records the rules of the port. Eight mutants of the port each fail its tests.
- golangci-lint 2.15.0 runs in `zmake lint`. The target checks the installed version.
- `internal/version` reads the `-ldflags` values, then the build information of the binary. `zmake build` sets the values from git.
- `internal/cli` has `version`, `scan`, `rules list` and `rules inspect`. `scan` and `rules` check their flags and then name the phase that delivers the work ([ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)).
- `internal/output.Writer` checks each write without `_ =`. errcheck and R7 both pass.
- The docs site has the CLI guide and the CLI reference. A test checks that the reference lists each command and each flag.
- nasa-lint: 68 functions, 138 assertions (average 2.03).

## Related
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) and [ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md).
