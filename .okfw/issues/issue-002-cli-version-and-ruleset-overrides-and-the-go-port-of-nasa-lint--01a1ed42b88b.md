---
id: "01a1ed42-b88b-7eca-a21d-cb5916646909"
title: "ISSUE-002: CLI, version and ruleset overrides, and the Go port of nasa-lint"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-2"
  - "cli"
  - "in-progress"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T03:42:05-06:00"
description: "Add the cobra CLI, the version package, golangci-lint and a Go port of nasa-lint.py."
---

# ISSUE-002: CLI, version and ruleset overrides, and the Go port of nasa-lint

**Status:** in-progress · **Phase:** 2

## Description
Add the commands of `PLAN.md` Phase 2 and make the Power of 10 check read Go code.

## Acceptance Criteria
- [ ] A Go port of `scripts/nasa-lint.py` checks R1, R2, R4, R5, R7 and R9, and `zmake lint` runs it.
- [ ] golangci-lint v2 runs at a pinned version with errcheck, funlen and govet.
- [ ] `lexis --version`, `lexis -v` and `lexis version` print the version line.
- [ ] `lexis scan` accepts the flags of `PLAN.md`.
- [ ] `lexis rules list` and `lexis rules inspect` show the resolved ruleset.
- [ ] The docs site has a CLI guide and a CLI reference.
- [ ] `zmake check` reports zero errors and zero warnings.

## Related
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md).
